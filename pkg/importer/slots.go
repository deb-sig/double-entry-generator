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
	if mapped.ignore {
		return order, true, nil
	}

	applyOutputText(profile, &order, &row)

	ignore := false
	merged := Actions{Amount: mapped.absolute, Currency: order.Currency}
	// Role bindings: file-level accounts first, then each matching rule's
	// accounts action, with from/to as shorthand for the from/to roles.
	bound := map[string]roleBinding{}
	directionRule := ""
	for role, account := range profile.Accounts {
		if account = strings.TrimSpace(account); account != "" {
			bound[role] = roleBinding{account: account, origin: ir.FieldOriginRule}
		}
	}
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
		if account := strings.TrimSpace(rule.Actions.Other); account != "" {
			bound["other"] = roleBinding{account: account, origin: ir.FieldOriginRule, ruleID: rule.ID}
		}
		switch d := strings.TrimSpace(rule.Actions.Direction); d {
		case "":
		case "outflow", "inflow":
			mapped.expense = d == "outflow"
			directionRule = rule.ID
		default:
			return ir.Order{}, false, fmt.Errorf("rule %q direction %q: use outflow or inflow", rule.ID, d)
		}
		for role, account := range rule.Actions.Accounts {
			if err := validateRole(role); err != nil {
				return ir.Order{}, false, fmt.Errorf("rule %q accounts: %w", rule.ID, err)
			}
			if account = strings.TrimSpace(account); account != "" {
				bound[role] = roleBinding{account: account, origin: ir.FieldOriginRule, ruleID: rule.ID}
			}
		}
		if account := strings.TrimSpace(rule.Actions.From.Account); account != "" {
			bound["from"] = roleBinding{account: account, origin: ir.FieldOriginRule, ruleID: rule.ID}
			putFieldSource(&order, ruleFieldSource(rule, "from", rule.Actions.From.Account))
		}
		if account := strings.TrimSpace(rule.Actions.To.Account); account != "" {
			bound["to"] = roleBinding{account: account, origin: ir.FieldOriginRule, ruleID: rule.ID}
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
	if directionRule != "" {
		if mapped.expense {
			order.Type = ir.TypeSend
		} else {
			order.Type = ir.TypeRecv
		}
		putFieldSource(&order, ir.FieldSource{Slot: "direction", RuleID: directionRule, Origin: ir.FieldOriginRule})
	}
	if strings.TrimSpace(merged.Amount) == "" {
		merged.Amount = mapped.absolute
	}
	if strings.TrimSpace(merged.Currency) == "" {
		merged.Currency = order.Currency
	}
	if mapped.branch != nil {
		// Multi-leg shape: the template's legs, bound through roles. Rule
		// postings (append mode) still add onto them.
		if err := renderRoleLegs(profile, &order, row, *mapped.branch, bound, merged); err != nil {
			return ir.Order{}, false, err
		}
		applyOutputMetadata(profile, &order)
		return order, false, nil
	}
	// self is the bill's own account and other the counterparty: on an
	// outflow self pays (from) and other receives (to); on an inflow they
	// swap. Explicit from/to always win.
	selfSide, otherSide := &merged.From, &merged.To
	selfSlot, otherSlot := "from", "to"
	if !mapped.expense {
		selfSide, otherSide = &merged.To, &merged.From
		selfSlot, otherSlot = "to", "from"
	}
	if b, ok := bound["self"]; ok && strings.TrimSpace(selfSide.Account) == "" {
		selfSide.Account = b.account
		putFieldSource(&order, ir.FieldSource{Slot: selfSlot, RuleID: b.ruleID, Origin: b.origin})
	}
	if b, ok := bound["other"]; ok && strings.TrimSpace(otherSide.Account) == "" {
		otherSide.Account = b.account
		putFieldSource(&order, ir.FieldSource{Slot: otherSlot, RuleID: b.ruleID, Origin: b.origin})
	}
	if strings.TrimSpace(merged.From.Account) == "" {
		if b, ok := bound["from"]; ok {
			merged.From.Account = b.account
			putFieldSource(&order, ir.FieldSource{Slot: "from", RuleID: b.ruleID, Origin: b.origin})
		}
	}
	if strings.TrimSpace(merged.To.Account) == "" {
		if b, ok := bound["to"]; ok {
			merged.To.Account = b.account
			putFieldSource(&order, ir.FieldSource{Slot: "to", RuleID: b.ruleID, Origin: b.origin})
		}
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
	applyOutputMetadata(profile, &order)
	return order, false, nil
}

// applyOutputText writes the user's default payee and narration, after the
// template mapping and before personal rules, which may still override.
func applyOutputText(profile *Profile, order *ir.Order, row *Row) {
	o := profile.Output
	if o.IsZero() {
		return
	}
	src := func(slot, expr string) ir.FieldSource {
		return ir.FieldSource{Slot: slot, RuleID: "output", Columns: expressionColumns(expr), Origin: ir.FieldOriginRule}
	}
	if o.Payee != "" {
		order.Peer = strings.TrimSpace(resolveActionValue(o.Payee, *row, *order))
		row.Payee = order.Peer
		putFieldSource(order, src("payee", o.Payee))
	}
	if o.Narration != "" {
		order.Item = strings.TrimSpace(resolveActionValue(o.Narration, *row, *order))
		row.Narration = order.Item
		putFieldSource(order, src("narration", o.Narration))
	}
}

// applyOutputMetadata filters metadata last, so keys added by rules are
// covered too. keep wins over drop when both are given.
func applyOutputMetadata(profile *Profile, order *ir.Order) {
	o := profile.Output
	if o.IsZero() || len(o.Metadata.Drop) == 0 && len(o.Metadata.Keep) == 0 {
		return
	}
	keep := map[string]bool{}
	for _, k := range o.Metadata.Keep {
		keep[strings.TrimSpace(k)] = true
	}
	drop := map[string]bool{}
	for _, k := range o.Metadata.Drop {
		drop[strings.TrimSpace(k)] = true
	}
	for key := range order.Metadata {
		remove := drop[key]
		if len(keep) > 0 {
			remove = !keep[key]
		}
		if remove {
			delete(order.Metadata, key)
			order.MetadataKeys = removeString(order.MetadataKeys, key)
			dropFieldSource(order, metadataSlot(key))
		}
	}
}

type slotMapped struct {
	row      Row
	order    ir.Order
	absolute string
	expense  bool
	branch   *LegBranch
	ignore   bool
}

type roleBinding struct {
	account string
	origin  ir.FieldOrigin
	ruleID  string
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
	row, err := rowWithVarSets(row, profile.Template.Vars, order)
	if err != nil {
		return slotMapped{}, err
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
	order.DeclaredMetadataKeys = keys

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
		payTime, err := parseDateIn(rendered, profile.Template.DateFormat, row.Loc)
		if err != nil {
			return slotMapped{}, fmt.Errorf("slots.date %q: %w", slots.Date, err)
		}
		order.PayTime = payTime
		putFieldSource(&order, templateFieldSource("date", slots.Date))
	}

	amount, renderedAmount, expense, err := resolveDirection(profile, row, order)
	if err != nil {
		return slotMapped{}, err
	}
	if amount.Sign() < 0 {
		order.Type = ir.TypeSend
	} else {
		order.Type = ir.TypeRecv
	}
	order.TypeOriginal = directionOriginal(profile, row, order)
	setOrderExactMoney(&order, amount)
	absolute := formatAmountLikeDecimal(amount.Abs(), renderedAmount)
	row.Amount = absolute
	if expense {
		row.Amount = "-" + absolute
	}
	putFieldSource(&order, templateFieldSource("amount", slots.Amount))

	// Template rules in slot mode only shape fields: a conditional payee,
	// a narration with the memo appended, a currency that depends on a
	// column, extra metadata, vars for legs. Accounts are rejected upfront.
	templateIgnore := false
	for _, rule := range enabledRules(applyTemplateRuleOverrides(profile.TemplateRules, profile.TemplateRuleOverrides)) {
		if !ruleInScope(rule, profile.ID) {
			continue
		}
		matches, err := ruleMatches(rule, row, order)
		if err != nil {
			return slotMapped{}, err
		}
		if !matches {
			continue
		}
		if err := applySlotPersonalActions(&order, &row, rule, &templateIgnore); err != nil {
			return slotMapped{}, err
		}
		if rule.Actions.Date != "" {
			rendered := strings.TrimSpace(resolveActionValue(rule.Actions.Date, row, order))
			if payTime, err := parseDateIn(rendered, profile.Template.DateFormat, row.Loc); err == nil {
				order.PayTime = payTime
				row.Date = rendered
				putFieldSource(&order, ruleFieldSource(rule, "date", rule.Actions.Date))
			}
		}
		if rule.Actions.Currency != "" {
			order.Currency = strings.TrimSpace(resolveActionValue(rule.Actions.Currency, row, order))
			row.Currency = order.Currency
			putFieldSource(&order, ruleFieldSource(rule, "currency", rule.Actions.Currency))
		}
		if len(rule.Actions.Vars) > 0 {
			row = rowWithVars(row, rule.Actions.Vars, order)
		}
	}
	if templateIgnore {
		return slotMapped{row: row, order: order, absolute: absolute, expense: expense, ignore: true}, nil
	}

	branch, err := selectLegBranch(profile, row, order)
	if err != nil {
		return slotMapped{}, err
	}
	if branch != nil {
		if branch.Payee != "" {
			order.Peer = strings.TrimSpace(renderRuleText(branch.Payee, row, order))
			row.Payee = order.Peer
			putFieldSource(&order, templateFieldSource("payee", branch.Payee))
		}
		if branch.Narration != "" {
			order.Item = strings.TrimSpace(renderRuleText(branch.Narration, row, order))
			row.Narration = order.Item
			putFieldSource(&order, templateFieldSource("narration", branch.Narration))
		}
		for key, expr := range branch.Metadata {
			value := strings.TrimSpace(renderRuleText(expr, row, order))
			if value == "" {
				continue
			}
			order.Metadata[key] = value
			row.Metadata[key] = value
			if !containsString(order.MetadataKeys, key) {
				order.MetadataKeys = append(order.MetadataKeys, key)
			}
			putFieldSource(&order, templateFieldSource(metadataSlot(key), expr))
		}
	}
	return slotMapped{row: row, order: order, absolute: absolute, expense: expense, branch: branch}, nil
}

// rowWithVarSets applies template var sets in order; a set with a `when`
// only applies where the condition holds, and later sets override.
func rowWithVarSets(row Row, sets VarSets, order ir.Order) (Row, error) {
	for i, set := range sets {
		if strings.TrimSpace(set.When) != "" {
			ok, err := evalWhen(set.When, row, order)
			if err != nil {
				return row, fmt.Errorf("template vars[%d] when %q: %w", i, set.When, err)
			}
			if !ok {
				continue
			}
		}
		row = rowWithVars(row, set.Vars, order)
	}
	return row, nil
}

// resolveDirection parses the amount and decides outflow/inflow according
// to template.direction (or the legacy amountSign). The returned amount is
// signed: negative for outflow.
func resolveDirection(profile *Profile, row Row, order ir.Order) (ir.Decimal, string, bool, error) {
	t := profile.Template
	d := t.Direction
	prefix := t.AmountPrefix

	if d.OutflowColumn != "" {
		out := strings.TrimSpace(renderRuleText(d.OutflowColumn, row, order))
		in := strings.TrimSpace(renderRuleText(d.InflowColumn, row, order))
		// A zero on one side next to a value on the other is a placeholder.
		switch {
		case isNonzeroCell(out, prefix):
			amount, err := parseAmountExact(out, prefix)
			if err != nil {
				return ir.Decimal{}, "", false, fmt.Errorf("direction.outflowColumn %q => %q: %w", d.OutflowColumn, out, err)
			}
			return amount.Abs().Neg(), out, true, nil
		case isNonzeroCell(in, prefix):
			amount, err := parseAmountExact(in, prefix)
			if err != nil {
				return ir.Decimal{}, "", false, fmt.Errorf("direction.inflowColumn %q => %q: %w", d.InflowColumn, in, err)
			}
			return amount.Abs(), in, false, nil
		case nonEmptyAmount(out):
			return ir.Decimal{}, out, true, nil
		case nonEmptyAmount(in):
			return ir.Decimal{}, in, false, nil
		}
		return ir.Decimal{}, "", false, fmt.Errorf("direction: neither %q nor %q has an amount", d.OutflowColumn, d.InflowColumn)
	}

	rendered, err := renderPostingTextStrict(t.Slots.Amount, row, order)
	if err != nil {
		return ir.Decimal{}, "", false, fmt.Errorf("slots.amount %q: %w", t.Slots.Amount, err)
	}
	amount, err := parseAmountExact(rendered, prefix)
	if err != nil {
		return ir.Decimal{}, "", false, fmt.Errorf("slots.amount %q => %q: %w", t.Slots.Amount, rendered, err)
	}

	var expense bool
	switch {
	case d.Column != "":
		got := strings.TrimSpace(renderRuleText(d.Column, row, order))
		switch {
		case containsTrimmed(d.Outflow, got):
			expense = true
		case containsTrimmed(d.Inflow, got):
			expense = false
		case d.Default == "outflow":
			expense = true
		case d.Default == "inflow":
			expense = false
		default:
			// Value in neither list: trust the amount's own sign.
			expense = amount.Sign() < 0
		}
	case !t.AmountSign.IsZero():
		expense = metadataNegates(order.Metadata, t.AmountSign)
	default:
		expense = amount.Sign() < 0
		if d.Invert {
			expense = !expense && amount.Sign() != 0
		}
	}
	if expense {
		amount = amount.Abs().Neg()
	} else {
		amount = amount.Abs()
	}
	return amount, rendered, expense, nil
}

func directionOriginal(profile *Profile, row Row, order ir.Order) string {
	d := profile.Template.Direction
	switch {
	case d.Column != "":
		return strings.TrimSpace(renderRuleText(d.Column, row, order))
	case profile.Template.AmountSign.Metadata != "":
		return order.Metadata[profile.Template.AmountSign.Metadata]
	default:
		return ""
	}
}

func containsTrimmed(values []string, want string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == want {
			return true
		}
	}
	return false
}

func selectLegBranch(profile *Profile, row Row, order ir.Order) (*LegBranch, error) {
	for i := range profile.Template.Legs {
		branch := &profile.Template.Legs[i]
		if strings.TrimSpace(branch.When) == "" {
			return branch, nil
		}
		ok, err := evalWhen(branch.When, row, order)
		if err != nil {
			return nil, fmt.Errorf("template legs[%d] when %q: %w", i, branch.When, err)
		}
		if ok {
			return branch, nil
		}
	}
	return nil, nil
}

// renderRoleLegs turns a leg branch into postings, binding each role to an
// account from the rules file or to the role's FIXME account.
func renderRoleLegs(profile *Profile, order *ir.Order, row Row, branch LegBranch, bound map[string]roleBinding, merged Actions) error {
	row = rowWithVars(row, merged.Vars, *order)
	expense := order.Type == ir.TypeSend
	for i, leg := range branch.Legs {
		role := strings.TrimSpace(leg.Role)
		account, src, err := bindRole(role, bound, expense)
		if err != nil {
			return fmt.Errorf("template legs %q leg %d: %w", branch.ID, i, err)
		}
		putFieldSource(order, src)

		if strings.TrimSpace(leg.Amount) == "" {
			// Bare account: Beancount balances it.
			order.Postings = append(order.Postings, ir.Posting{Line: account})
			continue
		}
		// Render the leg exactly as a hand-written posting line would be, so
		// amounts, costs and prices follow the same rules as rule postings.
		parts := []string{account, strings.TrimSpace(leg.Amount), firstNonEmptyString(leg.Currency, merged.Currency, order.Currency)}
		if cost := strings.TrimSpace(leg.Cost); cost != "" {
			if !strings.HasPrefix(cost, "{") {
				cost = "{" + cost + "}"
			}
			parts = append(parts, cost)
		}
		if price := strings.TrimSpace(leg.Price); price != "" {
			if !strings.HasPrefix(price, "@") {
				price = "@ " + price
			}
			parts = append(parts, price)
		}
		line, err := renderPostingTextStrict(strings.Join(nonEmptyStrings(parts), " "), row, *order)
		if err != nil {
			return fmt.Errorf("leg %s: %w", role, err)
		}
		order.Postings = append(order.Postings, ir.Posting{Line: strings.TrimSpace(line)})
	}
	for _, line := range merged.Postings {
		rendered, err := renderPostingTextStrict(line, row, *order)
		if err != nil {
			return err
		}
		if rendered = strings.TrimSpace(rendered); rendered != "" {
			order.Postings = append(order.Postings, ir.Posting{Line: rendered})
		}
	}
	return nil
}

func bindRole(role string, bound map[string]roleBinding, expense bool) (string, ir.FieldSource, error) {
	slot := "leg." + role
	if b, ok := bound[role]; ok {
		return b.account, ir.FieldSource{Slot: slot, RuleID: b.ruleID, Origin: b.origin}, nil
	}
	fallback, core := CoreRoles[role]
	if !core {
		return "", ir.FieldSource{}, fmt.Errorf("role %q has no account; bind it in the rules file under accounts", role)
	}
	switch role {
	case "other":
		fallback = fallbackIncome
		if expense {
			fallback = fallbackExpense
		}
	case "from":
		fallback = fallbackIncome
		if expense {
			fallback = fallbackAsset
		}
	case "to":
		fallback = fallbackAsset
		if expense {
			fallback = fallbackExpense
		}
	}
	return fallback, ir.FieldSource{Slot: slot, Origin: ir.FieldOriginEngine}, nil
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
		order.MetadataKeys = insertMetadataKey(order.MetadataKeys, declaredMetadataKeys(order), key)
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

// declaredMetadataKeys is the template's full key list, stashed on the
// order so rule-added keys can keep their declared position.
func declaredMetadataKeys(order *ir.Order) []string {
	return order.DeclaredMetadataKeys
}

// insertMetadataKey adds key to keys. A key the template declared goes
// where the declaration puts it relative to the keys already present; an
// undeclared key is appended.
func insertMetadataKey(keys, declared []string, key string) []string {
	if containsString(keys, key) {
		return keys
	}
	pos := -1
	for i, d := range declared {
		if d == key {
			pos = i
			break
		}
	}
	if pos < 0 {
		return append(keys, key)
	}
	after := map[string]bool{}
	for _, d := range declared[pos+1:] {
		after[d] = true
	}
	for i, k := range keys {
		if after[k] {
			out := make([]string, 0, len(keys)+1)
			out = append(out, keys[:i]...)
			out = append(out, key)
			return append(out, keys[i:]...)
		}
	}
	return append(keys, key)
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
	columnRefPattern   = regexp.MustCompile(`<([^<>=!\s][^<>]*)>`)
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
					if name == "" || strings.HasPrefix(name, "file.") || strings.HasPrefix(name, "var.") {
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
	b.WriteString("# 角色绑定：把模板里的角色一次绑到你的账户。没绑的角色会补 FIXME。\n")
	b.WriteString("accounts:\n")
	for _, role := range templateRoles(profile) {
		fmt.Fprintf(&b, "  %s: %s\n", role, roleSkeletonAccount(role))
	}
	b.WriteString("# 输出设置：每笔交易的默认写法和要保留的元数据；个人规则仍可逐笔覆盖。\n")
	b.WriteString("# output:\n")
	if slots.Payee != "" {
		fmt.Fprintf(&b, "#   payee: %s\n", slots.Payee)
	}
	if slots.Narration != "" {
		fmt.Fprintf(&b, "#   narration: %s\n", slots.Narration)
	}
	if len(slots.Metadata.Keys) > 0 {
		fmt.Fprintf(&b, "#   metadata:\n#     drop: [%s]\n", slots.Metadata.Keys[0])
	}
	b.WriteString("rules:\n")
	b.WriteString("  - id: 示例\n")
	b.WriteString("    when: payee ~ \"商户名\"\n")
	b.WriteString("    actions:\n")
	b.WriteString("      to: Expenses:FIXME\n")
	return b.String()
}

// templateRoles lists the roles a template's legs use, from/to first,
// in first-appearance order.
func templateRoles(profile *Profile) []string {
	roles := []string{"self", "other", "from", "to"}
	seen := map[string]bool{"self": true, "other": true, "from": true, "to": true}
	for _, branch := range profile.Template.Legs {
		for _, leg := range branch.Legs {
			role := strings.TrimSpace(leg.Role)
			if role != "" && !seen[role] {
				seen[role] = true
				roles = append(roles, role)
			}
		}
	}
	if len(profile.Template.Legs) > 0 {
		// Multi-leg templates seldom use from/to; list them last.
		return append(roles[4:], roles[:4]...)
	}
	return roles
}

func roleSkeletonAccount(role string) string {
	switch role {
	case "self", "from", "cash", "custody", "position":
		return "Assets:FIXME"
	case "to", "other", "fee", "gas":
		return "Expenses:FIXME"
	case "pnl":
		return "Income:FIXME"
	default:
		return "Equity:FIXME"
	}
}

func writeSlotLine(b *strings.Builder, name, expr string) {
	if strings.TrimSpace(expr) == "" {
		return
	}
	fmt.Fprintf(b, "#   %s: %s\n", name, expr)
}
