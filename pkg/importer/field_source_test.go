package importer

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

const fieldSourceBill = "" +
	"交易时间,商户,商品,收/支,金额(元),支付方式,交易单号\n" +
	"2026-05-21 10:30:00,一卡通充值,地铁,支出,18.90,零钱,42\n" +
	"2026-05-22 09:00:00,公司,工资,收入,100.00,银行卡,7\n"

func TestFieldSourcesMapColumnsRulesAndEngineFallback(t *testing.T) {
	profile := slotProfile()
	profile.PersonalRules = []Rule{{
		ID:      "一卡通",
		When:    `payee ~ "一卡通"`,
		Actions: Actions{From: TransferSide{Account: "Assets:Current:零钱"}},
	}, {
		ID:      "去掉订单号",
		When:    `metadata.orderId == "42"`,
		Actions: Actions{MetadataDrop: []string{"orderId"}},
	}}

	out := importFieldSourceBill(t, profile)
	expense := out.Orders[0]
	requireSource(t, expense, "date", ir.FieldOriginTemplate, "", "交易时间")
	requireSource(t, expense, "payee", ir.FieldOriginTemplate, "", "商户")
	requireSource(t, expense, "narration", ir.FieldOriginTemplate, "", "商品")
	requireSource(t, expense, "amount", ir.FieldOriginTemplate, "", "金额(元)")
	requireSource(t, expense, "currency", ir.FieldOriginTemplate, "")
	requireSource(t, expense, "metadata.method", ir.FieldOriginTemplate, "", "支付方式")
	requireSource(t, expense, "metadata.type", ir.FieldOriginTemplate, "", "收/支")
	if _, ok := findSource(expense, "metadata.orderId"); ok {
		t.Fatalf("dropped metadata still has a source: %#v", expense.Sources)
	}
	requireSource(t, expense, "from", ir.FieldOriginRule, "一卡通")
	requireSource(t, expense, "to", ir.FieldOriginEngine, "")
	if expense.MinusAccount != "Assets:Current:零钱" || expense.PlusAccount != "Expenses:FIXME" {
		t.Fatalf("accounts: %s / %s", expense.MinusAccount, expense.PlusAccount)
	}
	postings := joinPostings(expense)
	if !strings.Contains(postings, "Assets:Current:零钱") || !strings.Contains(postings, "Expenses:FIXME") {
		t.Fatalf("postings: %s", postings)
	}

	income := out.Orders[1]
	requireSource(t, income, "from", ir.FieldOriginEngine, "")
	requireSource(t, income, "to", ir.FieldOriginEngine, "")
	requireSource(t, income, "metadata.orderId", ir.FieldOriginTemplate, "", "交易单号")
	text := joinPostings(income)
	if !strings.Contains(text, "Income:FIXME") || !strings.Contains(text, "Assets:FIXME") {
		t.Fatalf("income fallback: %s", text)
	}
	if income.MinusAccount != "Income:FIXME" || income.PlusAccount != "Assets:FIXME" {
		t.Fatalf("income accounts: %s / %s", income.MinusAccount, income.PlusAccount)
	}
}

func TestFieldSourcesExplicitFIXMEIsARule(t *testing.T) {
	profile := slotProfile()
	profile.Template.Slots.Tags = "<支付方式>"
	profile.PersonalRules = []Rule{{
		ID:   "显式支出",
		When: `metadata.type == "支出"`,
		Actions: Actions{
			Payee:     `"固定商户"`,
			Narration: `<商品>-<商户>`,
			To:        TransferSide{Account: "Expenses:FIXME"},
			Metadata:  map[string]string{"shop": "<商户>"},
			Tag:       "<商品>",
			Link:      `"ticket"`,
			Note:      "<支付方式>",
		},
	}}

	expense := importFieldSourceBill(t, profile).Orders[0]
	if expense.Peer != "固定商户" || expense.Item != "地铁-一卡通充值" {
		t.Fatalf("rewritten slots: payee=%q narration=%q", expense.Peer, expense.Item)
	}
	requireSource(t, expense, "payee", ir.FieldOriginRule, "显式支出")
	requireSource(t, expense, "narration", ir.FieldOriginRule, "显式支出", "商品", "商户")
	requireSource(t, expense, "to", ir.FieldOriginRule, "显式支出")
	requireSource(t, expense, "from", ir.FieldOriginEngine, "")
	requireSource(t, expense, "metadata.shop", ir.FieldOriginRule, "显式支出", "商户")
	requireSource(t, expense, "note", ir.FieldOriginRule, "显式支出", "支付方式")
	if expense.PlusAccount != "Expenses:FIXME" || expense.MinusAccount != "Assets:FIXME" {
		t.Fatalf("accounts: %s / %s", expense.MinusAccount, expense.PlusAccount)
	}
	tags := sourcesFor(expense, "tags")
	if len(tags) != 2 || tags[0].Origin != ir.FieldOriginTemplate || tags[1].Origin != ir.FieldOriginRule || tags[1].RuleID != "显式支出" {
		t.Fatalf("tag sources: %#v", tags)
	}
	if !slices.Equal(tags[0].Columns, []string{"支付方式"}) || !slices.Equal(tags[1].Columns, []string{"商品"}) {
		t.Fatalf("tag columns: %#v", tags)
	}
	links := sourcesFor(expense, "links")
	if len(links) != 1 || links[0].Origin != ir.FieldOriginRule || len(links[0].Columns) != 0 {
		t.Fatalf("link sources: %#v", links)
	}
	if strings.Join(expense.MetadataKeys, ",") != "method,type,orderId,shop" {
		t.Fatalf("metadata order: %#v", expense.MetadataKeys)
	}
}

func TestFieldSourcesLaterRuleReplacesAmount(t *testing.T) {
	profile := slotProfile()
	profile.PersonalRules = []Rule{{
		ID:   "改金额",
		When: `metadata.orderId == "42"`,
		Actions: Actions{
			Amount:   `<交易单号>.number`,
			Currency: `"USD"`,
			From:     TransferSide{Account: "Assets:Cash"},
		},
	}, {
		ID:      "改账户",
		When:    `metadata.orderId == "42"`,
		Actions: Actions{From: TransferSide{Account: "Assets:Wallet"}},
	}}

	expense := importFieldSourceBill(t, profile).Orders[0]
	requireSource(t, expense, "amount", ir.FieldOriginRule, "改金额", "交易单号")
	requireSource(t, expense, "currency", ir.FieldOriginRule, "改金额")
	requireSource(t, expense, "from", ir.FieldOriginRule, "改账户")
	requireSource(t, expense, "to", ir.FieldOriginEngine, "")
	if expense.MinusAccount != "Assets:Wallet" {
		t.Fatalf("from account: %s", expense.MinusAccount)
	}
	postings := joinPostings(expense)
	if !strings.Contains(postings, "USD") || !strings.Contains(postings, "42") {
		t.Fatalf("postings: %s", postings)
	}
}

func TestLegacyTemplateRulesLeaveSourcesEmpty(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,支付方式\n2026-05-21 10:30:00,滴露,洗手液,支出,18.90,余额\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o644); err != nil {
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
		},
		TemplateRules: []Rule{{
			When:    `raw[收/支] == "支出"`,
			Actions: Actions{Type: "send", From: TransferSide{Account: "Assets:Alipay"}},
		}},
		PersonalRules: []Rule{{
			When:    `raw[交易对方] ~ "滴露"`,
			Actions: Actions{To: TransferSide{Account: "Expenses:Groceries"}},
		}},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	order := out.Orders[0]
	if order.Sources != nil {
		t.Fatalf("legacy sources: %#v", order.Sources)
	}
	if order.MinusAccount != "Assets:Alipay" || order.PlusAccount != "Expenses:Groceries" {
		t.Fatalf("legacy accounts: %s / %s", order.MinusAccount, order.PlusAccount)
	}
}

func TestExpressionColumns(t *testing.T) {
	cases := []struct {
		expr string
		want []string
	}{
		{expr: `<金额(元)>.number`, want: []string{"金额(元)"}},
		{expr: `<商品>-<商户>`, want: []string{"商品", "商户"}},
		{expr: `<商品> <商品>`, want: []string{"商品"}},
		{expr: `"固定<商户>"`},
		{expr: `[payee]`},
		{expr: `CNY`},
	}
	for _, tc := range cases {
		got := expressionColumns(tc.expr)
		if !slices.Equal(got, tc.want) {
			t.Fatalf("%q: got %#v want %#v", tc.expr, got, tc.want)
		}
	}
}

func importFieldSourceBill(t *testing.T, profile *Profile) *ir.IR {
	t.Helper()
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(csvPath, []byte(fieldSourceBill), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 2 {
		t.Fatalf("orders: %d", len(out.Orders))
	}
	return out
}

func requireSource(t *testing.T, order ir.Order, slot string, origin ir.FieldOrigin, ruleID string, columns ...string) {
	t.Helper()
	src, ok := findSource(order, slot)
	if !ok {
		t.Fatalf("missing source %s in %#v", slot, order.Sources)
	}
	if src.Origin != origin || src.RuleID != ruleID {
		t.Fatalf("source %s: origin %q rule %q, want origin %q rule %q", slot, src.Origin, src.RuleID, origin, ruleID)
	}
	got := src.Columns
	if got == nil {
		got = []string{}
	}
	want := columns
	if want == nil {
		want = []string{}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("source %s columns: got %#v want %#v", slot, src.Columns, columns)
	}
}

func findSource(order ir.Order, slot string) (ir.FieldSource, bool) {
	var found ir.FieldSource
	ok := false
	for _, src := range order.Sources {
		if src.Slot == slot {
			found = src
			ok = true
		}
	}
	return found, ok
}

func sourcesFor(order ir.Order, slot string) []ir.FieldSource {
	var out []ir.FieldSource
	for _, src := range order.Sources {
		if src.Slot == slot {
			out = append(out, src)
		}
	}
	return out
}
