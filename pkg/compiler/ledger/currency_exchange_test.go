package ledger

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func TestWriteCurrencyExchangeUsesTotalPrice(t *testing.T) {
	c := &config.Config{DefaultCurrency: "CNY"}
	i := &ir.IR{Orders: []ir.Order{{
		OrderType:    ir.OrderTypeCurrencyExchange,
		Peer:         "CIB Debit",
		Item:         "购汇 CNY/USD",
		PayTime:      time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		Money:        40,
		Amount:       40,
		Currency:     "CNY",
		MinusAccount: "Assets:DebitCard:CIB",
		PlusAccount:  "Assets:DebitCard:CIB",
		Units: map[ir.Unit]string{
			ir.TargetUnit: "USD",
		},
		Metadata: map[string]string{},
	}}}
	compiler, err := New("cib_debit", "ledger", "", false, c, i, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	var out bytes.Buffer
	if err := compiler.writeBill(&out, 0); err != nil {
		t.Fatalf("writeBill returned error: %v", err)
	}

	want := "Assets:DebitCard:CIB      40.00 USD @@ 40.00 CNY"
	if !strings.Contains(out.String(), want) {
		t.Fatalf("output missing %q:\n%s", want, out.String())
	}
}

func TestWriteSecuritiesTradeUsesOrderCurrency(t *testing.T) {
	c := &config.Config{DefaultCurrency: "CNY"}
	i := &ir.IR{Orders: []ir.Order{{
		OrderType:      ir.OrderTypeSecuritiesTrade,
		Type:           ir.TypeSend,
		Peer:           "IBKR",
		TxTypeOriginal: "0700",
		TypeOriginal:   "BUY",
		Item:           "0700-TENCENT HOLDINGS LTD",
		PayTime:        time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		Money:          780,
		Amount:         2,
		Price:          390,
		Commission:     1.23,
		Currency:       "HKD",
		ExtraAccounts: map[ir.Account]string{
			ir.CashAccount:       "Assets:Trade:IBKR:Cash",
			ir.PositionAccount:   "Assets:Trade:IBKR:Positions",
			ir.CommissionAccount: "Expenses:Trade:IBKR:Commission",
			ir.PnlAccount:        "Income:Trade:IBKR:PnL",
		},
		Metadata: map[string]string{},
	}}}
	compiler, err := New("ibkr", "ledger", "", false, c, i, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	var out bytes.Buffer
	if err := compiler.writeBill(&out, 0); err != nil {
		t.Fatalf("writeBill returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "780.00 HKD") {
		t.Fatalf("output missing order currency HKD:\n%s", got)
	}
	if strings.Contains(got, "780.00 CNY") {
		t.Fatalf("output used default currency instead of order currency:\n%s", got)
	}
}
