package ibkr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func TestTranslateParsesFlexStatementTradesAndCashTransactions(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="last 365 days" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U123" fromDate="2025-01-01" toDate="2025-01-31">
      <Trades>
        <Trade accountId="U123" currency="USD" assetCategory="STK" subCategory="ETF" symbol="QQQ" description="INVESCO QQQ TRUST SERIES 1" tradeID="T1" tradeDate="2025-01-16" dateTime="2025-01-16,12:23:56 EST" quantity="2" tradePrice="500.12" tradeMoney="1000.24" proceeds="-1000.24" ibCommission="-0.35" ibCommissionCurrency="USD" netCash="-1000.59" buySell="BUY" transactionID="TX1" />
        <Trade accountId="U123" currency="USD" assetCategory="STK" subCategory="COMMON" symbol="TSLA" description="TESLA INC" tradeID="T2" tradeDate="2025-01-17" dateTime="2025-01-17,09:30:00 EST" quantity="-1" tradePrice="330.50" tradeMoney="-330.50" proceeds="330.50" ibCommission="-0.36" ibCommissionCurrency="USD" netCash="330.14" buySell="SELL" transactionID="TX2" />
        <Trade accountId="U123" currency="HKD" assetCategory="CASH" symbol="USD.HKD" description="USD.HKD" tradeID="FX1" tradeDate="2025-01-18" quantity="10" tradePrice="7.8" tradeMoney="78" buySell="BUY" transactionID="FXTX1" />
      </Trades>
      <CashTransactions>
        <CashTransaction accountId="U123" currency="USD" assetCategory="STK" symbol="MSFT" description="MSFT CASH DIVIDEND" dateTime="2025-01-20,20:20:00 EST" settleDate="2025-01-20" amount="0.83" type="Dividends" transactionID="C1" />
        <CashTransaction accountId="U123" currency="USD" assetCategory="STK" symbol="MSFT" description="MSFT CASH DIVIDEND - US TAX" dateTime="2025-01-20,20:20:00 EST" settleDate="2025-01-20" amount="-0.08" type="Withholding Tax" transactionID="C2" />
      </CashTransactions>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	path := filepath.Join(t.TempDir(), "ibkr.xml")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := New().Translate(path)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	if len(got.Orders) != 5 {
		t.Fatalf("len(Orders) = %d, want 5", len(got.Orders))
	}

	buy := got.Orders[0]
	if buy.OrderType != ir.OrderTypeSecuritiesTrade || buy.Type != ir.TypeSend {
		t.Fatalf("buy type = (%s, %s), want securities send", buy.OrderType, buy.Type)
	}
	if buy.Item != "QQQ-INVESCO QQQ TRUST SERIES 1" || buy.TxTypeOriginal != "QQQ" {
		t.Fatalf("buy security = (%q, %q)", buy.Item, buy.TxTypeOriginal)
	}
	if buy.Amount != 2 || buy.Money != 1000.24 || buy.Price != 500.12 || buy.Commission != 0.35 || buy.Currency != "USD" {
		t.Fatalf("buy amounts = amount %v money %v price %v commission %v currency %s", buy.Amount, buy.Money, buy.Price, buy.Commission, buy.Currency)
	}
	if buy.Metadata["trade_id"] != "T1" || buy.Metadata["transaction_id"] != "TX1" {
		t.Fatalf("buy metadata = %#v", buy.Metadata)
	}

	sell := got.Orders[1]
	if sell.OrderType != ir.OrderTypeSecuritiesTrade || sell.Type != ir.TypeRecv {
		t.Fatalf("sell type = (%s, %s), want securities recv", sell.OrderType, sell.Type)
	}
	if sell.Amount != 1 || sell.Money != 330.50 || sell.Commission != 0.36 {
		t.Fatalf("sell amounts = amount %v money %v commission %v", sell.Amount, sell.Money, sell.Commission)
	}

	fx := got.Orders[2]
	if fx.OrderType != ir.OrderTypeCurrencyExchange {
		t.Fatalf("fx OrderType = %s, want currency exchange", fx.OrderType)
	}
	if fx.Currency != "HKD" || fx.Units[ir.TargetUnit] != "USD" || fx.Money != 78 || fx.Amount != 10 {
		t.Fatalf("fx = source %v %s target %v %s", fx.Money, fx.Currency, fx.Amount, fx.Units[ir.TargetUnit])
	}

	dividend := got.Orders[3]
	if dividend.OrderType != ir.OrderTypeNormal || dividend.Type != ir.TypeRecv || dividend.Money != 0.83 || dividend.Currency != "USD" {
		t.Fatalf("dividend = %#v", dividend)
	}

	tax := got.Orders[4]
	if tax.OrderType != ir.OrderTypeNormal || tax.Type != ir.TypeSend || tax.Money != 0.08 || tax.Currency != "USD" {
		t.Fatalf("tax = %#v", tax)
	}
}

func TestCashTradeToOrderRejectsUnknownBuySell(t *testing.T) {
	_, ok, err := cashTradeToOrder(map[string]string{
		"tradeDate":  "2025-01-18",
		"quantity":   "10",
		"tradeMoney": "78",
		"symbol":     "USD.HKD",
		"currency":   "HKD",
		"buySell":    "UNKNOWN",
	})
	if err == nil {
		t.Fatal("cashTradeToOrder returned nil error for unsupported buySell")
	}
	if ok {
		t.Fatal("cashTradeToOrder returned ok=true for unsupported buySell")
	}
}
