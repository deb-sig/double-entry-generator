package cgb_credit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// 测试广发信用卡 CSV 能按金额方向和摘要类型转换成 IR 交易。
func TestTranslateCgbCreditCsv(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "cgb-credit.csv")
	content := `交易日期,入账日期,交易摘要,交易金额,交易货币,入账金额,入账货币
2025/01/05,2025/01/06,（消费）示例商户A,15.90,人民币,15.90,人民币
2025/01/04,2025/01/04,（退款）示例商户B,-97.20,人民币,-97.20,人民币
2025/01/03,2025/01/03,（分期）银联分期00000000000/10/12期,387.42,人民币,387.42,人民币
2024/12/25,2024/12/26,（还款）云闪付 银联入账-张三,-1130.68,人民币,-1130.68,人民币
`
	if err := os.WriteFile(inputFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	provider := New()
	got, err := provider.Translate(inputFile)
	if err != nil {
		t.Fatalf("translate cgb credit csv: %v", err)
	}

	if len(got.Orders) != 4 {
		t.Fatalf("orders length = %d, want 4", len(got.Orders))
	}

	tests := []struct {
		index        int
		wantType     ir.Type
		wantMoney    float64
		wantCurrency string
		wantItem     string
		wantDate     string
		wantPostDate string
	}{
		{0, ir.TypeSend, 15.90, "CNY", "（消费）示例商户A", "2025-01-05", "2025-01-06"},
		{1, ir.TypeRecv, 97.20, "CNY", "（退款）示例商户B", "2025-01-04", "2025-01-04"},
		{2, ir.TypeSend, 387.42, "CNY", "（分期）银联分期00000000000/10/12期", "2025-01-03", "2025-01-03"},
		{3, ir.TypeRecv, 1130.68, "CNY", "（还款）云闪付 银联入账-张三", "2024-12-25", "2024-12-26"},
	}

	for _, tt := range tests {
		order := got.Orders[tt.index]
		if order.Type != tt.wantType {
			t.Errorf("order[%d].Type = %v, want %v", tt.index, order.Type, tt.wantType)
		}
		if order.Money != tt.wantMoney {
			t.Errorf("order[%d].Money = %.2f, want %.2f", tt.index, order.Money, tt.wantMoney)
		}
		if order.Currency != tt.wantCurrency {
			t.Errorf("order[%d].Currency = %q, want %q", tt.index, order.Currency, tt.wantCurrency)
		}
		if order.Item != tt.wantItem {
			t.Errorf("order[%d].Item = %q, want %q", tt.index, order.Item, tt.wantItem)
		}
		if got := order.PayTime.Format(dateLayout); got != tt.wantDate {
			t.Errorf("order[%d].PayTime = %q, want %q", tt.index, got, tt.wantDate)
		}
		if got := order.Metadata["recordDate"]; got != tt.wantPostDate {
			t.Errorf("order[%d].recordDate = %q, want %q", tt.index, got, tt.wantPostDate)
		}
		if got := order.Metadata["source"]; got != "广发银行信用卡" {
			t.Errorf("order[%d].source = %q, want 广发银行信用卡", tt.index, got)
		}
	}
}
