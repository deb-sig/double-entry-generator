package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const bankTemplate = `
schema: https://deg.dev/template-profile/v2
id: bank
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  sourceHeaders: [日期, 摘要, 金额, 余额, 流水号]
  slots:
    date: <日期>
    narration: <摘要>
    amount: <金额>.number
    metadata:
      serial: <流水号>
  balance:
    column: <余额>
`

func TestRunningBalanceAcceptsEitherRowOrder(t *testing.T) {
	profile := loadProfileYAML(t, bankTemplate)
	oldestFirst := "日期,摘要,金额,余额,流水号\n2026-01-01,工资,100.00,1100.00,a\n2026-01-02,咖啡,-30.00,1070.00,b\n2026-01-03,,0.00,1070.00,c\n"
	newestFirst := "日期,摘要,金额,余额,流水号\n2026-01-02,咖啡,-30.00,1070.00,b\n2026-01-01,工资,100.00,1100.00,a\n"
	for _, bill := range []string{oldestFirst, newestFirst} {
		if _, _, err := ImportBytes(profile, "bill.csv", []byte(bill)); err != nil {
			t.Fatalf("balance check should pass:\n%s\n%v", bill, err)
		}
	}
}

func TestRunningBalanceCatchesMisreadRow(t *testing.T) {
	profile := loadProfileYAML(t, bankTemplate)
	bill := "日期,摘要,金额,余额,流水号\n2026-01-01,工资,100.00,1100.00,a\n2026-01-02,咖啡,-3.00,1070.00,b\n"
	_, _, err := ImportBytes(profile, "bill.csv", []byte(bill))
	if err == nil || !strings.Contains(err.Error(), "running balance breaks between rows 1 and 2") {
		t.Fatalf("err = %v", err)
	}
}

func TestDedupeWithinBillAndAgainstLedger(t *testing.T) {
	dir := t.TempDir()
	ledger := filepath.Join(dir, "main.bean")
	os.WriteFile(ledger, []byte(`option "title" "x"

2026-01-01 * "工资"
	serial: "a"
	Assets:Bank 100.00 CNY
	Income:Salary -100.00 CNY

2026-01-05 * "超市" "买菜"
	Assets:Bank -42.00 CNY
	Expenses:Food 42.00 CNY
`), 0o644)
	profile := loadProfileYAML(t, bankTemplate+`
reconcile:
  dedupe:
    key: [metadata.serial]
    against: `+ledger+`
    window: 2d
  flagEngineFilled: "!"
`)
	bill := strings.Join([]string{
		"日期,摘要,金额,余额,流水号",
		"2026-01-01,工资,100.00,,a", // already in ledger by serial
		"2026-01-02,咖啡,-30.00,,b",
		"2026-01-02,咖啡,-30.00,,b",  // repeated in the bill
		"2026-01-06,菜市场,-42.00,,c", // near the ledger's 42.00 on 01-05
		"2026-01-20,房租,-1000.00,,", // no serial: never deduped
	}, "\n") + "\n"
	out, report, err := ImportBytes(profile, "bill.csv", []byte(bill))
	if err != nil {
		t.Fatal(err)
	}
	if report.DroppedInLedger != 1 || report.DroppedInBill != 1 {
		t.Fatalf("report = %+v", report)
	}
	if len(out.Orders) != 3 {
		t.Fatalf("orders = %d", len(out.Orders))
	}
	near := out.Orders[1]
	if near.Item != "菜市场" || near.Flag != "!" {
		t.Fatalf("near duplicate should be flagged: %+v", near)
	}
	// Every order here has engine-filled sides, so all three carry "!".
	for _, o := range out.Orders {
		if o.Flag != "!" {
			t.Fatalf("engine-filled order not flagged: %q %q", o.Item, o.Flag)
		}
	}
}

func TestDedupeKeyNeedsEveryField(t *testing.T) {
	profile := loadProfileYAML(t, bankTemplate+`
reconcile:
  dedupe:
    key: [date, amount, metadata.serial]
`)
	bill := "日期,摘要,金额,余额,流水号\n2026-01-01,a,1.00,,\n2026-01-01,b,1.00,,\n2026-01-01,c,1.00,,s\n2026-01-01,d,1.00,,s\n"
	out, report, err := ImportBytes(profile, "bill.csv", []byte(bill))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 3 || report.DroppedInBill != 1 {
		t.Fatalf("orders = %d report = %+v", len(out.Orders), report)
	}
}

func TestLedgerIndexParsesPayeeNarrationMetadataAmounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "l.bean")
	os.WriteFile(path, []byte("2026-02-01 * \"商户\" \"说明 \\\"引号\\\"\"\n\torderId: \"42\"\n\tAssets:Digital:微信 -1,234.50 CNY\n\tExpenses:Food 1234.50 CNY\n\n2026-02-02 ! \"只有说明\"\n\tAssets:A 1 CNY\n"), 0o644)
	idx, err := loadLedgerIndex(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.txns) != 2 {
		t.Fatalf("txns = %d", len(idx.txns))
	}
	first := idx.txns[0]
	if first.payee != "商户" || first.metadata["orderId"] != "42" || len(first.amounts) != 2 || first.amounts[0].Text(2) != "1234.50" {
		t.Fatalf("first = %+v", first)
	}
	if idx.txns[1].narration != "只有说明" || idx.txns[1].payee != "" {
		t.Fatalf("second = %+v", idx.txns[1])
	}
}
