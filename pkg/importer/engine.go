package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
	rows, err := ParseFile(profile, filename)
	if err != nil {
		return nil, err
	}
	orders := ir.New()
	if profile.IsV2() {
		orders.OpenAccounts = collectV2Accounts(profile)
	}
	for _, row := range rows {
		order, ignore, err := rowToOrder(profile, row)
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

func collectV2Accounts(profile *Profile) map[string]bool {
	accounts := map[string]bool{}
	for _, rule := range profile.Rules() {
		addRuleAccount(accounts, rule.Actions.From.Account)
		addRuleAccount(accounts, rule.Actions.To.Account)
		for _, line := range rule.Actions.Postings {
			addRuleAccount(accounts, strings.Fields(line)...)
		}
	}
	return accounts
}

func addRuleAccount(accounts map[string]bool, values ...string) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if strings.Contains(value, ":") && !strings.Contains(value, "[") {
			accounts[value] = true
			return
		}
	}
}

func ParseFile(profile *Profile, filename string) ([]Row, error) {
	if err := validateBillMatchesTemplate(profile, filename); err != nil {
		return nil, err
	}
	format := templateFileFormat(profile)
	switch format {
	case "csv":
		return parseCSV(profile, filename)
	case "xls":
		return parseXLS(profile, filename)
	case "xlsx":
		return parseXLSX(profile, filename)
	default:
		return nil, fmt.Errorf("unsupported template fileFormat %q", profile.Template.FileFormat)
	}
}

func templateFileFormat(profile *Profile) string {
	return normalizeFileFormat(profile.Template.FileFormat, "csv")
}

func billFileFormat(filename string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	return normalizeFileFormat(ext, "")
}

func normalizeFileFormat(format, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "txt", "text":
		return "csv"
	case "xls":
		return "xls"
	case "csv", "xlsx":
		return strings.ToLower(strings.TrimSpace(format))
	default:
		return fallback
	}
}

func validateBillMatchesTemplate(profile *Profile, filename string) error {
	templateFmt := templateFileFormat(profile)
	billFmt := billFileFormat(filename)
	if billFmt == "" {
		return fmt.Errorf("无法识别账单文件格式 %q，请使用 csv 或 xlsx", filepath.Ext(filename))
	}
	if templateFmt == billFmt {
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
		templateFmt,
		templateFmt,
		billFmt,
	)
}

func parseCSV(profile *Profile, filename string) ([]Row, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var r io.Reader = file
	if strings.EqualFold(profile.Template.Encoding, "gbk") || strings.EqualFold(profile.Template.Encoding, "gb18030") {
		r = transform.NewReader(file, simplifiedchinese.GB18030.NewDecoder())
	}
	if profile.Template.StripTabs {
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r = strings.NewReader(strings.ReplaceAll(string(b), "\t", ""))
	}

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	if delimiter := normalizeDelimiter(profile.Template.Delimiter); delimiter != 0 {
		reader.Comma = delimiter
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return recordsToRows(profile, records)
}

func parseXLSX(profile *Profile, filename string) ([]Row, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("xlsx has no sheets")
	}
	records, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	return recordsToRows(profile, records)
}

func parseXLS(profile *Profile, filename string) ([]Row, error) {
	wb, err := xls.Open(filename, "utf-8")
	if err != nil {
		return parseCSV(profile, filename)
	}
	if wb.NumSheets() == 0 {
		return nil, fmt.Errorf("xls has no sheets")
	}
	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("xls has no first sheet")
	}
	records := make([][]string, 0, int(sheet.MaxRow)+1)
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			records = append(records, nil)
			continue
		}
		record := make([]string, 0, row.LastCol())
		for col := 0; col < row.LastCol(); col++ {
			record = append(record, safeXLSCell(row, col))
		}
		records = append(records, record)
	}
	return recordsToRows(profile, records)
}

func safeXLSCell(row *xls.Row, col int) (value string) {
	defer func() {
		if recover() != nil {
			value = ""
		}
	}()
	return row.Col(col)
}

func recordsToRows(profile *Profile, records [][]string) ([]Row, error) {
	skip := profile.Template.SkipLeadingRows
	if skip < 0 {
		skip = 0
	}
	if len(records) <= skip {
		return nil, fmt.Errorf("no rows after skipLeadingRows=%d", skip)
	}
	headers := normalizeCells(profile.Template.SourceHeaders)
	start := skip
	if len(headers) == 0 {
		headers = normalizeCells(records[skip])
		start = skip + 1
	} else if sameCells(headers, normalizeCells(records[skip])) {
		start = skip + 1
	}
	if err := validateHeaders(profile, headers); err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(records)-start)
	for _, record := range records[start:] {
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
		if strings.TrimSpace(row.Amount) == "" {
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

func rowToOrder(profile *Profile, row Row) (ir.Order, bool, error) {
	amount, err := parseAmount(row.Amount, profile.Template.AmountPrefix)
	if err != nil {
		return ir.Order{}, false, fmt.Errorf("parse amount %q failed for date=%q payee=%q: %w", row.Amount, row.Date, row.Payee, err)
	}
	txType := inferType(row.Type, amount)
	if amount < 0 {
		amount = -amount
	}
	payTime, err := parseDate(row.Date, profile.Template.DateFormat)
	if err != nil {
		return ir.Order{}, false, err
	}
	order := ir.Order{
		OrderType:    ir.OrderTypeNormal,
		Peer:         row.Payee,
		Item:         row.Narration,
		Money:        amount,
		PayTime:      payTime,
		Type:         txType,
		TypeOriginal: row.Type,
		Currency:     profile.Template.DefaultCurrency,
		Metadata:     row.Metadata,
	}
	if !profile.IsV2() {
		order.MinusAccount = profile.Template.DefaultMinus
		order.PlusAccount = profile.Template.DefaultPlus
	}
	if row.Currency != "" {
		order.Currency = row.Currency
	}
	if order.Metadata == nil {
		order.Metadata = map[string]string{}
	}

	ignore := false
	mergedV2Actions := Actions{}
	for _, rule := range profile.Rules() {
		matches, err := ruleMatches(rule, row, order)
		if err != nil {
			return ir.Order{}, false, err
		}
		if !matches {
			continue
		}
		if profile.IsV2() {
			applyV2ScalarActions(&order, row, rule.Actions, &ignore)
			mergeV2Actions(&mergedV2Actions, rule.Actions)
			continue
		}
		if err := applyActions(&order, row, rule.Actions, &ignore, profile.IsV2()); err != nil {
			return ir.Order{}, false, err
		}
	}
	if profile.IsV2() {
		renderV2Postings(&order, row, mergedV2Actions)
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

func applyActions(order *ir.Order, row Row, actions Actions, ignore *bool, v2 bool) error {
	if actions.Ignore {
		*ignore = true
	}
	if actions.Type != "" {
		order.Type = inferType(actions.Type, order.Money)
		order.TypeOriginal = actions.Type
	}
	if actions.Payee != "" {
		order.Peer = resolveValue(actions.Payee, row)
	}
	if actions.Narration != "" {
		order.Item = resolveValue(actions.Narration, row)
	}
	if actions.Amount != "" {
		if amount, err := parseAmount(resolveValue(actions.Amount, row), ""); err == nil {
			if amount < 0 {
				amount = -amount
			}
			order.Money = amount
		}
	}
	if actions.Currency != "" {
		order.Currency = resolveValue(actions.Currency, row)
	}
	if !actions.To.IsZero() {
		if v2 {
			if posting, ok := renderTransferPosting(actions.To, actions.Amount, actions.Currency, "+", row, *order); ok {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		} else {
			order.PlusAccount = resolveValue(actions.To.Account, row)
		}
	}
	if !actions.From.IsZero() {
		if v2 {
			if posting, ok := renderTransferPosting(actions.From, actions.Amount, actions.Currency, "-", row, *order); ok {
				order.Postings = append(order.Postings, ir.Posting{Line: posting})
			}
		} else {
			order.MinusAccount = resolveValue(actions.From.Account, row)
		}
	}
	if v2 {
		for _, line := range actions.Postings {
			rendered := strings.TrimSpace(renderRuleText(line, row, *order))
			if rendered != "" {
				order.Postings = append(order.Postings, ir.Posting{Line: rendered})
			}
		}
	}
	if actions.Tag != "" {
		order.Tags = append(order.Tags, splitList(actions.Tag)...)
	}
	order.Tags = append(order.Tags, actions.Tags...)
	if actions.Metadata != nil {
		if order.Metadata == nil {
			order.Metadata = map[string]string{}
		}
		for key, value := range actions.Metadata {
			if v2 {
				order.Metadata[key] = renderRuleText(value, row, *order)
			} else {
				order.Metadata[key] = resolveValue(value, row)
			}
		}
	}
	if actions.Commission != "" {
		if commission, err := parseAmount(resolveValue(actions.Commission, row), ""); err == nil {
			order.Commission = commission
		}
	}
	if actions.CommissionAccount != "" {
		if order.ExtraAccounts == nil {
			order.ExtraAccounts = map[ir.Account]string{}
		}
		order.ExtraAccounts[ir.CommissionAccount] = resolveValue(actions.CommissionAccount, row)
	}
	if actions.PnlAccount != "" {
		if order.ExtraAccounts == nil {
			order.ExtraAccounts = map[ir.Account]string{}
		}
		order.ExtraAccounts[ir.PnlAccount] = resolveValue(actions.PnlAccount, row)
	}
	return nil
}

func applyV2ScalarActions(order *ir.Order, row Row, actions Actions, ignore *bool) {
	if actions.Ignore {
		*ignore = true
	}
	if actions.Payee != "" {
		order.Peer = renderRuleText(actions.Payee, row, *order)
	}
	if actions.Narration != "" {
		order.Item = renderRuleText(actions.Narration, row, *order)
	}
	if actions.Tag != "" {
		order.Tags = append(order.Tags, splitList(actions.Tag)...)
	}
	order.Tags = append(order.Tags, actions.Tags...)
	if actions.Metadata != nil {
		if order.Metadata == nil {
			order.Metadata = map[string]string{}
		}
		for key, value := range actions.Metadata {
			order.Metadata[key] = renderRuleText(value, row, *order)
		}
	}
}

func mergeV2Actions(base *Actions, next Actions) {
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
	base.Postings = append(base.Postings, next.Postings...)
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

func renderV2Postings(order *ir.Order, row Row, actions Actions) {
	if !actions.To.IsZero() {
		if posting, ok := renderTransferPosting(actions.To, actions.Amount, actions.Currency, "+", row, *order); ok {
			order.Postings = append(order.Postings, ir.Posting{Line: posting})
		}
	}
	if !actions.From.IsZero() {
		if posting, ok := renderTransferPosting(actions.From, actions.Amount, actions.Currency, "-", row, *order); ok {
			order.Postings = append(order.Postings, ir.Posting{Line: posting})
		}
	}
	for _, line := range actions.Postings {
		rendered := strings.TrimSpace(renderPostingText(line, row, *order))
		if rendered != "" {
			order.Postings = append(order.Postings, ir.Posting{Line: rendered})
		}
	}
}

func renderTransferPosting(side TransferSide, defaultAmount, defaultCurrency, direction string, row Row, order ir.Order) (string, bool) {
	account := strings.TrimSpace(renderRuleText(side.Account, row, order))
	if account == "" {
		return "", false
	}
	amount := firstNonEmptyString(side.Amount, defaultAmount)
	currency := firstNonEmptyString(side.Currency, defaultCurrency)
	if amount == "" {
		amount = "[amount].number"
	}
	amount = forceAmountDirection(amount, direction)
	parts := []string{account, renderPostingText(amount, row, order)}
	if currency != "" {
		parts = append(parts, renderRuleText(currency, row, order))
	}
	return strings.Join(nonEmptyStrings(parts), " "), true
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

func fieldValue(field string, row Row, order ir.Order) string {
	field = strings.TrimSpace(field)
	if base, suffix, ok := strings.Cut(field, "."); ok && (suffix == "time" || suffix == "date" || suffix == "timestamp") {
		value := fieldValue(base, row, order)
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
	switch strings.ToLower(field) {
	case "date":
		return row.Date
	case "amount":
		return row.Amount
	case "currency":
		return row.Currency
	case "payee", "peer":
		return row.Payee
	case "narration", "item":
		return row.Narration
	case "type":
		return row.Type
	case "minusaccount", "minus_account":
		return order.MinusAccount
	case "plusaccount", "plus_account":
		return order.PlusAccount
	default:
		if strings.HasPrefix(field, "metadata.") {
			return row.Metadata[strings.TrimPrefix(field, "metadata.")]
		}
		if strings.HasPrefix(field, "raw.") {
			return row.Raw[strings.TrimPrefix(field, "raw.")]
		}
		return row.Raw[field]
	}
}

func parseAmount(value, prefix string) (float64, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, strings.TrimSpace(prefix))
	replacer := strings.NewReplacer(",", "", "¥", "", "￥", "", "$", "", "CNY", "", "RMB", "")
	value = strings.TrimSpace(replacer.Replace(value))
	if base, _, ok := strings.Cut(value, "("); ok {
		value = strings.TrimSpace(base)
	}
	if base, _, ok := strings.Cut(value, "（"); ok {
		value = strings.TrimSpace(base)
	}
	return strconv.ParseFloat(value, 64)
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

func normalizeDelimiter(value string) rune {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "comma", ",":
		return ','
	case "\\t", "tab", "tsv":
		return '\t'
	case "semicolon", ";":
		return ';'
	default:
		return []rune(value)[0]
	}
}

func normalizeCells(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		value = strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
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

var columnExprPattern = regexp.MustCompile(`\[([^\]]+)\]((?:\.extract\((?:r)?"[^"]*"\)|\.extract\((?:r)?'[^']*'\)|\.[A-Za-z0-9_+\-!]+)*)`)

func renderRuleText(value string, row Row, order ir.Order) string {
	return columnExprPattern.ReplaceAllStringFunc(value, func(match string) string {
		return evalColumnString(match, row, order)
	})
}

func renderPostingText(value string, row Row, order ir.Order) string {
	rendered := renderRuleText(value, row, order)
	return evalSimpleArithmetic(rendered)
}

func evalColumnString(expr string, row Row, order ir.Order) string {
	if !strings.HasPrefix(expr, "[") {
		return expr
	}
	end := strings.Index(expr, "]")
	if end < 0 {
		return expr
	}
	field := expr[1:end]
	value := row.Raw[field]
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
	if strings.HasPrefix(rest, "extract(") {
		end := closingMethodParen(rest)
		if end < 0 {
			return rest, "", ""
		}
		return "extract", strings.TrimPrefix(strings.TrimSuffix(rest[len("extract("):end], `"`), `"`), rest[end+1:]
	}
	i := 0
	for i < len(rest) && (isIdentByte(rest[i]) || rest[i] == '+' || rest[i] == '-' || rest[i] == '!') {
		i++
	}
	return rest[:i], "", rest[i:]
}

func closingMethodParen(value string) int {
	var quote byte
	for i := len("extract("); i < len(value); i++ {
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
		n, err := parseAmount(value, "")
		if err != nil {
			return value
		}
		return formatAmountLike(math.Abs(n), value)
	case "-":
		n, err := parseAmount(value, "")
		if err != nil {
			return value
		}
		return formatAmountLike(-math.Abs(n), value)
	case "!":
		n, err := parseAmount(value, "")
		if err != nil {
			return value
		}
		return formatAmountLike(-n, value)
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
	}
	return value
}

func normalizeAmountString(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer(",", "", "¥", "", "￥", "", "$", "", "CNY", "", "RMB", "")
	value = strings.TrimSpace(replacer.Replace(value))
	if base, _, ok := strings.Cut(value, "("); ok {
		value = strings.TrimSpace(base)
	}
	if base, _, ok := strings.Cut(value, "（"); ok {
		value = strings.TrimSpace(base)
	}
	return value
}

func formatAmountLike(amount float64, original string) string {
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

var simpleArithmeticPattern = regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s*([*/+-])\s*(-?\d+(?:\.\d+)?)`)

func evalSimpleArithmetic(value string) string {
	for {
		loc := simpleArithmeticPattern.FindStringSubmatchIndex(value)
		if loc == nil {
			return value
		}
		match := value[loc[0]:loc[1]]
		parts := simpleArithmeticPattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return value
		}
		left, err1 := strconv.ParseFloat(parts[1], 64)
		right, err2 := strconv.ParseFloat(parts[3], 64)
		if err1 != nil || err2 != nil {
			return value
		}
		var out float64
		switch parts[2] {
		case "*":
			out = left * right
		case "/":
			if right == 0 {
				return value
			}
			out = left / right
		case "+":
			out = left + right
		case "-":
			out = left - right
		}
		precision := max(decimalPlaces(parts[1]), decimalPlaces(parts[3]))
		if parts[2] == "*" {
			precision = decimalPlaces(parts[1]) + decimalPlaces(parts[3])
		}
		if precision < 2 {
			precision = 2
		}
		value = value[:loc[0]] + trimFraction(strconv.FormatFloat(out, 'f', precision, 64), 2) + value[loc[1]:]
	}
}

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
