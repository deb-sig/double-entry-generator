package importer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/reader"
)

type Row struct {
	Date      string
	Amount    string
	Currency  string
	Payee     string
	Narration string
	Type      string
	Metadata  map[string]string
	Raw       map[string]string
}

func ImportFile(profile *Profile, filename string) (*ir.IR, error) {
	if err := profile.ValidateCapabilities(); err != nil {
		return nil, err
	}
	rows, err := ParseFile(profile, filename)
	if err != nil {
		return nil, err
	}
	orders := ir.New()
	collectRuleOpenAccounts(orders, profile.Rules())
	for _, row := range rows {
		order, ignore, err := rowToImportOrder(profile, row)
		if err != nil {
			if profile.Template.SkipInvalidRows {
				continue
			}
			return nil, err
		}
		if ignore {
			continue
		}
		orders.Orders = append(orders.Orders, order)
	}
	return orders, nil
}

func collectRuleOpenAccounts(orders *ir.IR, rules []Rule) {
	for _, rule := range rules {
		collectStaticAccount(orders, rule.Actions.From.Account)
		collectStaticAccount(orders, rule.Actions.To.Account)
		for _, value := range rule.Actions.Vars {
			collectStaticAccount(orders, value)
		}
		for _, line := range rule.Actions.Postings {
			collectStaticAccount(orders, accountFromPostingTemplate(line))
		}
	}
}

func collectStaticAccount(orders *ir.IR, account string) {
	account = strings.TrimSpace(account)
	if account == "" || strings.Contains(account, "<") || strings.Contains(account, "[") {
		return
	}
	if !isAccountName(account) {
		return
	}
	orders.OpenAccounts[account] = true
}

func accountFromPostingTemplate(line string) string {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func isAccountName(value string) bool {
	switch {
	case strings.HasPrefix(value, "Assets:"),
		strings.HasPrefix(value, "Liabilities:"),
		strings.HasPrefix(value, "Equity:"),
		strings.HasPrefix(value, "Income:"),
		strings.HasPrefix(value, "Expenses:"):
		return true
	default:
		return false
	}
}

// ReaderConfig is the reader block the engine will use for this profile.
// A `reader:` block wins; otherwise the legacy template.* file fields are
// translated so old templates keep working unchanged.
func ReaderConfig(profile *Profile) reader.Config {
	if profile.Reader != nil {
		cfg := *profile.Reader
		if cfg.Format == "" {
			cfg.Format = profile.Template.FileFormat
		}
		if cfg.Encoding == "" {
			cfg.Encoding = profile.Template.Encoding
		}
		if cfg.Delimiter == "" {
			cfg.Delimiter = profile.Template.Delimiter
		}
		cfg.StripTabs = cfg.StripTabs || profile.Template.StripTabs
		cfg.KeepBlankLines = cfg.KeepBlankLines || profile.Template.HeaderLocate
		return cfg.Normalize()
	}
	return reader.Config{
		Format:         profile.Template.FileFormat,
		Encoding:       profile.Template.Encoding,
		Delimiter:      profile.Template.Delimiter,
		StripTabs:      profile.Template.StripTabs,
		KeepBlankLines: profile.Template.HeaderLocate,
	}.Normalize()
}

func ParseFile(profile *Profile, filename string) ([]Row, error) {
	if err := validateBillMatchesTemplate(profile, filename); err != nil {
		return nil, err
	}
	table, err := reader.ReadFile(filename, ReaderConfig(profile))
	if err != nil {
		return nil, err
	}
	return tableToRows(profile, table)
}

// ParseBytes is ParseFile for callers that hold the bill in memory (the
// browser build). name carries the extension used for format checks.
func ParseBytes(profile *Profile, name string, data []byte) ([]Row, error) {
	if err := validateBillMatchesTemplate(profile, name); err != nil {
		return nil, err
	}
	table, err := reader.ReadBytes(name, data, ReaderConfig(profile))
	if err != nil {
		return nil, err
	}
	return tableToRows(profile, table)
}

func tableToRows(profile *Profile, table reader.Table) ([]Row, error) {
	if len(profile.Shape) > 0 {
		headers, rows, err := shapeTable(profile, table)
		if err != nil {
			return nil, err
		}
		if err := validateHeaders(profile, headers); err != nil {
			return nil, err
		}
		return buildRowsFromRecords(profile, headers, rows)
	}
	if table.Headers != nil {
		// Named readers (json, xml, text) already know their columns; there
		// is no header row to locate and nothing to skip.
		headers := normalizeCells(table.Headers)
		if err := validateHeaders(profile, headers); err != nil {
			return nil, err
		}
		return buildRowsFromRecords(profile, headers, table.Rows)
	}
	return recordsToRows(profile, table.Rows)
}

func validateBillMatchesTemplate(profile *Profile, filename string) error {
	cfg := ReaderConfig(profile)
	billFmt := reader.FormatForFile(filename)
	if billFmt == "" {
		return fmt.Errorf("无法识别账单文件格式 %q，请使用 csv、xlsx、xls、json、xml、txt 或 pdf", filepath.Ext(filename))
	}
	if cfg.Compatible(filename) {
		return nil
	}
	templateID := profile.ID
	if templateID == "" {
		templateID = profile.Name
	}
	if templateID == "" {
		templateID = "template"
	}
	return fmt.Errorf(
		"账单文件 %s（%s）与模板 %q 的 fileFormat=%q 不匹配；请将账单导出为 %s，或改用 fileFormat=%q 的模板（本地 profile YAML 或 registry 中的对应模板）",
		filepath.Base(filename),
		billFmt,
		templateID,
		cfg.Format,
		cfg.Format,
		billFmt,
	)
}

func recordsToRows(profile *Profile, records [][]string) ([]Row, error) {
	headers, rows, err := legacyHeaders(profile, records)
	if err != nil {
		return nil, err
	}
	if err := validateHeaders(profile, headers); err != nil {
		return nil, err
	}
	return buildRowsFromRecords(profile, headers, rows)
}

// legacyHeaders applies the template.* header fields (skipLeadingRows,
// sourceHeaders, headerLocate) and returns the header names and the data
// rows after them.
func legacyHeaders(profile *Profile, records [][]string) ([]string, [][]string, error) {
	skip := profile.Template.SkipLeadingRows
	if skip < 0 {
		skip = 0
	}
	if len(records) <= skip {
		return nil, nil, fmt.Errorf("no rows after skipLeadingRows=%d", skip)
	}

	if profile.Template.HeaderLocate {
		return recordsToRowsHeaderLocate(profile, records, skip)
	}

	headers := normalizeCells(profile.Template.SourceHeaders)
	start := skip
	if len(headers) == 0 {
		headers = normalizeCells(records[skip])
		start = skip + 1
	} else if sameCells(headers, normalizeCells(records[skip])) {
		start = skip + 1
	}
	return headers, records[start:], nil
}

func recordsToRowsHeaderLocate(profile *Profile, records [][]string, skip int) ([]string, [][]string, error) {
	wanted := normalizeCells(profile.Template.SourceHeaders)
	if len(wanted) == 0 {
		return nil, nil, fmt.Errorf("headerLocate requires non-empty sourceHeaders")
	}
	wantedSet := make(map[string]struct{}, len(wanted))
	for _, name := range wanted {
		if name == "" {
			continue
		}
		if _, dup := wantedSet[name]; dup {
			return nil, nil, fmt.Errorf("sourceHeaders contains duplicate column name %q", name)
		}
		wantedSet[name] = struct{}{}
	}
	if len(wantedSet) == 0 {
		return nil, nil, fmt.Errorf("headerLocate requires non-empty sourceHeaders")
	}

	scanEnd := len(records)
	if maxRows := profile.Template.HeaderScanMaxRows; maxRows > 0 {
		if skip+maxRows < scanEnd {
			scanEnd = skip + maxRows
		}
	}

	type candidate struct {
		index   int
		headers []string
	}
	var candidates []candidate
	for i := skip; i < scanEnd; i++ {
		headers := normalizeCells(records[i])
		if emptyRecord(headers) {
			continue
		}
		if !rowContainsAllHeaders(headers, wantedSet) {
			continue
		}
		// Candidate header row: duplicate column names fail closed immediately.
		if err := rejectDuplicateHeaderNames(headers); err != nil {
			return nil, nil, fmt.Errorf("header row at index %d: %w", i, err)
		}
		candidates = append(candidates, candidate{index: i, headers: headers})
	}
	switch len(candidates) {
	case 0:
		return nil, nil, fmt.Errorf(
			"headerLocate: no header row containing all sourceHeaders within scan window (skipLeadingRows=%d, headerScanMaxRows=%d)",
			skip, profile.Template.HeaderScanMaxRows,
		)
	case 1:
		// ok
	default:
		idxs := make([]string, len(candidates))
		for i, c := range candidates {
			idxs[i] = fmt.Sprintf("%d", c.index)
		}
		return nil, nil, fmt.Errorf(
			"headerLocate: ambiguous header rows at indices [%s]; refine skipLeadingRows or sourceHeaders",
			strings.Join(idxs, ", "),
		)
	}

	found := candidates[0]
	return found.headers, records[found.index+1:], nil
}

func rejectDuplicateHeaderNames(headers []string) error {
	seen := make(map[string]struct{}, len(headers))
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, ok := seen[h]; ok {
			return fmt.Errorf("duplicate column name %q", h)
		}
		seen[h] = struct{}{}
	}
	return nil
}

func rowContainsAllHeaders(headers []string, wanted map[string]struct{}) bool {
	available := make(map[string]struct{}, len(headers))
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		available[h] = struct{}{}
	}
	for name := range wanted {
		if _, ok := available[name]; !ok {
			return false
		}
	}
	return true
}

func buildRowsFromRecords(profile *Profile, headers []string, records [][]string) ([]Row, error) {
	rows := make([]Row, 0, len(records))
	for _, record := range records {
		record = normalizeCells(record)
		if emptyRecord(record) {
			continue
		}
		raw := make(map[string]string, len(headers))
		for i, h := range headers {
			raw[h] = cell(record, i)
		}
		metadata := map[string]string{}
		if !profile.IsV2() {
			metadata = make(map[string]string, len(profile.Template.Metadata))
			for key, source := range profile.Template.Metadata {
				metadata[key] = raw[source]
			}
		}
		date := raw[profile.Template.Columns.Date]
		if profile.Template.Columns.Time != "" && raw[profile.Template.Columns.Time] != "" {
			date = strings.TrimSpace(date + " " + raw[profile.Template.Columns.Time])
		}
		row := Row{
			Date:      date,
			Amount:    rowAmount(profile, raw),
			Currency:  raw[profile.Template.Columns.Currency],
			Payee:     raw[profile.Template.Columns.Payee],
			Narration: raw[profile.Template.Columns.Narration],
			Type:      rowType(profile, raw),
			Metadata:  metadata,
			Raw:       raw,
		}
		if !profile.IsV2() && strings.TrimSpace(row.Amount) == "" {
			continue
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func validateHeaders(profile *Profile, headers []string) error {
	available := map[string]bool{}
	for _, header := range headers {
		available[header] = true
	}
	required := []string{
		profile.Template.Columns.Date,
		profile.Template.Columns.Amount,
		profile.Template.Columns.AmountIn,
		profile.Template.Columns.AmountOut,
		profile.Template.Columns.Payee,
		profile.Template.Columns.Narration,
		profile.Template.Columns.Type,
		profile.Template.Columns.Currency,
	}
	for _, header := range required {
		if header != "" && !available[header] {
			return fmt.Errorf("template header %q not found after skipLeadingRows=%d", header, profile.Template.SkipLeadingRows)
		}
	}
	return nil
}

func sameCells(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

func rowAmount(profile *Profile, raw map[string]string) string {
	if profile.Template.Columns.Amount != "" {
		return raw[profile.Template.Columns.Amount]
	}
	if profile.Template.Columns.AmountOut != "" {
		if value := strings.TrimSpace(raw[profile.Template.Columns.AmountOut]); value != "" && value != "-" {
			return "-" + strings.TrimPrefix(value, "-")
		}
	}
	if profile.Template.Columns.AmountIn != "" {
		value := strings.TrimSpace(raw[profile.Template.Columns.AmountIn])
		if value == "-" {
			return ""
		}
		return value
	}
	return ""
}

func rowType(profile *Profile, raw map[string]string) string {
	if profile.Template.Columns.Type != "" {
		return raw[profile.Template.Columns.Type]
	}
	if profile.Template.Columns.AmountOut != "" && nonEmptyAmount(raw[profile.Template.Columns.AmountOut]) {
		return "支出"
	}
	if profile.Template.Columns.AmountIn != "" && nonEmptyAmount(raw[profile.Template.Columns.AmountIn]) {
		return "收入"
	}
	return ""
}

func nonEmptyAmount(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && value != "-"
}

func rowToImportOrder(profile *Profile, row Row) (ir.Order, bool, error) {
	if !profile.IsV2() {
		return rowToOrder(profile, row)
	}
	return rowToV2Order(profile, row)
}

func rowToV2Order(profile *Profile, row Row) (ir.Order, bool, error) {
	order := ir.Order{
		OrderType: ir.OrderTypeNormal,
		Currency:  profile.Template.DefaultCurrency,
		Metadata:  row.Metadata,
	}
	if order.Metadata == nil {
		order.Metadata = map[string]string{}
	}
	if row.Date != "" {
		if payTime, err := parseDate(row.Date, profile.Template.DateFormat); err == nil {
			order.PayTime = payTime
		}
	}
	if row.Amount != "" {
		amount, err := parseAmountExact(row.Amount, profile.Template.AmountPrefix)
		if err != nil {
			return ir.Order{}, false, fmt.Errorf("parse amount %q failed for date=%q payee=%q: %w", row.Amount, row.Date, row.Payee, err)
		}
		order.Type = inferTypeDecimal(row.Type, amount)
		order.TypeOriginal = row.Type
		setOrderExactMoney(&order, amount)
	}
	order.Peer = row.Payee
	order.Item = row.Narration
	if row.Currency != "" {
		order.Currency = row.Currency
	}

	ignore := false
	mergedV2Actions := Actions{}
	for _, rule := range profile.Rules() {
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
		if err := applyV2ScalarActions(&order, row, rule.Actions, &ignore, profile.Template.DateFormat); err != nil {
			return ir.Order{}, false, err
		}
		if err := mergeV2Actions(&mergedV2Actions, rule.Actions); err != nil {
			return ir.Order{}, false, err
		}
	}
	if ignore {
		return order, true, nil
	}
	if order.PayTime.IsZero() {
		return ir.Order{}, false, fmt.Errorf("runtime v2 rule did not set date")
	}
	if err := renderV2Postings(&order, row, mergedV2Actions); err != nil {
		return ir.Order{}, false, err
	}
	return order, false, nil
}

func rowToOrder(profile *Profile, row Row) (ir.Order, bool, error) {
	amount, err := parseAmountExact(row.Amount, profile.Template.AmountPrefix)
	if err != nil {
		return ir.Order{}, false, fmt.Errorf("parse amount %q failed for date=%q payee=%q: %w", row.Amount, row.Date, row.Payee, err)
	}
	txType := inferTypeDecimal(row.Type, amount)
	payTime, err := parseDate(row.Date, profile.Template.DateFormat)
	if err != nil {
		return ir.Order{}, false, err
	}
	order := ir.Order{
		OrderType:    ir.OrderTypeNormal,
		Peer:         row.Payee,
		Item:         row.Narration,
		PayTime:      payTime,
		Type:         txType,
		TypeOriginal: row.Type,
		Currency:     profile.Template.DefaultCurrency,
		Metadata:     row.Metadata,
	}
	setOrderExactMoney(&order, amount)
	order.MinusAccount = profile.Template.DefaultMinus
	order.PlusAccount = profile.Template.DefaultPlus
	if row.Currency != "" {
		order.Currency = row.Currency
	}
	if order.Metadata == nil {
		order.Metadata = map[string]string{}
	}

	ignore := false
	for _, rule := range profile.Rules() {
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
		if err := applyActions(&order, row, rule.Actions, &ignore, false); err != nil {
			return ir.Order{}, false, err
		}
	}
	return order, ignore, nil
}

func ruleMatches(rule Rule, row Row, order ir.Order) (bool, error) {
	if strings.TrimSpace(rule.When) == "" {
		return true, nil
	}
	ok, err := evalWhen(rule.When, row, order)
	if err != nil {
		if rule.ID != "" {
			return false, fmt.Errorf("rule %q when %q failed: %w", rule.ID, rule.When, err)
		}
		return false, fmt.Errorf("rule when %q failed: %w", rule.When, err)
	}
	return ok, nil
}

func ruleInScope(rule Rule, profileID string) bool {
	if rule.TemplateID == "" {
		return true
	}
	return profileID != "" && rule.TemplateID == profileID
}

func applyActions(order *ir.Order, row Row, actions Actions, ignore *bool, v2 bool) error {
	// Sticky ignore (Mirato): true latches; later false/absent never clears.
	// Later matching rules still run (accounts/tags may change) but order is dropped.
	if actions.Ignore {
		*ignore = true
	}
	if actions.Type != "" {
		if d, ok := orderMoneyDecimal(*order); ok {
			order.Type = inferTypeDecimal(actions.Type, d)
		} else {
			order.Type = inferType(actions.Type, order.Money)
		}
		order.TypeOriginal = actions.Type
	}
	if actions.Note != "" {
		order.Note = resolveValue(actions.Note, row)
	}
	if actions.Payee != "" {
		order.Peer = resolveValue(actions.Payee, row)
	}
	if actions.Narration != "" {
		order.Item = resolveValue(actions.Narration, row)
	}
	if actions.Amount != "" {
		amount, err := parseAmountExact(resolveValue(actions.Amount, row), "")
		if err != nil {
			return fmt.Errorf("actions.amount %q: %w", actions.Amount, err)
		}
		setOrderExactMoney(order, amount)
	}
	if actions.Currency != "" {
		order.Currency = resolveValue(actions.Currency, row)
	}
	if !actions.To.IsZero() {
		if v2 {
			posting, err := renderTransferPosting(actions.To, actions.Amount, actions.Currency, "+", row, *order)
			if err != nil {
				return err
			}
			if posting != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		} else {
			order.PlusAccount = resolveValue(actions.To.Account, row)
		}
	}
	if !actions.From.IsZero() {
		if v2 {
			posting, err := renderTransferPosting(actions.From, actions.Amount, actions.Currency, "-", row, *order)
			if err != nil {
				return err
			}
			if posting != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		} else {
			order.MinusAccount = resolveValue(actions.From.Account, row)
		}
	}
	if v2 {
		for _, line := range actions.Postings {
			rendered, err := renderPostingTextStrict(line, row, *order)
			if err != nil {
				return err
			}
			rendered = strings.TrimSpace(rendered)
			if rendered != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: rendered})
			}
		}
	}
	if actions.Tag != "" {
		if v2 {
			order.Tags = append(order.Tags, splitList(resolveActionValue(actions.Tag, row, *order))...)
		} else {
			order.Tags = append(order.Tags, splitList(actions.Tag)...)
		}
	}
	order.Tags = append(order.Tags, actions.Tags...)
	if actions.Metadata != nil {
		if order.Metadata == nil {
			order.Metadata = map[string]string{}
		}
		for key, value := range actions.Metadata {
			if v2 {
				order.Metadata[key] = resolveActionValue(value, row, *order)
			} else {
				order.Metadata[key] = resolveValue(value, row)
			}
		}
	}
	return nil
}

func applyV2ScalarActions(order *ir.Order, row Row, actions Actions, ignore *bool, dateFormat string) error {
	// Sticky ignore — see applyActions.
	if actions.Ignore {
		*ignore = true
	}
	if actions.Date != "" {
		if payTime, err := parseDate(resolveActionValue(actions.Date, row, *order), dateFormat); err == nil {
			order.PayTime = payTime
		}
	}
	if actions.Amount != "" {
		rendered, err := renderPostingTextStrict(actions.Amount, row, *order)
		if err != nil {
			return fmt.Errorf("actions.amount %q: %w", actions.Amount, err)
		}
		amount, err := parseAmountExact(rendered, "")
		if err != nil {
			return fmt.Errorf("actions.amount %q => %q: %w", actions.Amount, rendered, err)
		}
		order.Type = inferTypeDecimal(order.TypeOriginal, amount)
		setOrderExactMoney(order, amount)
	}
	if actions.Type != "" {
		order.TypeOriginal = resolveActionValue(actions.Type, row, *order)
		if d, ok := orderMoneyDecimal(*order); ok {
			order.Type = inferTypeDecimal(order.TypeOriginal, d)
		} else {
			order.Type = inferType(order.TypeOriginal, order.Money)
		}
	}
	if actions.Note != "" {
		order.Note = resolveActionValue(actions.Note, row, *order)
	}
	if actions.Payee != "" {
		order.Peer = resolveActionValue(actions.Payee, row, *order)
	}
	if actions.Narration != "" {
		order.Item = resolveActionValue(actions.Narration, row, *order)
	}
	if actions.Currency != "" {
		order.Currency = resolveActionValue(actions.Currency, row, *order)
	}
	if actions.Tag != "" {
		order.Tags = append(order.Tags, splitList(resolveActionValue(actions.Tag, row, *order))...)
	}
	order.Tags = append(order.Tags, actions.Tags...)
	if actions.Flag != "" {
		order.Flag = strings.TrimSpace(resolveActionValue(actions.Flag, row, *order))
	}
	if actions.Link != "" {
		link := strings.TrimSpace(resolveActionValue(actions.Link, row, *order))
		if link != "" {
			order.Links = append(order.Links, link)
		}
	}
	if actions.Metadata != nil {
		if order.Metadata == nil {
			order.Metadata = map[string]string{}
		}
		for key, value := range actions.Metadata {
			// Mirato retains empty metadata values; do not delete on "".
			order.Metadata[key] = resolveActionValue(value, row, *order)
		}
	}
	return nil
}

func mergeV2Actions(base *Actions, next Actions) error {
	if !next.From.IsZero() {
		base.From = mergeTransferSide(base.From, next.From)
	}
	if !next.To.IsZero() {
		base.To = mergeTransferSide(base.To, next.To)
	}
	if next.Amount != "" {
		base.Amount = next.Amount
	}
	if next.Currency != "" {
		base.Currency = next.Currency
	}
	if len(next.Vars) > 0 {
		if base.Vars == nil {
			base.Vars = map[string]string{}
		}
		for key, value := range next.Vars {
			base.Vars[key] = value
		}
	}
	// Empty replace must clear the complete set (Mirato semantics). Do not
	// require len(Postings)>0 — that silently ignored replace: [].
	modeRaw := strings.TrimSpace(next.PostingsMode)
	if len(next.Postings) > 0 || modeRaw != "" {
		mode, err := normalizePostingsMode(next.PostingsMode)
		if err != nil {
			return err
		}
		if mode == "replace" {
			base.Postings = append([]string(nil), next.Postings...)
			base.PostingsMode = "replace"
		} else if len(next.Postings) > 0 {
			base.Postings = append(base.Postings, next.Postings...)
			if base.PostingsMode == "" {
				base.PostingsMode = "append"
			}
		}
	}
	return nil
}

func mergeTransferSide(base, next TransferSide) TransferSide {
	if next.Account != "" {
		base.Account = next.Account
	}
	if next.Amount != "" {
		base.Amount = next.Amount
	}
	if next.Currency != "" {
		base.Currency = next.Currency
	}
	return base
}

func renderV2Postings(order *ir.Order, row Row, actions Actions) error {
	row = rowWithVars(row, actions.Vars, *order)
	mode, err := normalizePostingsMode(actions.PostingsMode)
	if err != nil {
		return err
	}
	// replace (even empty): Mirato complete leg list — do not also render
	// automatic from/to (empty replace clears legs; non-empty replaces them).
	skipAutoLegs := mode == "replace"
	ccyFallback := firstNonEmptyString(actions.Currency, order.Currency)
	if !skipAutoLegs {
		if !actions.To.IsZero() {
			posting, err := renderTransferPosting(actions.To, actions.Amount, ccyFallback, "+", row, *order)
			if err != nil {
				return err
			}
			if posting != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		}
		if !actions.From.IsZero() {
			posting, err := renderTransferPosting(actions.From, actions.Amount, ccyFallback, "-", row, *order)
			if err != nil {
				return err
			}
			if posting != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		}
	}
	for _, line := range actions.Postings {
		rendered, err := renderPostingTextStrict(line, row, *order)
		if err != nil {
			return err
		}
		rendered = strings.TrimSpace(rendered)
		if rendered != "" {
			order.Postings = append(order.Postings, ir.Posting{Line: rendered})
		}
	}
	return nil
}

func rowWithVars(row Row, vars map[string]string, order ir.Order) Row {
	if len(vars) == 0 {
		return row
	}
	raw := make(map[string]string, len(row.Raw)+len(vars))
	for key, value := range row.Raw {
		raw[key] = value
	}
	withVars := row
	withVars.Raw = raw
	for key, value := range vars {
		raw["var."+key] = renderPostingText(value, withVars, order)
	}
	return withVars
}

func renderTransferPosting(side TransferSide, defaultAmount, defaultCurrency, direction string, row Row, order ir.Order) (string, error) {
	account := strings.TrimSpace(resolveActionValue(side.Account, row, order))
	if account == "" {
		return "", nil
	}
	amount := firstNonEmptyString(side.Amount, defaultAmount)
	currency := firstNonEmptyString(side.Currency, defaultCurrency)
	if amount == "" {
		amount = "[amount].number"
	}
	amount = forceAmountDirection(amount, direction)
	renderedAmount, err := renderPostingTextStrict(amount, row, order)
	if err != nil {
		return "", fmt.Errorf("transfer amount %q: %w", amount, err)
	}
	if _, err := parseAmountExact(renderedAmount, ""); err != nil {
		return "", fmt.Errorf("transfer amount %q => %q: %w", amount, renderedAmount, err)
	}
	parts := []string{account, renderedAmount}
	if currency != "" {
		parts = append(parts, resolveActionValue(currency, row, order))
	}
	return strings.Join(nonEmptyStrings(parts), " "), nil
}

func forceAmountDirection(expr, direction string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return expr
	}
	if strings.Contains(expr, ".+") || strings.Contains(expr, ".-") || strings.Contains(expr, ".!") {
		return expr
	}
	if strings.HasPrefix(expr, "+") || strings.HasPrefix(expr, "-") || strings.HasPrefix(expr, "!") {
		return expr
	}
	loc := columnExprPattern.FindStringIndex(expr)
	if loc != nil {
		return expr[:loc[1]] + "." + direction + expr[loc[1]:]
	}
	if direction == "-" {
		return "-(" + expr + ")"
	}
	return expr
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return out
}

// columnLookupMode distinguishes Mirato shared protocols so one <col> helper
// cannot silently mix condition logical fallback with action raw-only refs.
type columnLookupMode int

const (
	lookupCondition columnLookupMode = iota
	lookupActionRaw
)

// fieldValue is the shared condition lookup (Mirato baseline). Prefer
// conditionFieldValue / actionColumnValue at call sites that know the context.
func fieldValue(field string, row Row, order ir.Order) string {
	return conditionFieldValue(field, row, order)
}

// conditionFieldValue implements Mirato↔DEG shared condition lookup:
//  1. exact Raw key wins (including present empty string);
//  2. else exact lowercase logical payee|narration|amount|date|currency;
//  3. else "" — custom columns keep literal identity (no raw./metadata.
//     namespace strip, no case fold, no peer/item aliases).
//
// date.time / date.date / date.timestamp suffixes remain for DEG native when.
func conditionFieldValue(field string, row Row, order ir.Order) string {
	field = strings.TrimSpace(field)
	if base, suffix, ok := strings.Cut(field, "."); ok && (suffix == "time" || suffix == "date" || suffix == "timestamp") {
		value := conditionFieldValue(base, row, order)
		if base == "date" || base == "交易时间" || value == "" {
			if suffix == "time" {
				return order.PayTime.Format("15:04")
			}
			if suffix == "timestamp" {
				return strconv.FormatInt(order.PayTime.Unix(), 10)
			}
			return order.PayTime.Format("2006-01-02")
		}
		if t, err := parseDate(value, ""); err == nil {
			if suffix == "time" {
				return t.Format("15:04")
			}
			if suffix == "timestamp" {
				return strconv.FormatInt(t.Unix(), 10)
			}
			return t.Format("2006-01-02")
		}
	}
	if v, ok := row.Raw[field]; ok {
		return v
	}
	switch field {
	// 不可变原始值：`original_*` / `imported_*` 永远读账单原始行，不受前面规则改写影响。
	// 顺序语义（本批确认：规则按序执行、后面的条件看到前面的改写）下，作者要用原始值
	// 匹配就必须有显式入口 —— 与 Actual Budget 的 `imported payee` 同一套思路。
	case "original_amount", "imported_amount":
		return row.Amount
	case "original_payee", "imported_payee":
		return row.Payee
	case "original_narration", "imported_narration":
		return row.Narration
	case "original_date", "imported_date":
		return row.Date
	case "original_currency", "imported_currency":
		return row.Currency
	case "original_type", "imported_type":
		return row.Type
	case "date":
		if !order.PayTime.IsZero() {
			return order.PayTime.Format("2006-01-02")
		}
		return row.Date
	case "amount":
		if order.ExactMoney != nil {
			return order.ExactMoney.Text(0)
		}
		return row.Amount
	case "currency":
		if order.Currency != "" {
			return order.Currency
		}
		return row.Currency
	case "payee":
		if order.Peer != "" {
			return order.Peer
		}
		return row.Payee
	case "narration":
		if order.Item != "" {
			return order.Item
		}
		return row.Narration
	default:
		if order.Metadata != nil {
			if v, ok := order.Metadata[field]; ok {
				return v
			}
			if v, ok := order.Metadata["metadata."+field]; ok {
				return v
			}
		}
		return ""
	}
}

// actionColumnValue implements Mirato __from_column / __regex / __replace:
// column refs resolve against Raw only. Present empty stays empty; absent → "".
// Never falls back to logical Row/Order fields (that is condition-only).
func actionColumnValue(field string, row Row) string {
	field = strings.TrimSpace(field)
	if v, ok := row.Raw[field]; ok {
		return v
	}
	return ""
}

func parseAmount(value, prefix string) (float64, error) {
	d, err := ParseAmountDecimal(value, prefix)
	if err != nil {
		return 0, err
	}
	return d.Float64Approx(), nil
}

func parseAmountExact(value, prefix string) (ir.Decimal, error) {
	return ParseAmountDecimal(value, prefix)
}

func parseDate(value, layout string) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		normalizeDateLayout(layout),
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"20060102 15:04:05",
		"20060102 150405",
		"20060102",
		"01/02/2006",
		"02/01/2006",
		"01/02/2006 15:04:05",
		"02/01/2006 15:04:05",
		"01/02",
		"02/01",
		time.RFC3339,
	}
	for _, candidate := range layouts {
		if candidate == "" {
			continue
		}
		if t, err := time.ParseInLocation(candidate, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse date %q failed", value)
}

func inferType(value string, amount float64) ir.Type {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "recv", "income", "in", "收入", "收", "入账":
		return ir.TypeRecv
	case "send", "expense", "out", "支出", "支", "出账":
		return ir.TypeSend
	}
	if amount < 0 {
		return ir.TypeSend
	}
	return ir.TypeSend
}

func inferTypeDecimal(value string, amount ir.Decimal) ir.Type {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "recv", "income", "in", "收入", "收", "入账":
		return ir.TypeRecv
	case "send", "expense", "out", "支出", "支", "出账":
		return ir.TypeSend
	}
	if amount.Sign() < 0 {
		return ir.TypeSend
	}
	return ir.TypeSend
}

func resolveValue(value string, row Row) string {
	const prefix = "__from_column:"
	if strings.HasPrefix(value, prefix) {
		return row.Raw[strings.TrimSpace(strings.TrimPrefix(value, prefix))]
	}
	if strings.HasPrefix(value, "raw.") {
		return row.Raw[strings.TrimPrefix(value, "raw.")]
	}
	if strings.HasPrefix(value, "raw[") && strings.HasSuffix(value, "]") {
		field := strings.TrimSuffix(strings.TrimPrefix(value, "raw["), "]")
		return row.Raw[field]
	}
	if strings.HasPrefix(value, "account:") {
		return value
	}
	return value
}

func normalizeDateLayout(layout string) string {
	layout = strings.TrimSpace(layout)
	replacer := strings.NewReplacer(
		"yyyy", "2006",
		"YYYY", "2006",
		"MM", "01",
		"dd", "02",
		"DD", "02",
		"HH", "15",
		"hh", "15",
		"mm", "04",
		"ss", "05",
	)
	return replacer.Replace(layout)
}

func normalizeCells(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		value = strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = strings.TrimSuffix(strings.TrimPrefix(value, `"`), `"`)
		}
		if strings.HasPrefix(value, `="`) && strings.HasSuffix(value, `"`) {
			value = strings.TrimSuffix(strings.TrimPrefix(value, `="`), `"`)
		}
		out[i] = value
	}
	return out
}

func emptyRecord(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func cell(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return strings.TrimSpace(values[index])
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

var columnExprPattern = regexp.MustCompile(`(?:\[([^\]]+)\]|<([^>]+)>)((?:\.(?:extract|format)\((?:r)?"[^"]*"\)|\.(?:extract|format)\((?:r)?'[^']*'\)|\.replace\((?:r)?"[^"]*",(?:r)?"[^"]*"\)|\.replace\((?:r)?'[^']*',(?:r)?'[^']*'\)|\.[A-Za-z0-9_]+|\.[+\-!])*)`)

// resolveActionValue implements the Mirato↔DEG action literal/ref protocol:
//   - fully quoted "..." / '...' => string literal (escapes: \\ \" \' \n \r \t)
//   - otherwise interpolate <col> / [col] refs (and methods) via renderRuleText
//
// Fixed text such as "<金额>", "payee", or "1+2" must be quoted so it is not
// treated as a column ref or arithmetic expression.
func resolveActionValue(value string, row Row, order ir.Order) string {
	if lit, ok := parseActionLiteral(value); ok {
		return lit
	}
	return renderRuleText(value, row, order)
}

func renderRuleTextMode(value string, row Row, order ir.Order, mode columnLookupMode) string {
	return columnExprPattern.ReplaceAllStringFunc(value, func(match string) string {
		return evalColumnString(match, row, order, mode)
	})
}

func parseActionLiteral(value string) (string, bool) {
	if len(value) < 2 {
		return "", false
	}
	quote := value[0]
	if quote != '"' && quote != '\'' {
		return "", false
	}
	if value[len(value)-1] != quote {
		return "", false
	}
	var b strings.Builder
	escaped := false
	for i := 1; i < len(value)-1; i++ {
		c := value[i]
		if escaped {
			switch c {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\', '"', '\'':
				b.WriteByte(c)
			default:
				b.WriteByte('\\')
				b.WriteByte(c)
			}
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == quote {
			// Unescaped closing quote before the end => not a single literal.
			return "", false
		}
		b.WriteByte(c)
	}
	if escaped {
		b.WriteByte('\\')
	}
	return b.String(), true
}

func renderRuleText(value string, row Row, order ir.Order) string {
	// Explicit protocol by delimiter (do not unify into one lookup):
	//   <col>  → Mirato __from_column/__regex/__replace family: Raw only
	//   [col]  → DEG-native refs (e.g. [amount].number on transfer legs):
	//            shared condition lookup (raw then five logicals)
	return columnExprPattern.ReplaceAllStringFunc(value, func(match string) string {
		mode := lookupActionRaw
		if strings.HasPrefix(match, "[") {
			mode = lookupCondition
		}
		return evalColumnString(match, row, order, mode)
	})
}

func renderPostingText(value string, row Row, order ir.Order) string {
	out, err := renderPostingTextStrict(value, row, order)
	if err != nil {
		// Soft legacy path for non-money call sites (metadata/note): leave text unchanged.
		if lit, ok := parseActionLiteral(value); ok {
			return lit
		}
		return evalSimpleArithmetic(renderRuleText(value, row, order))
	}
	return out
}

func renderPostingTextStrict(value string, row Row, order ir.Order) (string, error) {
	if lit, ok := parseActionLiteral(value); ok {
		// Quoted literals skip interpolation/arithmetic but still validate money.
		if err := validatePostingAmountTokens(lit); err != nil {
			return "", err
		}
		return lit, nil
	}
	rendered := renderRuleText(value, row, order)
	// An account is an identifier, not an arithmetic region. Numeric and
	// hyphenated account segments must survive posting evaluation unchanged.
	// Account-only postings (elided amount) must never be rewritten.
	prefix := ""
	trimmed := strings.TrimLeftFunc(rendered, unicode.IsSpace)
	leadingWS := rendered[:len(rendered)-len(trimmed)]
	if end := strings.IndexFunc(trimmed, unicode.IsSpace); end < 0 {
		if strings.Contains(trimmed, ":") {
			out := leadingWS + trimmed
			if err := validatePostingAmountTokens(out); err != nil {
				return "", err
			}
			return out, nil
		}
	} else if strings.Contains(trimmed[:end], ":") {
		prefix = leadingWS + trimmed[:end]
		rendered = trimmed[end:]
	}
	out, err := evalArithmeticInTextStrict(rendered)
	if err != nil {
		return "", err
	}
	out = prefix + out
	if err := validatePostingAmountTokens(out); err != nil {
		return "", err
	}
	return out, nil
}

func evalColumnString(expr string, row Row, order ir.Order, mode columnLookupMode) string {
	if !strings.HasPrefix(expr, "[") && !strings.HasPrefix(expr, "<") {
		return expr
	}
	close := "]"
	if strings.HasPrefix(expr, "<") {
		close = ">"
	}
	end := strings.Index(expr, close)
	if end < 0 {
		return expr
	}
	field := expr[1:end]
	var value string
	switch mode {
	case lookupActionRaw:
		value = actionColumnValue(field, row)
	default:
		value = conditionFieldValue(field, row, order)
	}
	rest := expr[end+1:]
	for rest != "" {
		if !strings.HasPrefix(rest, ".") {
			break
		}
		rest = strings.TrimPrefix(rest, ".")
		method, arg, tail := nextMethod(rest)
		value = applyColumnMethod(value, method, arg, row, order)
		rest = tail
	}
	return value
}

func nextMethod(rest string) (string, string, string) {
	if strings.HasPrefix(rest, "extract(") || strings.HasPrefix(rest, "format(") || strings.HasPrefix(rest, "replace(") {
		name, _, _ := strings.Cut(rest, "(")
		end := closingMethodParen(rest)
		if end < 0 {
			return rest, "", ""
		}
		arg := rest[len(name)+1 : end]
		if name != "replace" {
			arg = strings.Trim(arg, `"'`)
		}
		return name, arg, rest[end+1:]
	}
	if rest != "" && (rest[0] == '+' || rest[0] == '-' || rest[0] == '!') {
		return rest[:1], "", rest[1:]
	}
	i := 0
	for i < len(rest) && isMethodIdentByte(rest[i]) {
		i++
	}
	return rest[:i], "", rest[i:]
}

func isMethodIdentByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func closingMethodParen(value string) int {
	var quote byte
	start := strings.Index(value, "(")
	if start < 0 {
		return -1
	}
	for i := start + 1; i < len(value); i++ {
		c := value[i]
		if quote != 0 {
			if c == '\\' && i+1 < len(value) {
				i++
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == ')' {
			return i
		}
	}
	return -1
}

func applyColumnMethod(value, method, arg string, row Row, order ir.Order) string {
	switch method {
	case "trim":
		return strings.TrimSpace(value)
	case "number":
		return normalizeAmountString(value)
	case "+":
		n, err := parseAmountExact(value, "")
		if err != nil {
			return value
		}
		return formatAmountLikeDecimal(n.Abs(), value)
	case "-":
		n, err := parseAmountExact(value, "")
		if err != nil {
			return value
		}
		return formatAmountLikeDecimal(n.Abs().Neg(), value)
	case "!":
		n, err := parseAmountExact(value, "")
		if err != nil {
			return value
		}
		return formatAmountLikeDecimal(n.Neg(), value)
	case "format":
		return formatValue(value, arg)
	case "date", "time", "timestamp":
		if t, err := parseDate(value, ""); err == nil {
			switch method {
			case "date":
				return t.Format("2006-01-02")
			case "time":
				return t.Format("15:04")
			case "timestamp":
				return strconv.FormatInt(t.Unix(), 10)
			}
		}
	case "extract":
		pattern := strings.TrimSpace(arg)
		pattern = strings.TrimPrefix(pattern, "r")
		pattern = strings.Trim(pattern, `"'`)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return ""
		}
		matches := re.FindStringSubmatch(value)
		if len(matches) > 1 {
			return matches[1]
		}
		if len(matches) == 1 {
			return matches[0]
		}
		return ""
	case "replace":
		from, to, ok := splitReplaceArgs(arg)
		if !ok {
			return value
		}
		return strings.ReplaceAll(value, from, to)
	}
	return value
}

func splitReplaceArgs(arg string) (string, string, bool) {
	arg = strings.TrimSpace(arg)
	parts := []string{}
	var quote byte
	var b strings.Builder
	started := false
	for i := 0; i < len(arg); i++ {
		c := arg[i]
		if quote != 0 {
			if c == '\\' && i+1 < len(arg) {
				b.WriteByte(arg[i+1])
				i++
				continue
			}
			if c == quote {
				parts = append(parts, b.String())
				b.Reset()
				quote = 0
				started = false
				continue
			}
			b.WriteByte(c)
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			started = true
			continue
		}
		if c == ',' || c == ' ' || c == '\t' {
			continue
		}
		if !started {
			return "", "", false
		}
	}
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func normalizeAmountString(value string) string {
	cleaned := CleanAmount(value)
	// Only strip clearly non-numeric trailing notes (e.g. "12.00 (备注)").
	// Arithmetic-looking parentheses such as "1(2+3)" / "12（+3）" are kept so
	// later strict parsing rejects them instead of silently truncating.
	cleaned, _ = splitTrailingNonNumericAnnotation(cleaned)
	return cleaned
}

func formatAmountLike(amount float64, original string) string {
	// Legacy helper retained for float call sites; prefer formatAmountLikeDecimal.
	d, err := ir.ParseDecimal(strconv.FormatFloat(amount, 'f', -1, 64))
	if err != nil {
		precision := 2
		cleaned := normalizeAmountString(original)
		if dot := strings.LastIndex(cleaned, "."); dot >= 0 {
			precision = len(cleaned) - dot - 1
		}
		if precision < 2 {
			precision = 2
		}
		return strconv.FormatFloat(amount, 'f', precision, 64)
	}
	return formatAmountLikeDecimal(d, original)
}

func formatAmountLikeDecimal(amount ir.Decimal, original string) string {
	minScale := uint32(2)
	cleaned := normalizeAmountString(original)
	if dot := strings.LastIndex(cleaned, "."); dot >= 0 {
		frac := cleaned[dot+1:]
		if e := strings.IndexAny(frac, "eE"); e >= 0 {
			frac = frac[:e]
		}
		if uint32(len(frac)) > minScale {
			minScale = uint32(len(frac))
		}
	}
	text0 := amount.Text(0)
	if dot := strings.LastIndex(text0, "."); dot >= 0 {
		if uint32(len(text0)-dot-1) > minScale {
			minScale = uint32(len(text0) - dot - 1)
		}
	}
	return amount.Text(minScale)
}

func formatValue(value, pattern string) string {
	pattern = strings.TrimSpace(pattern)
	pattern = strings.Trim(pattern, `"'`)
	if pattern == "" {
		return value
	}
	if strings.ContainsAny(pattern, "fFeEgG") {
		d, err := ParseAmountDecimal(value, "")
		if err != nil {
			return value
		}
		out, ok := formatDecimalPrintfExact(d, pattern)
		if !ok {
			// Unsupported / lossy float formats: return unchanged rather than
			// silently wrong float64 output.
			return value
		}
		return out
	}
	return fmt.Sprintf(pattern, strings.TrimSpace(value))
}

// formatDecimalPrintfExact formats d with a single %[.N]f verb, preserving
// literal prefix/suffix (e.g. "-%.2f" → "-12.34"). No float64 path.
func formatDecimalPrintfExact(d ir.Decimal, pattern string) (string, bool) {
	if strings.ContainsAny(pattern, "eEgG") {
		return "", false
	}
	percent, ok := indexSinglePrintfVerb(pattern)
	if !ok {
		return "", false
	}
	prefix := pattern[:percent]
	rest := pattern[percent+1:]
	prec := 6 // fmt default for %f
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		n := 0
		i := 0
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			n = n*10 + int(rest[i]-'0')
			i++
		}
		if i == 0 {
			return "", false
		}
		prec = n
		rest = rest[i:]
	}
	if rest == "" || (rest[0] != 'f' && rest[0] != 'F') {
		return "", false
	}
	suffix := rest[1:]
	// Reject remaining % verbs / flags / width that we do not implement exactly.
	if strings.ContainsRune(suffix, '%') || strings.ContainsAny(prefix, "eEgG") {
		return "", false
	}
	frac := fractionalDigits(d)
	if frac > prec {
		return "", false
	}
	return prefix + d.Text(uint32(prec)) + suffix, true
}

func printfFloatPrecision(pattern string) (int, bool) {
	// Supports literal%[.]Nf literal (e.g. "-%.2f") without float64.
	if !strings.ContainsAny(pattern, "fF") || strings.ContainsAny(pattern, "eEgG") {
		return 0, false
	}
	percent, ok := indexSinglePrintfVerb(pattern)
	if !ok {
		return 0, false
	}
	rest := pattern[percent+1:]
	prec := 6
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		n, i := 0, 0
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			n = n*10 + int(rest[i]-'0')
			i++
		}
		if i == 0 {
			return 0, false
		}
		prec = n
		rest = rest[i:]
	}
	if rest == "" || (rest[0] != 'f' && rest[0] != 'F') {
		return 0, false
	}
	return prec, true
}

// indexSinglePrintfVerb finds the sole unescaped % in pattern.
func indexSinglePrintfVerb(pattern string) (int, bool) {
	idx := -1
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '%' {
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '%' {
			i++
			continue
		}
		if idx >= 0 {
			return 0, false
		}
		idx = i
	}
	if idx < 0 {
		return 0, false
	}
	return idx, true
}

func fractionalDigits(d ir.Decimal) int {
	text := d.Text(0)
	if dot := strings.LastIndex(text, "."); dot >= 0 {
		return len(text) - dot - 1
	}
	return 0
}

func evalSimpleArithmetic(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value
	}
	// Only rewrite pure arithmetic / signed numerics; leave account lines alone.
	if !looksLikeArithmetic(trimmed) && !isPlainNumberToken(trimmed) {
		out, err := evalArithmeticInTextStrict(trimmed)
		if err != nil {
			return value
		}
		return out
	}
	out, err := evalArithmeticExpression(trimmed)
	if err != nil {
		// Soft helper: leave expression unchanged on failure (runtime paths use strict).
		return value
	}
	return formatAmountLikeDecimal(out, trimmed)
}

func isPlainNumberToken(value string) bool {
	_, err := ir.ParseDecimal(CleanAmount(value))
	return err == nil && !strings.ContainsAny(value, "*/+") && !strings.Contains(value, "--")
}

// evalArithmeticInTextStrict evaluates embedded arithmetic with standard precedence.
func evalArithmeticInTextStrict(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value, nil
	}
	if looksLikeArithmetic(trimmed) || isPlainNumberToken(trimmed) {
		if containsBinaryArithmetic(trimmed) || strings.ContainsAny(trimmed, "()") {
			out, err := evalArithmeticExpression(trimmed)
			if err != nil {
				return "", err
			}
			return formatAmountLikeDecimal(out, trimmed), nil
		}
		if looksLikeArithmetic(trimmed) && (strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+")) {
			out, err := evalArithmeticExpression(trimmed)
			if err != nil {
				return "", err
			}
			return formatAmountLikeDecimal(out, trimmed), nil
		}
	}
	return rewriteArithmeticRegionsStrict(value)
}

// evalEmbeddedArithmetic is the soft leftmost-compatible wrapper retained for
// legacy call sites; new money paths must use evalArithmeticInTextStrict.
func evalEmbeddedArithmetic(value string) string {
	out, err := evalArithmeticInTextStrict(value)
	if err != nil {
		return value
	}
	return out
}

var simpleArithmeticPattern = regexp.MustCompile(`(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)\s*([*/+-])\s*(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)`)

func trimFraction(value string, minPrecision int) string {
	dot := strings.LastIndex(value, ".")
	if dot < 0 {
		return value
	}
	for len(value)-dot-1 > minPrecision && strings.HasSuffix(value, "0") {
		value = strings.TrimSuffix(value, "0")
	}
	return value
}

func decimalPlaces(value string) int {
	if dot := strings.LastIndex(value, "."); dot >= 0 {
		return len(value) - dot - 1
	}
	return 0
}
