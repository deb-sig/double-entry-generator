package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
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
	for _, row := range rows {
		order, ignore, err := rowToOrder(profile, row)
		if err != nil {
			return nil, err
		}
		if ignore {
			continue
		}
		orders.Orders = append(orders.Orders, order)
	}
	return orders, nil
}

func ParseFile(profile *Profile, filename string) ([]Row, error) {
	if err := validateBillMatchesTemplate(profile, filename); err != nil {
		return nil, err
	}
	format := templateFileFormat(profile)
	switch format {
	case "csv":
		return parseCSV(profile, filename)
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
		return "xlsx"
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

func recordsToRows(profile *Profile, records [][]string) ([]Row, error) {
	skip := profile.Template.SkipLeadingRows
	if skip < 0 {
		skip = 0
	}
	if len(records) <= skip {
		return nil, fmt.Errorf("no header row after skipLeadingRows=%d", skip)
	}
	headers := normalizeCells(records[skip])
	rows := make([]Row, 0, len(records)-skip-1)
	for _, record := range records[skip+1:] {
		record = normalizeCells(record)
		if emptyRecord(record) {
			continue
		}
		raw := make(map[string]string, len(headers))
		for i, h := range headers {
			raw[h] = cell(record, i)
		}
		metadata := make(map[string]string, len(profile.Template.Metadata))
		for key, source := range profile.Template.Metadata {
			metadata[key] = raw[source]
		}
		row := Row{
			Date:      raw[profile.Template.Columns.Date],
			Amount:    raw[profile.Template.Columns.Amount],
			Currency:  raw[profile.Template.Columns.Currency],
			Payee:     raw[profile.Template.Columns.Payee],
			Narration: raw[profile.Template.Columns.Narration],
			Type:      raw[profile.Template.Columns.Type],
			Metadata:  metadata,
			Raw:       raw,
		}
		rows = append(rows, row)
	}
	return rows, nil
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
		MinusAccount: profile.Template.DefaultMinus,
		PlusAccount:  profile.Template.DefaultPlus,
		Metadata:     row.Metadata,
	}
	if row.Currency != "" {
		order.Currency = row.Currency
	}
	if order.Metadata == nil {
		order.Metadata = map[string]string{}
	}

	ignore := false
	for _, rule := range profile.Rules() {
		matches, err := ruleMatches(rule, row, order)
		if err != nil {
			return ir.Order{}, false, err
		}
		if !matches {
			continue
		}
		applyActions(&order, row, rule.Actions, &ignore)
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

func applyActions(order *ir.Order, row Row, actions Actions, ignore *bool) {
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
	if actions.From != "" {
		order.MinusAccount = resolveValue(actions.From, row)
	}
	if actions.To != "" {
		order.PlusAccount = resolveValue(actions.To, row)
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
			order.Metadata[key] = resolveValue(value, row)
		}
	}
	if actions.Commission != "" {
		if commission, err := parseAmount(resolveValue(actions.Commission, row), ""); err == nil {
			order.Commission = commission
		}
	}
	if actions.PnlAccount != "" {
		if order.ExtraAccounts == nil {
			order.ExtraAccounts = map[ir.Account]string{}
		}
		order.ExtraAccounts[ir.PnlAccount] = resolveValue(actions.PnlAccount, row)
	}
}

func fieldValue(field string, row Row, order ir.Order) string {
	field = strings.TrimSpace(field)
	if base, suffix, ok := strings.Cut(field, "."); ok && (suffix == "time" || suffix == "date") {
		value := fieldValue(base, row, order)
		if base == "date" || base == "交易时间" || value == "" {
			if suffix == "time" {
				return order.PayTime.Format("15:04")
			}
			return order.PayTime.Format("2006-01-02")
		}
		if t, err := parseDate(value, ""); err == nil {
			if suffix == "time" {
				return t.Format("15:04")
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
		out[i] = strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
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
