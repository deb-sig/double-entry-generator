package cib_debit

import (
	"path/filepath"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/xuri/excelize/v2"
)

func TestParseCurrencyFromAccountLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "cny", line: "001 活期储蓄存款 人民币", want: "CNY"},
		{name: "hkd", line: "002 活期储蓄存款 港币 现汇", want: "HKD"},
		{name: "usd", line: "003 活期储蓄存款 美元 现汇", want: "USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCurrency(tt.line)
			if got != tt.want {
				t.Fatalf("parseCurrency(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestTranslateFilesMergesExcelSubAccounts(t *testing.T) {
	dir := t.TempDir()
	files := []struct {
		name       string
		subAccount string
		row        []string
	}{
		{
			name:       "cny.xlsx",
			subAccount: "001 活期储蓄存款 人民币",
			row:        []string{"2024-01-01 00:00:00", "2024-01-01", "10.00", "", "90.00", "购汇", "示例用户", "示例银行", "PEER-MOCK-CNY-FX", "", "手机银行", ""},
		},
		{
			name:       "hkd.xlsx",
			subAccount: "002 活期储蓄存款 港币 现汇",
			row:        []string{"2024-01-01 00:00:00", "2024-01-01", "", "20.00", "20.00", "购汇", "示例用户", "示例银行", "PEER-MOCK-HKD-FX", "", "手机银行", ""},
		},
		{
			name:       "usd.xlsx",
			subAccount: "003 活期储蓄存款 美元 现汇",
			row:        []string{"2024-01-02 00:00:00", "2024-01-02", "", "30.00", "30.00", "购汇", "示例用户", "示例银行", "PEER-MOCK-USD-FX", "", "手机银行", ""},
		},
	}

	paths := make([]string, 0, len(files))
	for _, fixture := range files {
		path := filepath.Join(dir, fixture.name)
		writeTestWorkbook(t, path, fixture.subAccount, [][]string{fixture.row})
		paths = append(paths, path)
	}

	got, err := New().TranslateFiles(paths)
	if err != nil {
		t.Fatalf("TranslateFiles returned error: %v", err)
	}
	if len(got.Orders) != 2 {
		t.Fatalf("len(orders) = %d, want 2", len(got.Orders))
	}
	exchange := got.Orders[0]
	if exchange.OrderType != ir.OrderTypeCurrencyExchange {
		t.Fatalf("first order type = %s, want %s", exchange.OrderType, ir.OrderTypeCurrencyExchange)
	}
	if exchange.Currency != "CNY" || exchange.Money != 10 {
		t.Fatalf("source side = %.2f %s, want 10.00 CNY", exchange.Money, exchange.Currency)
	}
	if exchange.Amount != 20 || exchange.Units[ir.TargetUnit] != "HKD" {
		t.Fatalf("target side = %.2f %s, want 20.00 HKD", exchange.Amount, exchange.Units[ir.TargetUnit])
	}
	if exchange.Metadata["accountNum"] != "CARD-MOCK-001" {
		t.Fatalf("exchange accountNum metadata missing")
	}
	if got.Orders[1].Currency != "USD" {
		t.Fatalf("second order currency = %q, want USD", got.Orders[1].Currency)
	}
}

func writeTestWorkbook(t *testing.T, path, subAccount string, rows [][]string) {
	t.Helper()

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	values := [][]string{
		{"兴业银行交易明细"},
		{},
		{"账户别名"},
		{"账户户名", "示例用户"},
		{"账户账号", "CARD-MOCK-001"},
		{"卡内账户", subAccount},
		{"起始日期", "2024-01-01"},
		{"截止日期", "2024-12-31"},
		{"下载日期", "2026-05-10 02:54:19"},
		{},
		{"交易时间", "记账日", "支出", "收入", "账户余额", "摘要", "对方户名", "对方银行", "对方账号", "用途", "交易渠道", "备注"},
	}
	values = append(values, rows...)
	values = append(values, []string{"说明", "交易明细涉及您的个人隐私，请妥善处理，避免信息篡改或泄露，交易明细内容仅供个人参考。"})

	for rowIdx, row := range values {
		for colIdx, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err != nil {
				t.Fatalf("CoordinatesToCellName returned error: %v", err)
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				t.Fatalf("SetCellValue returned error: %v", err)
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs returned error: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestTranslateRowHandlesSignedDebitAndCreditColumns(t *testing.T) {
	provider := New()
	provider.Currency = "HKD"
	provider.AccountNum = "CARD-MOCK-001"
	provider.SubAccount = "002 活期储蓄存款 港币 现汇"

	rows := [][]string{
		{"2024-01-01 00:00:00", "2024-01-01", "", "20.00", "20.00", "购汇", "示例用户", "示例银行", "PEER-MOCK-HKD-FX", "", "手机银行", ""},
		{"2024-01-02 00:00:00", "2024-01-02", "-20.00", "", "40.00", "当日冲账", "MOCK USER", "MOCK BANK", "PEER-MOCK-HKD-OUT", "冲正", "手机银行", ""},
		{"2024-01-03 00:00:00", "2024-01-03", "40.00", "", "0.00", "转账转出", "MOCK USER", "MOCK BANK", "PEER-MOCK-HKD-OUT", "测试用途", "手机银行", ""},
	}

	for _, row := range rows {
		if err := provider.translateRow(row); err != nil {
			t.Fatalf("translateRow returned error: %v", err)
		}
	}

	got, err := provider.convertToIR()
	if err != nil {
		t.Fatalf("convertToIR returned error: %v", err)
	}
	if len(got.Orders) != 3 {
		t.Fatalf("len(orders) = %d, want 3", len(got.Orders))
	}

	wantTypes := []ir.Type{ir.TypeRecv, ir.TypeRecv, ir.TypeSend}
	wantAmounts := []float64{20, 20, 40}
	for idx, order := range got.Orders {
		if order.Type != wantTypes[idx] {
			t.Fatalf("order %d type = %s, want %s", idx, order.Type, wantTypes[idx])
		}
		if order.Money != wantAmounts[idx] {
			t.Fatalf("order %d money = %v, want %v", idx, order.Money, wantAmounts[idx])
		}
		if order.Currency != "HKD" {
			t.Fatalf("order %d currency = %q, want HKD", idx, order.Currency)
		}
		if order.Metadata["accountNum"] != "CARD-MOCK-001" {
			t.Fatalf("order %d accountNum metadata missing", idx)
		}
		if order.Metadata["subAccount"] != "002 活期储蓄存款 港币 现汇" {
			t.Fatalf("order %d subAccount metadata missing", idx)
		}
	}
}

func TestTranslateRowReturnsMalformedAmountError(t *testing.T) {
	provider := New()

	err := provider.translateRow([]string{
		"2024-01-01 00:00:00",
		"2024-01-01",
		"not-a-number",
		"",
		"90.00",
		"转账转出",
	})
	if err == nil {
		t.Fatal("translateRow returned nil error for malformed amount")
	}
	if len(provider.Orders) != 0 {
		t.Fatalf("len(orders) = %d, want 0", len(provider.Orders))
	}
}
