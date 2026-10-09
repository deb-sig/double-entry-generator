package importer

import (
	"strings"
	"testing"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// Statement-level values (account number, card alias, statement month) sit
// above the table. capture lifts them into <file.x> for slots and rules.
func TestCaptureFileValues(t *testing.T) {
	out := importCSV(t, `
schema: https://deg.dev/template-profile/v2
shape:
  - capture: { pattern: '账[\s\p{Zs}]*号[：:][\s\p{Zs}]*(?P<account>\S+)', scanRows: 5 }
  - capture: { pattern: '(?P<year>\d{4})年(?P<month>\d{2})月' }
  - locateHeader: { anchor: [日期, 金额] }
template:
  fileFormat: csv
  dateFormat: yyyy/MM/dd
  defaultCurrency: CNY
  slots:
    date: <file.year>/<日期>
    amount: <金额>.number
    metadata: { account: <file.account> }
personalRules:
  - id: 按账号
    when: <file.account> == "6222001"
    actions: { to: Expenses:Known }
`, "招商银行信用卡对账单 2024年01月\n账　号： 6222001\n日期,金额\n12/30,-5\n01/02,-7\n")
	if len(out.Orders) != 2 {
		t.Fatalf("orders = %d", len(out.Orders))
	}
	if out.Orders[0].Metadata["account"] != "6222001" || out.Orders[0].PayTime.Year() != 2024 {
		t.Fatalf("order = %+v", out.Orders[0])
	}
	if !strings.Contains(joinPostings(out.Orders[0]), "Expenses:Known") {
		t.Fatalf("rule on <file.x> did not match:\n%s", joinPostings(out.Orders[0]))
	}

	profile := loadProfileYAML(t, `
shape:
  - capture: { pattern: '卡别名: (?P<card>\S+)', scanRows: 2 }
template: { fileFormat: csv, slots: { date: <d>, amount: <a> } }
`)
	if _, err := ParseBytes(profile, "b.csv", []byte("说明\n其他\nd,a\n")); err == nil || !strings.Contains(err.Error(), "card") {
		t.Fatalf("missing capture should name the group, got %v", err)
	}
}

// Statements without a year: capture the statement year/month, then pick
// last year for months after the statement month, all with vars.
func TestCaptureComposesYearInference(t *testing.T) {
	out := importCSV(t, `
schema: https://deg.dev/template-profile/v2
shape:
  - capture: { pattern: '(?P<year>\d{4})年(?P<month>\d{2})月' }
  - locateHeader: { anchor: [交易日, 金额] }
template:
  fileFormat: csv
  defaultCurrency: CNY
  vars:
    - vars: { prev: '<file.year> - 1' }
    - vars: { year: <file.year> }
    - when: '<交易日>.extract("^(\d+)") > <file.month>'
      vars: { year: '<var.prev>.format("%.0f")' }
  slots:
    date: <var.year>/<交易日>
    amount: <金额>.number
`, "账单 2024年01月\n交易日,金额\n12/30,-5\n01/02,-7\n")
	if y := out.Orders[0].PayTime.Year(); y != 2023 {
		t.Fatalf("12/30 on a January statement should be 2023, got %d", y)
	}
	if y := out.Orders[1].PayTime.Year(); y != 2024 {
		t.Fatalf("01/02 should be 2024, got %d", y)
	}
}

func TestFailIfStopsWithRowAndMessage(t *testing.T) {
	profile := loadProfileYAML(t, `
schema: https://deg.dev/template-profile/v2
shape:
  - failIf: '<类型> != "消费" && <类型> != "还款"'
    message: 出现了模板不认识的交易类型，账单格式可能变了
template: { fileFormat: csv, slots: { date: <d>, amount: <a>.number } }
`)
	_, err := ParseBytes(profile, "b.csv", []byte("d,类型,a\n2026-01-01,消费,1\n2026-01-02,分期,2\n"))
	if err == nil || !strings.Contains(err.Error(), "账单格式可能变了") || !strings.Contains(err.Error(), "data row 2") || !strings.Contains(err.Error(), "分期") {
		t.Fatalf("err = %v", err)
	}
}

func TestTimezonePinsWallClock(t *testing.T) {
	tpl := `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  defaultCurrency: CNY
  timezone: Asia/Shanghai
  slots: { date: <d>, amount: <a>.number }
personalRules:
  - id: 按时间戳
    when: <d>.timestamp == 1704038400
    actions: { to: Expenses:Matched }
`
	out := importCSV(t, tpl, "d,a\n2024-01-01 00:00:00,-1\n")
	o := out.Orders[0]
	if _, off := o.PayTime.Zone(); off != 8*3600 {
		t.Fatalf("zone offset = %d", off)
	}
	if o.PayTime.Unix() != 1704038400 {
		t.Fatalf("unix = %d", o.PayTime.Unix())
	}
	if !strings.Contains(joinPostings(o), "Expenses:Matched") {
		t.Fatalf(".timestamp should use the template zone:\n%s", joinPostings(o))
	}
	bad := loadProfileYAML(t, "template: { slots: { date: <d>, amount: <a> }, timezone: Mars/Base }")
	if err := validateTemplate(*bad); err == nil {
		t.Fatal("unknown timezone should be rejected")
	}
}

// other is the counterparty: one rule covers a purchase and its refund.
func TestOtherRoleFollowsDirection(t *testing.T) {
	out := importCSV(t, `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  slots: { date: <d>, payee: <p>, amount: <a>.number }
accounts: { self: Assets:Card }
personalRules:
  - id: 咖啡店
    when: payee ~ "咖啡"
    actions: { other: Expenses:Coffee }
`, "d,p,a\n2026-01-01,咖啡店,-30\n2026-01-02,咖啡店,30\n")
	buy, refund := joinPostings(out.Orders[0]), joinPostings(out.Orders[1])
	if !strings.Contains(buy, "Expenses:Coffee 30") || !strings.Contains(buy, "Assets:Card -30") {
		t.Fatalf("purchase:\n%s", buy)
	}
	if !strings.Contains(refund, "Expenses:Coffee -30") || !strings.Contains(refund, "Assets:Card 30") {
		t.Fatalf("refund:\n%s", refund)
	}
}

// Only the user knows which address is theirs; a rule can say which way
// the money moved.
func TestRuleDirectionOverride(t *testing.T) {
	out := importCSV(t, `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: ETH
  slots: { date: <d>, amount: <v>.number, metadata: { from: <from>, to: <to> } }
accounts: { self: Assets:Wallet }
personalRules:
  - id: 我发出
    when: metadata.from == "0xme"
    actions: { direction: outflow, other: Expenses:Sent }
  - id: 我收到
    when: metadata.to == "0xme"
    actions: { direction: inflow, other: Income:Received }
`, "d,from,to,v\n2026-01-01,0xme,0xyou,1\n2026-01-02,0xyou,0xme,2\n")
	sent, recv := out.Orders[0], out.Orders[1]
	if sent.Type != ir.TypeSend || !strings.Contains(joinPostings(sent), "Assets:Wallet -1") || !strings.Contains(joinPostings(sent), "Expenses:Sent 1") {
		t.Fatalf("sent %v:\n%s", sent.Type, joinPostings(sent))
	}
	if recv.Type != ir.TypeRecv || !strings.Contains(joinPostings(recv), "Assets:Wallet 2") || !strings.Contains(joinPostings(recv), "Income:Received -2") {
		t.Fatalf("recv %v:\n%s", recv.Type, joinPostings(recv))
	}
	if _, err := ParseRulesFile([]byte("personalRules: [{id: x, actions: {direction: sideways}}]")); err != nil {
		t.Fatal(err)
	}
}

// A newest-first statement comes out oldest-first, same-day order kept.
func TestNewestFirstBillIsReversed(t *testing.T) {
	profile := loadProfileYAML(t, `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  slots: { date: <d>, narration: <n>, amount: <a>.number }
`)
	out, _, err := ImportBytes(profile, "b.csv", []byte("d,n,a\n2026-01-02,晚,-3\n2026-01-01,中,-2\n2026-01-01,早,-1\n"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, o := range out.Orders {
		got = append(got, o.Item)
	}
	if strings.Join(got, ",") != "早,中,晚" {
		t.Fatalf("order = %v", got)
	}
	_ = time.UTC
}

func TestVarsKeepBillTextLiteral(t *testing.T) {
	row := Row{Raw: map[string]string{"日": "12/29", "y": "2024"}}
	got := rowWithVars(row, map[string]string{"md": "<日>", "prev": "<y> - 1"}, ir.Order{})
	if got.Raw["var.md"] != "12/29" {
		t.Fatalf("bill text was evaluated: %q", got.Raw["var.md"])
	}
	if got.Raw["var.prev"] != "2023.00" && got.Raw["var.prev"] != "2023" {
		t.Fatalf("template arithmetic not evaluated: %q", got.Raw["var.prev"])
	}
}
