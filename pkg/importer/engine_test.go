package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"gopkg.in/yaml.v3"
)

func TestImportFileAppliesTemplateAndPersonalRules(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(csvPath, []byte("交易时间,交易对方,商品,收/支,金额,支付方式\n2026-05-21 10:30:00,滴露,洗手液,支出,18.90,余额\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	profile := &Profile{
		ID: "test",
		Template: Template{
			FileFormat:      "csv",
			DateFormat:      "yyyy-MM-dd HH:mm:ss",
			DefaultMinus:    "Assets:FIXME",
			DefaultPlus:     "Expenses:FIXME",
			DefaultCurrency: "CNY",
			Columns: ColumnMapping{
				Date:      "交易时间",
				Amount:    "金额",
				Payee:     "交易对方",
				Narration: "商品",
				Type:      "收/支",
			},
			Metadata: map[string]string{"method": "支付方式"},
		},
		TemplateRules: []Rule{
			{
				When:    `raw[收/支] == "支出"`,
				Actions: Actions{Type: "send", From: TransferSide{Account: "Assets:Alipay"}},
			},
		},
		PersonalRules: []Rule{
			{
				When:    `raw[交易对方] ~ "滴露"`,
				Actions: Actions{To: TransferSide{Account: "Expenses:Groceries"}},
			},
		},
	}

	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(out.Orders))
	}
	order := out.Orders[0]
	if order.Type != ir.TypeSend {
		t.Fatalf("expected send type, got %s", order.Type)
	}
	if order.MinusAccount != "Assets:Alipay" {
		t.Fatalf("unexpected minus account: %s", order.MinusAccount)
	}
	if order.PlusAccount != "Expenses:Groceries" {
		t.Fatalf("unexpected plus account: %s", order.PlusAccount)
	}
	if order.Metadata["method"] != "余额" {
		t.Fatalf("metadata not mapped: %#v", order.Metadata)
	}
}

func TestTemplateRefKinds(t *testing.T) {
	tests := []struct {
		ref       string
		http      bool
		localPath bool
	}{
		{ref: "wechat"},
		{ref: "wechat@2026-04-28"},
		{ref: "https://example.com/wechat.yaml", http: true},
		{ref: "http://example.com/wechat.yaml", http: true},
		{ref: "./wechat.yaml", localPath: true},
		{ref: "../templates/wechat.yaml", localPath: true},
		{ref: "/tmp/wechat.yaml", localPath: true},
		{ref: "~/wechat.yaml", localPath: true},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			if got := IsHTTPURL(tt.ref); got != tt.http {
				t.Fatalf("IsHTTPURL(%q) = %v, want %v", tt.ref, got, tt.http)
			}
			if got := IsLocalPathRef(tt.ref); got != tt.localPath {
				t.Fatalf("IsLocalPathRef(%q) = %v, want %v", tt.ref, got, tt.localPath)
			}
		})
	}
}

func TestRawTimeSuffixCondition(t *testing.T) {
	row := Row{
		Raw: map[string]string{"交易时间": "2026-05-21 12:30:00"},
	}
	ok, err := evalWhen(`raw[交易时间].time >= "11:00" && raw[交易时间].time <= "15:00"`, row, ir.Order{})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected raw[交易时间].time condition to match")
	}
}

func TestGroupedOrConditionCanBeCombinedWithAnd(t *testing.T) {
	row := Row{
		Raw: map[string]string{
			"交易创建时间": "2026-05-21 18:30:00",
			"订单标题":   "牛肉火锅",
			"收/支":    "支出",
		},
	}
	expr := `(<订单标题> ~ "单人餐" || <订单标题> ~ "牛肉火锅") && raw[交易创建时间].time >= "16:00" && raw[交易创建时间].time <= "21:00" && <收/支> == "支出"`
	ok, err := evalWhen(expr, row, ir.Order{})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected grouped OR condition with trailing AND clauses to match")
	}
}

func TestRuleYAMLSupportsAngleFieldReferencesWithoutBreakingArrays(t *testing.T) {
	var profile Profile
	err := yaml.Unmarshal([]byte(`
schema: https://deg.dev/template-profile/v2
id: 测试
template:
  columns:
    date: 交易时间
    amount: 金额
personalRules:
  - id: 元数据
    when: <收/支> == 收入
    actions:
      metadata:
        category: <交易分类>
      amount: <金额>.number
      tags:
        - income
        - service
`), &profile)
	if err != nil {
		t.Fatal(err)
	}
	rule := profile.PersonalRules[0]
	if rule.When != "<收/支> == 收入" {
		t.Fatalf("unexpected when: %q", rule.When)
	}
	if rule.Actions.Metadata["category"] != "<交易分类>" {
		t.Fatalf("unexpected metadata shorthand: %#v", rule.Actions.Metadata)
	}
	if rule.Actions.Amount != "<金额>.number" {
		t.Fatalf("unexpected amount: %q", rule.Actions.Amount)
	}
	if got := strings.Join(rule.Actions.Tags, ","); got != "income,service" {
		t.Fatalf("tags should remain an array, got %#v", rule.Actions.Tags)
	}
	row := Row{Raw: map[string]string{
		"收/支":  "收入",
		"交易分类": "餐饮美食",
		"金额":   "18.90",
	}}
	ok, err := evalWhen(rule.When, row, ir.Order{})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected angle field condition to match")
	}
	if got := renderRuleText(rule.Actions.Metadata["category"], row, ir.Order{}); got != "餐饮美食" {
		t.Fatalf("unexpected rendered metadata: %q", got)
	}
}

func TestRuleExpressionErrorIsReturned(t *testing.T) {
	profile := testProfile()
	profile.TemplateRules = []Rule{
		{
			ID:      "错误规则",
			When:    `raw[收/支] ==`,
			Actions: Actions{Ignore: true},
		},
	}
	_, err := ImportFile(profile, writeTestCSV(t))
	if err == nil {
		t.Fatal("expected invalid rule expression error")
	}
	if !strings.Contains(err.Error(), "错误规则") {
		t.Fatalf("expected rule id in error, got %v", err)
	}
}

func TestParseFileRejectsBillTemplateFormatMismatch(t *testing.T) {
	profile := testProfile()
	profile.ID = "wechat"
	profile.Template.FileFormat = "csv"
	_, err := ParseFile(profile, filepath.Join("..", "..", "example", "wechat", "example-wechat-records.xlsx"))
	if err == nil {
		t.Fatal("expected format mismatch error")
	}
	if !strings.Contains(err.Error(), `fileFormat="csv"`) || !strings.Contains(err.Error(), "xlsx") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTemplateRuleOverrideCanDisableRule(t *testing.T) {
	disabled := false
	profile := testProfile()
	profile.TemplateRules = []Rule{
		{
			ID:      "忽略滴露",
			When:    `raw[交易对方] == "滴露"`,
			Actions: Actions{Ignore: true},
		},
	}
	profile.TemplateRuleOverrides = []Rule{
		{ID: "忽略滴露", Enabled: &disabled},
	}
	out, err := ImportFile(profile, writeTestCSV(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected disabled template rule to keep order, got %d orders", len(out.Orders))
	}
}

func TestSourceHeadersAndSplitAmountColumns(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(csvPath, []byte("2026-05-21,10:30:00,午餐,,18.90\n2026-05-22,09:00:00,工资,100.00,\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := &Profile{
		ID: "headerless",
		Template: Template{
			FileFormat:      "csv",
			DateFormat:      "yyyy-MM-dd HH:mm:ss",
			SourceHeaders:   []string{"date", "time", "payee", "in", "out"},
			DefaultMinus:    "Assets:FIXME",
			DefaultPlus:     "Expenses:FIXME",
			DefaultCurrency: "CNY",
			Columns: ColumnMapping{
				Date:      "date",
				Time:      "time",
				AmountIn:  "in",
				AmountOut: "out",
				Payee:     "payee",
			},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(out.Orders))
	}
	if out.Orders[0].Type != ir.TypeSend {
		t.Fatalf("expected first order send, got %s", out.Orders[0].Type)
	}
	if out.Orders[1].Type != ir.TypeRecv {
		t.Fatalf("expected second order recv, got %s", out.Orders[1].Type)
	}
}

func TestV2ActionsBuildExplicitPostings(t *testing.T) {
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.DefaultMinus = ""
	profile.Template.DefaultPlus = ""
	profile.PersonalRules = []Rule{
		{
			When: `[交易对方] == "滴露"`,
			Actions: Actions{
				To:       TransferSide{Account: "Expenses:Groceries"},
				Amount:   "[金额].number",
				Currency: "CNY",
				Postings: []string{
					`Expenses:Fees [金额].number.! CNY`,
				},
			},
		},
	}
	out, err := ImportFile(profile, writeTestCSV(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(out.Orders))
	}
	postings := out.Orders[0].Postings
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings, got %#v", postings)
	}
	if postings[0].Line != "Expenses:Groceries 18.90 CNY" {
		t.Fatalf("unexpected to posting: %q", postings[0].Line)
	}
	if postings[1].Line != "Expenses:Fees -18.90 CNY" {
		t.Fatalf("unexpected extra posting: %q", postings[1].Line)
	}
	if out.Orders[0].MinusAccount != "" || out.Orders[0].PlusAccount != "" {
		t.Fatalf("v2 should not fill default accounts: minus=%q plus=%q", out.Orders[0].MinusAccount, out.Orders[0].PlusAccount)
	}
}

func TestRuntimeImporterCollectsStaticRuleOpenAccounts(t *testing.T) {
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.DefaultMinus = ""
	profile.Template.DefaultPlus = ""
	profile.PersonalRules = []Rule{
		{
			ID:   "未命中静态账户",
			When: `<交易对方> == 不存在`,
			Actions: Actions{
				From: TransferSide{Account: "Assets:Static"},
				To:   TransferSide{Account: "Expenses:Static"},
				Postings: []string{
					"Income:Static <金额>.number CNY",
					"<动态账户> <金额>.number CNY",
				},
			},
		},
	}
	out, err := ImportFile(profile, writeTestCSV(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, account := range []string{"Assets:Static", "Expenses:Static", "Income:Static"} {
		if !out.OpenAccounts[account] {
			t.Fatalf("expected static account %s to be collected: %#v", account, out.OpenAccounts)
		}
	}
	if out.OpenAccounts["<动态账户>"] {
		t.Fatalf("dynamic account expression should not be collected: %#v", out.OpenAccounts)
	}
}

func TestV2TransferSideOverridesAmountAndCurrency(t *testing.T) {
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.PersonalRules = []Rule{
		{
			Actions: Actions{
				From:     TransferSide{Account: "Assets:Cash", Amount: "[金额].number", Currency: "CNY"},
				To:       TransferSide{Account: "Assets:Bank", Amount: "[金额].number * 0.99", Currency: "CNY"},
				Postings: []string{`Expenses:Commission [金额].number.+ * 0.01 CNY`},
			},
		},
	}
	out, err := ImportFile(profile, writeTestCSV(t))
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, posting := range out.Orders[0].Postings {
		got = append(got, posting.Line)
	}
	want := []string{
		"Assets:Bank 18.711 CNY",
		"Assets:Cash -18.90 CNY",
		"Expenses:Commission 0.189 CNY",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("postings mismatch\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestSkipInvalidRows(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(csvPath, []byte("date,payee,amount\n2026-05-21,午餐,18.90\n合计,,18.90\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := &Profile{
		ID: "skip-summary",
		Template: Template{
			FileFormat:      "csv",
			SkipInvalidRows: true,
			DefaultMinus:    "Assets:FIXME",
			DefaultPlus:     "Expenses:FIXME",
			DefaultCurrency: "CNY",
			Columns: ColumnMapping{
				Date:   "date",
				Amount: "amount",
				Payee:  "payee",
			},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected 1 order after skipping summary row, got %d", len(out.Orders))
	}
}

func testProfile() *Profile {
	return &Profile{
		ID: "test",
		Template: Template{
			FileFormat:      "csv",
			DateFormat:      "yyyy-MM-dd HH:mm:ss",
			DefaultMinus:    "Assets:FIXME",
			DefaultPlus:     "Expenses:FIXME",
			DefaultCurrency: "CNY",
			Columns: ColumnMapping{
				Date:      "交易时间",
				Amount:    "金额",
				Payee:     "交易对方",
				Narration: "商品",
				Type:      "收/支",
			},
		},
	}
}

func writeTestCSV(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(csvPath, []byte("交易时间,交易对方,商品,收/支,金额,支付方式\n2026-05-21 10:30:00,滴露,洗手液,支出,18.90,余额\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return csvPath
}
