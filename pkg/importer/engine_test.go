package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
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
				Actions: Actions{Type: "send", MinusAccount: "Assets:Alipay"},
			},
		},
		PersonalRules: []Rule{
			{
				When:    `raw[交易对方] ~ "滴露"`,
				Actions: Actions{PlusAccount: "Expenses:Groceries"},
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
		{ref: "wechat@2026.05"},
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
