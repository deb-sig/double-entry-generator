package embed

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

const template = `
schema: https://deg.dev/template-profile/v2
id: bank
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  sourceHeaders: [日期, 商户, 摘要, 金额, 流水号]
  slots:
    date: <日期>
    payee: <商户>
    narration: <摘要>
    amount: <金额>.number
    metadata:
      serial: <流水号>
`

const rules = `
accounts:
  self: Assets:Bank:Card
personalRules:
- id: coffee
  when: payee ~ "咖啡"
  actions:
    other: Expenses:Coffee
`

const bill = "日期,商户,摘要,金额,流水号\n2026-01-01,公司,工资,100.00,a\n2026-01-02,咖啡馆,拿铁,-30.50,b\n"

func TestRunReturnsStructuredTransactions(t *testing.T) {
	res := Run(template, rules, "bill.csv", []byte(bill))
	if !res.OK || res.Error != "" {
		t.Fatalf("run failed: %+v", res)
	}
	if res.Contract != ContractVersion || len(res.Transactions) != 2 {
		t.Fatalf("contract %d, %d transactions", res.Contract, len(res.Transactions))
	}
	coffee := res.Transactions[1]
	if coffee.Date != "2026-01-02" || coffee.Payee != "咖啡馆" || coffee.Narration != "拿铁" {
		t.Fatalf("coffee = %+v", coffee)
	}
	if coffee.From != "Assets:Bank:Card" || coffee.To != "Expenses:Coffee" || coffee.Amount != "30.50" || coffee.Currency != "CNY" {
		t.Fatalf("coffee legs = %+v", coffee)
	}
	if len(coffee.Postings) != 2 || coffee.Postings[0] != (Posting{Account: "Expenses:Coffee", Amount: "30.50", Currency: "CNY", Line: coffee.Postings[0].Line}) {
		t.Fatalf("postings = %+v", coffee.Postings)
	}
	if coffee.NeedsAccount {
		t.Fatalf("coffee is fully bound by rules: %+v", coffee.Sources)
	}
	if len(coffee.Metadata) != 1 || coffee.Metadata[0] != [2]string{"serial", "b"} {
		t.Fatalf("metadata = %v", coffee.Metadata)
	}
	var ruleSource bool
	for _, s := range coffee.Sources {
		if s.Rule == "coffee" && s.Origin == "rule" {
			ruleSource = true
		}
	}
	if !ruleSource {
		t.Fatalf("sources should name the coffee rule: %+v", coffee.Sources)
	}
	salary := res.Transactions[0]
	if !salary.NeedsAccount || res.NeedsAccount != 1 {
		t.Fatalf("salary has no counterparty rule, needsAccount should be set: %+v / %d", salary, res.NeedsAccount)
	}
	if !strings.Contains(res.Beancount, "Expenses:Coffee") || !strings.Contains(res.Beancount, `2026-01-02 * "咖啡馆" "拿铁"`) {
		t.Fatalf("beancount text missing the coffee entry:\n%s", res.Beancount)
	}
}

func TestRunReportsErrorsInsteadOfPanicking(t *testing.T) {
	for name, tc := range map[string][3]string{
		"bad template": {"template: [", "", bill},
		"bad rules":    {template, "personalRules: {", bill},
		"wrong bill":   {template, "", "no,such,headers\n1,2,3\n"},
	} {
		res := Run(tc[0], tc[1], "bill.csv", []byte(tc[2]))
		if res.OK || res.Error == "" || res.Transactions == nil || res.Warnings == nil {
			t.Fatalf("%s: want an error result with empty lists, got %+v", name, res)
		}
	}
}

func TestJSONIsStableAndConcurrencySafe(t *testing.T) {
	want := string(JSON(template, rules, "bill.csv", []byte(bill)))
	var parsed map[string]any
	if err := json.Unmarshal([]byte(want), &parsed); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"contract", "ok", "beancount", "transactions", "needsAccount", "warnings"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("JSON lacks %q: %s", key, want)
		}
	}
	var wg sync.WaitGroup
	errs := make(chan string, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := string(JSON(template, rules, "bill.csv", []byte(bill))); got != want {
				errs <- got
			}
		}()
	}
	wg.Wait()
	close(errs)
	for got := range errs {
		t.Fatalf("concurrent run differs:\n%s\nwant\n%s", got, want)
	}
}

func TestParsePostingKeepsCostAndPrice(t *testing.T) {
	got := parsePosting("Assets:Broker:Positions 100.00 SH600000 {10.500 CNY} @@ 1050.00 CNY")
	want := Posting{Account: "Assets:Broker:Positions", Amount: "100.00", Currency: "SH600000", Cost: "10.500 CNY", Price: "@@ 1050.00 CNY", Line: got.Line}
	if got != want {
		t.Fatalf("got %+v", got)
	}
	if p := parsePosting("Income:Broker:PnL"); p.Account != "Income:Broker:PnL" || p.Amount != "" {
		t.Fatalf("auto-balanced leg = %+v", p)
	}
}
