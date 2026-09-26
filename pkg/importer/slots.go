package importer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

const (
	fallbackAsset   = "Assets:FIXME"
	fallbackExpense = "Expenses:FIXME"
	fallbackIncome  = "Income:FIXME"
)

// rowToSlotOrder fills Beancount slots from the template mapping, then lets
// personal rules assign accounts and edit metadata. Template rules are not
// applied: account names are not part of the template contract.
func rowToSlotOrder(profile *Profile, row Row) (ir.Order, bool, error) {
	mapped, err := applySlotMapping(profile, row)
	if err != nil {
		return ir.Order{}, false, err
	}
	order := mapped.order
	row = mapped.row

	ignore := false
	merged := Actions{Amount: mapped.absolute, Currency: order.Currency}
	for _, rule := range enabledRules(profile.PersonalRules) {
		if !ruleInScope(rule, profile.ID) {
			continue
		}
		matches, err := ruleMatches(rule, row, order)
		if err != nil {
			return ir.Order{}, false, err
		}
		if !matches {
			continue
		}
		if err := applySlotPersonalActions(&order, &row, rule, &ignore); err != nil {
			return ir.Order{}, false, err
		}
		if err := mergeV2Actions(&merged, rule.Actions); err != nil {
			return ir.Order{}, false, err
		}
		if account := strings.TrimSpace(rule.Actions.From.Account); account != "" {
			putFieldSource(&order, ruleFieldSource(rule, "from", rule.Actions.From.Account))
		}
		if account := strings.TrimSpace(rule.Actions.To.Account); account != "" {
			putFieldSource(&order, ruleFieldSource(rule, "to", rule.Actions.To.Account))
		}
		if strings.TrimSpace(rule.Actions.Amount) != "" {
			putFieldSource(&order, ruleFieldSource(rule, "amount", rule.Actions.Amount))
		}
		if strings.TrimSpace(rule.Actions.Currency) != "" {
			putFieldSource(&order, ruleFieldSource(rule, "currency", rule.Actions.Currency))
		}
	}
	if ignore {
		return order, true, nil
	}
	if order.PayTime.IsZero() {
		return ir.Order{}, false, fmt.Errorf("slots.date did not set date")
	}
	if strings.TrimSpace(merged.Amount) == "" {
		merged.Amount = mapped.absolute
	}
	if strings.TrimSpace(merged.Currency) == "" {
		merged.Currency = order.Currency
	}
	fromMissing := strings.TrimSpace(merged.From.Account) == ""
	toMissing := strings.TrimSpace(merged.To.Account) == ""
	fillMissingSides(&merged, mapped.expense)
	if fromMissing {
		putFieldSource(&order, ir.FieldSource{Slot: "from", Origin: ir.FieldOriginEngine})
	}
	if toMissing {
		putFieldSource(&order, ir.FieldSource{Slot: "to", Origin: ir.FieldOriginEngine})
	}
	if err := renderV2Postings(&order, row, merged); err != nil {
		return ir.Order{}, false, err
	}
	order.MinusAccount = strings.TrimSpace(merged.From.Account)
	order.PlusAccount = strings.TrimSpace(merged.To.Account)
	return order, false, nil
}

type slotMapped struct {
	row      Row
	order    ir.Order
	absolute string
	expense  bool
}

func applySlotMapping(profile *Profile, row Row) (slotMapped, error) {
	slots := profile.Template.Slots
	order := ir.Order{
		OrderType: ir.OrderTypeNormal,
		Currency:  profile.Template.DefaultCurrency,
		Metadata:  map[string]string{},
	}
	if row.Metadata == nil {
		row.Metadata = map[string]string{}
	}

	keys := append([]string{}, slots.Metadata.Keys...)
	for _, key := range keys {
		value := strings.TrimSpace(renderRuleText(slots.Metadata.Values[key], row, order))
		if value == "" {
			continue
		}
		order.Metadata[key] = value
		row.Metadata[key] = value
		putFieldSource(&order, templateFieldSource(metadataSlot(key), slots.Metadata.Values[key]))
	}
	order.MetadataKeys = metadataKeyOrder(keys, order.Metadata)

	if slots.Payee != "" {
		order.Peer = strings.TrimSpace(renderRuleText(slots.Payee, row, order))
		row.Payee = order.Peer
		putFieldSource(&order, templateFieldSource("payee", slots.Payee))
	}
	if slots.Narration != "" {
		order.Item = strings.TrimSpace(renderRuleText(slots.Narration, row, order))
		row.Narration = order.Item
		putFieldSource(&order, templateFieldSource("narration", slots.Narration))
	}
	if slots.Flag != "" {
		order.Flag = strings.TrimSpace(renderRuleText(slots.Flag, row, order))
		if order.Flag != "" {
			putFieldSource(&order, templateFieldSource("flag", slots.Flag))
		}
	}
	if slots.Tags != "" {
		order.Tags = splitList(renderRuleText(slots.Tags, row, order))
		if len(order.Tags) > 0 {
			putFieldSource(&order, templateFieldSource("tags", slots.Tags))
		}
	}
	if slots.Links != "" {
		order.Links = splitList(renderRuleText(slots.Links, row, order))
		if len(order.Links) > 0 {
			putFieldSource(&order, templateFieldSource("links", slots.Links))
		}
	}
	if slots.Currency != "" {
		order.Currency = strings.TrimSpace(renderRuleText(slots.Currency, row, order))
		row.Currency = order.Currency
		if order.Currency != "" {
			putFieldSource(&order, templateFieldSource("currency", slots.Currency))
		}
	} else if strings.TrimSpace(order.Currency) != "" {
		putFieldSource(&order, ir.FieldSource{Slot: "currency", Origin: ir.FieldOriginTemplate})
	}
	if slots.Date != "" {
		rendered := strings.TrimSpace(renderRuleText(slots.Date, row, order))
		row.Date = rendered
		payTime, err := parseDate(rendered, profile.Template.DateFormat)
		if err != nil {
			return slotMapped{}, fmt.Errorf("slots.date %q: %w", slots.Date, err)
		}
		order.PayTime = payTime
		putFieldSource(&order, templateFieldSource("date", slots.Date))
	}

	renderedAmount, err := renderPostingTextStrict(slots.Amount, row, order)
	if err != nil {
		return slotMapped{}, fmt.Errorf("slots.amount %q: %w", slots.Amount, err)
	}
	amount, err := parseAmountExact(renderedAmount, profile.Template.AmountPrefix)
	if err != nil {
		return slotMapped{}, fmt.Errorf("slots.amount %q => %q: %w", slots.Amount, renderedAmount, err)
	}
	expense := metadataNegates(order.Metadata, profile.Template.AmountSign)
	if expense && amount.Sign() > 0 {
		amount = amount.Neg()
	}
	if !expense && amount.Sign() < 0 {
		expense = true
	}
	if amount.Sign() < 0 {
		order.Type = ir.TypeSend
	} else {
		order.Type = ir.TypeRecv
	}
	order.TypeOriginal = order.Metadata[profile.Template.AmountSign.Metadata]
	setOrderExactMoney(&order, amount)
	absolute := formatAmountLikeDecimal(amount.Abs(), renderedAmount)
	row.Amount = absolute
	if expense {
		row.Amount = "-" + absolute
	}
	putFieldSource(&order, templateFieldSource("amount", slots.Amount))
	return slotMapped{row: row, order: order, absolute: absolute, expense: expense}, nil
}

func metadataNegates(meta map[string]string, sign AmountSign) bool {
	if strings.TrimSpace(sign.Metadata) == "" {
		return false
	}
	got := strings.TrimSpace(meta[sign.Metadata])
	for _, item := range sign.Negate {
		if got == strings.TrimSpace(item) {
			return true
		}
	}
	return false
}

func applySlotPersonalActions(order *ir.Order, row *Row, rule Rule, ignore *bool) error {
	actions := rule.Actions
	if actions.Ignore {
		*ignore = true
	}
	if actions.Payee != "" {
		order.Peer = resolveActionValue(actions.Payee, *row, *order)
		row.Payee = order.Peer
		putFieldSource(order, ruleFieldSource(rule, "payee", actions.Payee))
	}
	if actions.Narration != "" {
		order.Item = resolveActionValue(actions.Narration, *row, *order)
		row.Narration = order.Item
		putFieldSource(order, ruleFieldSource(rule, "narration", actions.Narration))
	}
	if actions.Flag != "" {
		order.Flag = strings.TrimSpace(resolveActionValue(actions.Flag, *row, *order))
		if order.Flag == "" {
			dropFieldSource(order, "flag")
		} else {
			putFieldSource(order, ruleFieldSource(rule, "flag", actions.Flag))
		}
	}
	if actions.Link != "" {
		link := strings.TrimSpace(resolveActionValue(actions.Link, *row, *order))
		if link != "" {
			order.Links = append(order.Links, link)
			putFieldSource(order, ruleFieldSource(rule, "links", actions.Link))
		}
	}
	if actions.Tag != "" {
		tags := splitList(resolveActionValue(actions.Tag, *row, *order))
		if len(tags) > 0 {
			order.Tags = append(order.Tags, tags...)
			putFieldSource(order, ruleFieldSource(rule, "tags", actions.Tag))
		}
	}
	if len(actions.Tags) > 0 {
		order.Tags = append(order.Tags, actions.Tags...)
		putFieldSource(order, ir.FieldSource{Slot: "tags", RuleID: rule.ID, Origin: ir.FieldOriginRule})
	}
	if actions.Note != "" {
		order.Note = resolveActionValue(actions.Note, *row, *order)
		putFieldSource(order, ruleFieldSource(rule, "note", actions.Note))
	}
	if order.Metadata == nil {
		order.Metadata = map[string]string{}
	}
	for key, value := range actions.Metadata {
		rendered := strings.TrimSpace(resolveActionValue(value, *row, *order))
		if rendered == "" {
			delete(order.Metadata, key)
			delete(row.Metadata, key)
			order.MetadataKeys = removeString(order.MetadataKeys, key)
			dropFieldSource(order, metadataSlot(key))
			continue
		}
		order.Metadata[key] = rendered
		if row.Metadata == nil {
			row.Metadata = map[string]string{}
		}
		row.Metadata[key] = rendered
		if !containsString(order.MetadataKeys, key) {
			order.MetadataKeys = append(order.MetadataKeys, key)
		}
		putFieldSource(order, ruleFieldSource(rule, metadataSlot(key), value))
	}
	for _, key := range actions.MetadataDrop {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		delete(order.Metadata, key)
		delete(row.Metadata, key)
		order.MetadataKeys = removeString(order.MetadataKeys, key)
		dropFieldSource(order, metadataSlot(key))
	}
	return nil
}

func fillMissingSides(actions *Actions, expense bool) {
	if strings.TrimSpace(actions.From.Account) == "" {
		if expense {
			actions.From.Account = fallbackAsset
		} else {
			actions.From.Account = fallbackIncome
		}
	}
	if strings.TrimSpace(actions.To.Account) == "" {
		if expense {
			actions.To.Account = fallbackExpense
		} else {
			actions.To.Account = fallbackAsset
		}
	}
}

func metadataKeyOrder(declared []string, values map[string]string) []string {
	out := make([]string, 0, len(declared))
	for _, key := range declared {
		if strings.TrimSpace(values[key]) != "" {
			out = append(out, key)
		}
	}
	return out
}

func templateFieldSource(slot, expr string) ir.FieldSource {
	return ir.FieldSource{
		Slot:    slot,
		Columns: expressionColumns(expr),
		Origin:  ir.FieldOriginTemplate,
	}
}

func ruleFieldSource(rule Rule, slot, expr string) ir.FieldSource {
	return ir.FieldSource{
		Slot:    slot,
		RuleID:  rule.ID,
		Columns: expressionColumns(expr),
		Origin:  ir.FieldOriginRule,
	}
}

func metadataSlot(key string) string {
	return "metadata." + key
}

// putFieldSource keeps the latest writer for a scalar slot. Tags and links
// accumulate one entry per write.
func putFieldSource(order *ir.Order, src ir.FieldSource) {
	if src.Slot == "tags" || src.Slot == "links" {
		order.Sources = append(order.Sources, src)
		return
	}
	for i := range order.Sources {
		if order.Sources[i].Slot == src.Slot {
			order.Sources[i] = src
			return
		}
	}
	order.Sources = append(order.Sources, src)
}

func dropFieldSource(order *ir.Order, slot string) {
	out := make([]ir.FieldSource, 0, len(order.Sources))
	for _, src := range order.Sources {
		if src.Slot != slot {
			out = append(out, src)
		}
	}
	order.Sources = out
}

// expressionColumns lists <列名> references. Bracket refs are logical names,
// and a fully quoted literal contributes no column.
func expressionColumns(expr string) []string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}
	if _, ok := parseActionLiteral(expr); ok {
		return nil
	}
	var cols []string
	seen := map[string]struct{}{}
	for _, match := range columnExprPattern.FindAllStringSubmatch(expr, -1) {
		if !strings.HasPrefix(match[0], "<") {
			continue
		}
		name := strings.TrimSpace(match[2])
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		cols = append(cols, name)
	}
	return cols
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func removeString(values []string, drop string) []string {
	out := values[:0:0]
	for _, value := range values {
		if value != drop {
			out = append(out, value)
		}
	}
	return out
}

var (
	columnRefPattern   = regexp.MustCompile(`<([^>]+)>`)
	metadataRefPattern = regexp.MustCompile(`metadata\.([A-Za-z][A-Za-z0-9_-]*)`)
)

// PersonalRuleWarnings reports personal rules that name a column or metadata
// key the current template no longer provides. Slot names are not warned.
func PersonalRuleWarnings(profile *Profile, rules []Rule, resolvedRef, recordedRef string) []string {
	if profile == nil || !profile.Template.HasSlotContract() {
		return nil
	}
	columns := map[string]struct{}{}
	for _, header := range profile.Template.SourceHeaders {
		header = strings.TrimSpace(header)
		if header != "" {
			columns[header] = struct{}{}
		}
	}
	checkColumns := len(columns) > 0
	metaKeys := map[string]struct{}{}
	for _, key := range profile.Template.Slots.Metadata.Keys {
		metaKeys[key] = struct{}{}
	}
	var warnings []string
	versionNote := ""
	if recordedRef != "" && resolvedRef != "" && recordedRef != resolvedRef {
		versionNote = fmt.Sprintf("（规则文件记录 %s，当前导入 %s）", recordedRef, resolvedRef)
	}
	for _, rule := range rules {
		label := rule.ID
		if label == "" {
			label = rule.When
		}
		texts := []string{rule.When, rule.Actions.Payee, rule.Actions.Narration, rule.Actions.Amount, rule.Actions.From.Account, rule.Actions.To.Account, rule.Actions.Note, rule.Actions.Tag, rule.Actions.Flag, rule.Actions.Link}
		for _, value := range rule.Actions.Metadata {
			texts = append(texts, value)
		}
		for _, text := range texts {
			if checkColumns {
				for _, match := range columnRefPattern.FindAllStringSubmatch(text, -1) {
					name := strings.TrimSpace(match[1])
					if name == "" {
						continue
					}
					if _, ok := columns[name]; ok {
						continue
					}
					warnings = append(warnings, fmt.Sprintf("规则 %q 引用了列 %q，当前模板没有这一列%s", label, name, versionNote))
				}
			}
			for _, match := range metadataRefPattern.FindAllStringSubmatch(text, -1) {
				key := match[1]
				if _, ok := metaKeys[key]; ok {
					continue
				}
				warnings = append(warnings, fmt.Sprintf("规则 %q 引用了元数据 %q，当前模板没有这个键%s", label, key, versionNote))
			}
		}
	}
	return warnings
}

// SlotSkeletonComments documents the slot contract for a generated personal rules file.
func SlotSkeleton(templateRef string, profile *Profile) string {
	var b strings.Builder
	ref := strings.TrimSpace(templateRef)
	if ref == "" {
		ref = profile.ID
	}
	b.WriteString("# template: " + ref + "\n")
	b.WriteString("# 个人规则请引用槽位名和元数据键。账单列名变更时，只要槽位还在就不必改规则。\n")
	b.WriteString("# 直接写 <列名> 时，列被删掉会在导入时提示。\n")
	b.WriteString("# 语法说明: https://deb-sig.github.io/double-entry-generator/providers/template/\n")
	slots := profile.Template.Slots
	b.WriteString("# 槽位映射:\n")
	writeSlotLine(&b, "date", slots.Date)
	writeSlotLine(&b, "payee", slots.Payee)
	writeSlotLine(&b, "narration", slots.Narration)
	writeSlotLine(&b, "amount", slots.Amount)
	writeSlotLine(&b, "currency", slots.Currency)
	writeSlotLine(&b, "flag", slots.Flag)
	if len(slots.Metadata.Keys) > 0 {
		b.WriteString("# 元数据键:\n")
		for _, key := range slots.Metadata.Keys {
			fmt.Fprintf(&b, "#   %s: %s\n", key, slots.Metadata.Values[key])
		}
	}
	fmt.Fprintf(&b, "template: %s\n", ref)
	b.WriteString("personalRules:\n")
	b.WriteString("  - id: 示例\n")
	b.WriteString("    when: payee ~ \"商户名\"\n")
	b.WriteString("    actions:\n")
	b.WriteString("      to: Expenses:FIXME\n")
	return b.String()
}

func writeSlotLine(b *strings.Builder, name, expr string) {
	if strings.TrimSpace(expr) == "" {
		return
	}
	fmt.Fprintf(b, "#   %s: %s\n", name, expr)
}
