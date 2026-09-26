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
		if err := applySlotPersonalActions(&order, &row, rule.Actions, &ignore); err != nil {
			return ir.Order{}, false, err
		}
		if err := mergeV2Actions(&merged, rule.Actions); err != nil {
			return ir.Order{}, false, err
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
	fillMissingSides(&merged, mapped.expense)
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
	}
	order.MetadataKeys = metadataKeyOrder(keys, order.Metadata)

	if slots.Payee != "" {
		order.Peer = strings.TrimSpace(renderRuleText(slots.Payee, row, order))
		row.Payee = order.Peer
	}
	if slots.Narration != "" {
		order.Item = strings.TrimSpace(renderRuleText(slots.Narration, row, order))
		row.Narration = order.Item
	}
	if slots.Flag != "" {
		order.Flag = strings.TrimSpace(renderRuleText(slots.Flag, row, order))
	}
	if slots.Tags != "" {
		order.Tags = splitList(renderRuleText(slots.Tags, row, order))
	}
	if slots.Links != "" {
		order.Links = splitList(renderRuleText(slots.Links, row, order))
	}
	if slots.Currency != "" {
		order.Currency = strings.TrimSpace(renderRuleText(slots.Currency, row, order))
		row.Currency = order.Currency
	}
	if slots.Date != "" {
		rendered := strings.TrimSpace(renderRuleText(slots.Date, row, order))
		row.Date = rendered
		payTime, err := parseDate(rendered, profile.Template.DateFormat)
		if err != nil {
			return slotMapped{}, fmt.Errorf("slots.date %q: %w", slots.Date, err)
		}
		order.PayTime = payTime
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

func applySlotPersonalActions(order *ir.Order, row *Row, actions Actions, ignore *bool) error {
	if actions.Ignore {
		*ignore = true
	}
	if actions.Payee != "" {
		order.Peer = resolveActionValue(actions.Payee, *row, *order)
		row.Payee = order.Peer
	}
	if actions.Narration != "" {
		order.Item = resolveActionValue(actions.Narration, *row, *order)
		row.Narration = order.Item
	}
	if actions.Flag != "" {
		order.Flag = strings.TrimSpace(resolveActionValue(actions.Flag, *row, *order))
	}
	if actions.Link != "" {
		link := strings.TrimSpace(resolveActionValue(actions.Link, *row, *order))
		if link != "" {
			order.Links = append(order.Links, link)
		}
	}
	if actions.Tag != "" {
		order.Tags = append(order.Tags, splitList(resolveActionValue(actions.Tag, *row, *order))...)
	}
	order.Tags = append(order.Tags, actions.Tags...)
	if actions.Note != "" {
		order.Note = resolveActionValue(actions.Note, *row, *order)
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
	}
	for _, key := range actions.MetadataDrop {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		delete(order.Metadata, key)
		delete(row.Metadata, key)
		order.MetadataKeys = removeString(order.MetadataKeys, key)
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
