package importer

import (
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func importCSV(t *testing.T, profileYAML, csv string) *ir.IR {
	t.Helper()
	profile := loadProfileYAML(t, profileYAML)
	if err := validateTemplate(*profile); err != nil {
		t.Fatalf("validate: %v", err)
	}
	rows, err := ParseBytes(profile, "bill.csv", []byte(csv))
	if err != nil {
		t.Fatal(err)
	}
	out := ir.New()
	for _, row := range rows {
		order, ignore, err := rowToImportOrder(profile, row)
		if err != nil {
			t.Fatal(err)
		}
		if !ignore {
			out.Orders = append(out.Orders, order)
		}
	}
	return out
}

const securitiesTemplate = `
schema: https://deg.dev/template-profile/v2
id: broker
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  sourceHeaders: [成交日期, 操作, 证券代码, 证券名称, 成交数量, 成交价格, 成交金额, 手续费, 股东代码]
  defaultCurrency: CNY
  slots:
    date: <成交日期>
    narration: <操作>-<证券名称>
    amount: <成交金额>.number
    metadata:
      code: <证券代码>
  vars:
    security: SH<证券代码>
  legs:
    - id: 买入
      when: <操作> == "证券买入"
      legs:
        - { role: cash,     amount: "-<成交金额>", currency: CNY }
        - { role: position, amount: "<成交数量>", currency: "<var.security>", cost: "<成交价格> CNY", price: "@@ <成交金额> CNY" }
        - { role: cash,     amount: "-<手续费>", currency: CNY }
        - { role: fee,      amount: "<手续费>",  currency: CNY }
    - id: 红利
      when: <操作> == "红利入账"
      narration: 红利-<证券名称>
      legs:
        - { role: cash, amount: "<成交金额>" }
        - { role: pnl,  amount: "-<成交金额>" }
`

const securitiesBill = "成交日期,操作,证券代码,证券名称,成交数量,成交价格,成交金额,手续费,股东代码\n" +
	"2026-03-01,证券买入,510300,HS300ETF,100,4.000,400.00,5.00,A1\n" +
	"2026-03-02,红利入账,510300,HS300ETF,0,0,12.50,0,A1\n" +
	"2026-03-03,银行转证券,,,0,0,1000.00,0,A1\n"

func TestLegsBindRolesToUserAccounts(t *testing.T) {
	profile := securitiesTemplate + `
accounts:
  cash: Assets:Broker:Cash
  position: Assets:Broker:Positions
  fee: Expenses:Broker:Commission
  pnl: Income:Broker:PnL
  from: Assets:Bank
  to: Assets:Broker:Cash
`
	out := importCSV(t, profile, securitiesBill)
	if len(out.Orders) != 3 {
		t.Fatalf("orders = %d", len(out.Orders))
	}
	buy := joinPostings(out.Orders[0])
	for _, want := range []string{
		"Assets:Broker:Cash -400.00 CNY",
		"Assets:Broker:Positions 100 SH510300 {4.000 CNY} @@ 400.00 CNY",
		"Assets:Broker:Cash -5.00 CNY",
		"Expenses:Broker:Commission 5.00 CNY",
	} {
		if !strings.Contains(buy, want) {
			t.Errorf("buy postings missing %q:\n%s", want, buy)
		}
	}
	if out.Orders[0].Metadata["code"] != "510300" {
		t.Errorf("metadata = %v", out.Orders[0].Metadata)
	}
	div := out.Orders[1]
	if div.Item != "红利-HS300ETF" {
		t.Errorf("branch narration = %q", div.Item)
	}
	if text := joinPostings(div); !strings.Contains(text, "Income:Broker:PnL -12.50 CNY") || !strings.Contains(text, "Assets:Broker:Cash 12.50 CNY") {
		t.Errorf("dividend postings:\n%s", text)
	}
	// No branch matched: plain two-leg shape bound through from/to roles.
	transfer := joinPostings(out.Orders[2])
	if !strings.Contains(transfer, "Assets:Broker:Cash 1000.00 CNY") || !strings.Contains(transfer, "Assets:Bank -1000.00 CNY") {
		t.Errorf("transfer postings:\n%s", transfer)
	}
}

func TestLegsFallBackToRoleFIXME(t *testing.T) {
	out := importCSV(t, securitiesTemplate, securitiesBill)
	buy := joinPostings(out.Orders[0])
	for _, want := range []string{"Assets:FIXME -400.00 CNY", "Assets:FIXME 100 SH510300", "Expenses:FIXME 5.00 CNY"} {
		if !strings.Contains(buy, want) {
			t.Errorf("missing %q:\n%s", want, buy)
		}
	}
	if !strings.Contains(joinPostings(out.Orders[1]), "Income:FIXME -12.50") {
		t.Errorf("pnl fallback:\n%s", joinPostings(out.Orders[1]))
	}
	var engineLegs int
	for _, src := range out.Orders[0].Sources {
		if strings.HasPrefix(src.Slot, "leg.") && src.Origin == ir.FieldOriginEngine {
			engineLegs++
		}
	}
	if engineLegs == 0 {
		t.Errorf("engine-filled legs should be recorded in Sources: %+v", out.Orders[0].Sources)
	}
}

func TestLegsRuleAccountsOverrideFileAccounts(t *testing.T) {
	profile := securitiesTemplate + `
accounts:
  cash: Assets:Broker:Cash
personalRules:
  - id: 沪深300单独记
    when: metadata.code == "510300"
    actions:
      accounts:
        position: Assets:Broker:Positions:HS300
`
	out := importCSV(t, profile, securitiesBill)
	buy := joinPostings(out.Orders[0])
	if !strings.Contains(buy, "Assets:Broker:Positions:HS300 100 SH510300") || !strings.Contains(buy, "Assets:Broker:Cash -400.00") {
		t.Errorf("postings:\n%s", buy)
	}
	var ruleBound bool
	for _, src := range out.Orders[0].Sources {
		if src.Slot == "leg.position" && src.RuleID == "沪深300单独记" {
			ruleBound = true
		}
	}
	if !ruleBound {
		t.Errorf("rule binding not recorded: %+v", out.Orders[0].Sources)
	}
}

func TestLegsRejectAccountsAndUnknownRolesInTemplate(t *testing.T) {
	var p Profile
	err := yamlUnmarshal(`
template:
  legs:
    - legs:
        - { role: cash, account: Assets:Mine, amount: "1" }
`, &p)
	if err == nil || !strings.Contains(err.Error(), "accounts:") {
		t.Fatalf("account in template leg should be rejected, got %v", err)
	}
	profile := loadProfileYAML(t, `
template:
  slots: { date: <d>, amount: <a> }
  legs:
    - legs:
        - { role: margin, amount: "1" }
`)
	if err := validateTemplate(*profile); err == nil || !strings.Contains(err.Error(), "margin") {
		t.Fatalf("unknown role should be rejected, got %v", err)
	}
	profile = loadProfileYAML(t, `
template:
  slots: { date: <d>, amount: <a> }
  vars: { cash: Assets:Broker:Cash }
`)
	if err := validateTemplate(*profile); err == nil || !strings.Contains(err.Error(), "vars.cash") {
		t.Fatalf("account in template vars should be rejected, got %v", err)
	}
}

func TestLegsCustomRoleMustBeBound(t *testing.T) {
	tpl := `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  sourceHeaders: [d, a]
  defaultCurrency: CNY
  slots: { date: <d>, amount: <a>.number }
  legs:
    - legs:
        - { role: x-margin, amount: "<a>" }
        - { role: cash, amount: "-<a>" }
`
	profile := loadProfileYAML(t, tpl)
	rows, err := ParseBytes(profile, "b.csv", []byte("d,a\n2026-01-01,5\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := rowToImportOrder(profile, rows[0]); err == nil || !strings.Contains(err.Error(), "x-margin") {
		t.Fatalf("unbound x- role must error, got %v", err)
	}
	out := importCSV(t, tpl+"\naccounts: { x-margin: Liabilities:Broker:Margin }\n", "d,a\n2026-01-01,5\n")
	if !strings.Contains(joinPostings(out.Orders[0]), "Liabilities:Broker:Margin 5 CNY") {
		t.Errorf("postings:\n%s", joinPostings(out.Orders[0]))
	}
}

func TestDirectionForms(t *testing.T) {
	base := `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  sourceHeaders: [d, kind, a, out, in]
  slots: { date: <d>, amount: <a>.number }
`
	bill := "d,kind,a,out,in\n2026-01-01,支出,10,10,\n2026-01-02,收入,20,,20\n2026-01-03,其他,-30,,\n"

	byColumn := importCSV(t, base+`
  direction: { column: <kind>, outflow: [支出], inflow: [收入] }
`, bill)
	if byColumn.Orders[0].Type != ir.TypeSend || byColumn.Orders[1].Type != ir.TypeRecv || byColumn.Orders[2].Type != ir.TypeSend {
		t.Errorf("column form types: %v %v %v", byColumn.Orders[0].Type, byColumn.Orders[1].Type, byColumn.Orders[2].Type)
	}
	if byColumn.Orders[0].TypeOriginal != "支出" {
		t.Errorf("TypeOriginal = %q", byColumn.Orders[0].TypeOriginal)
	}

	twoColumns := importCSV(t, strings.Replace(base, "amount: <a>.number", "", 1)+`
  direction: { outflowColumn: <out>, inflowColumn: <in> }
`, "d,kind,a,out,in\n2026-01-01,,,10.50,\n2026-01-02,,,,20\n")
	if got := joinPostings(twoColumns.Orders[0]); !strings.Contains(got, "Assets:FIXME -10.50 CNY") || !strings.Contains(got, "Expenses:FIXME 10.50 CNY") {
		t.Errorf("outflow column:\n%s", got)
	}
	if got := joinPostings(twoColumns.Orders[1]); !strings.Contains(got, "Income:FIXME -20.00 CNY") {
		t.Errorf("inflow column:\n%s", got)
	}

	bySign := importCSV(t, base, "d,kind,a,out,in\n2026-01-01,,-7,,\n2026-01-02,,8,,\n")
	if bySign.Orders[0].Type != ir.TypeSend || bySign.Orders[1].Type != ir.TypeRecv {
		t.Errorf("sign form types: %v %v", bySign.Orders[0].Type, bySign.Orders[1].Type)
	}
}

func TestDirectionValidation(t *testing.T) {
	for _, tpl := range []string{
		`template: { slots: { date: <d>, amount: <a> }, direction: { outflowColumn: <o> } }`,
		`template: { slots: { date: <d>, amount: <a> }, direction: { outflow: [支出] } }`,
		`template: { slots: { date: <d> }, direction: { column: <k> } }`,
	} {
		profile := loadProfileYAML(t, tpl)
		if err := validateTemplate(*profile); err == nil {
			t.Errorf("expected validation error for %s", tpl)
		}
	}
}

func TestSlotSkeletonListsRoles(t *testing.T) {
	profile := loadProfileYAML(t, securitiesTemplate)
	text := SlotSkeleton("broker@2026-01-01", profile)
	for _, want := range []string{"accounts:", "  self: Assets:FIXME", "  cash: Assets:FIXME", "  position: Assets:FIXME", "  fee: Expenses:FIXME", "  pnl: Income:FIXME", "rules:"} {
		if !strings.Contains(text, want) {
			t.Errorf("skeleton missing %q:\n%s", want, text)
		}
	}
}

func TestSelfRoleFollowsDirection(t *testing.T) {
	tpl := `
schema: https://deg.dev/template-profile/v2
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  sourceHeaders: [d, kind, a]
  slots: { date: <d>, amount: <a>.number }
  direction: { column: <kind>, outflow: [支出], inflow: [收入], default: outflow }
accounts: { self: Assets:Bank }
`
	out := importCSV(t, tpl, "d,kind,a\n2026-01-01,支出,10\n2026-01-02,收入,20\n2026-01-03,其他,5\n")
	if got := joinPostings(out.Orders[0]); !strings.Contains(got, "Assets:Bank -10") || !strings.Contains(got, "Expenses:FIXME 10") {
		t.Errorf("outflow:\n%s", got)
	}
	if got := joinPostings(out.Orders[1]); !strings.Contains(got, "Assets:Bank 20") || !strings.Contains(got, "Income:FIXME -20") {
		t.Errorf("inflow:\n%s", got)
	}
	if out.Orders[2].Type != ir.TypeSend {
		t.Errorf("default direction should be outflow, got %v", out.Orders[2].Type)
	}
}
