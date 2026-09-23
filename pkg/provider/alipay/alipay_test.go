package alipay

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func TestAlipayConfigIsNotSerialized(t *testing.T) {
	provider := New()
	provider.Config = &Config{KeepRefundRecords: true}

	data, err := json.Marshal(provider)
	if err != nil {
		t.Fatalf("marshal provider: %v", err)
	}
	if strings.Contains(string(data), "keepRefundRecords") {
		t.Fatal("expected runtime config to be excluded from provider JSON")
	}
}

func TestPostProcessRemovesMatchedRefundPairByDefault(t *testing.T) {
	provider := New()
	processed := provider.postProcess(refundPairIR())

	if got := len(processed.Orders); got != 0 {
		t.Fatalf("expected matched refund pair to be removed, got %d orders", got)
	}
}

func TestPostProcessKeepsMatchedRefundPairWhenConfigured(t *testing.T) {
	provider := New()
	provider.Config = &Config{KeepRefundRecords: true}
	processed := provider.postProcess(refundPairIR())

	if got := len(processed.Orders); got != 2 {
		t.Fatalf("expected matched refund pair to be kept, got %d orders", got)
	}
}

func TestPostProcessKeepsPartialRefund(t *testing.T) {
	provider := New()
	processed := provider.postProcess(&ir.IR{Orders: []ir.Order{
		refundOrder("order-1-refund", 10),
		originalOrder("order-1", 20),
	}})

	if got := len(processed.Orders); got != 2 {
		t.Fatalf("expected partial refund pair to remain, got %d orders", got)
	}
}

func TestPostProcessClosedNonIncomeExpenseRemainsIndependent(t *testing.T) {
	provider := New()
	provider.Config = &Config{KeepRefundRecords: true}
	processed := provider.postProcess(&ir.IR{Orders: []ir.Order{
		refundOrder("order-1-refund", 10),
		originalOrder("order-1", 10),
		{
			Metadata: map[string]string{
				"status":  "交易关闭",
				"type":    "不计收支",
				"orderId": "closed-1",
			},
		},
	}})

	if got := len(processed.Orders); got != 2 {
		t.Fatalf("expected only closed transaction to be removed, got %d orders", got)
	}
}

func refundPairIR() *ir.IR {
	return &ir.IR{Orders: []ir.Order{
		refundOrder("order-1-refund", 10),
		originalOrder("order-1", 10),
	}}
}

func refundOrder(orderID string, money float64) ir.Order {
	return ir.Order{
		Money:    money,
		Category: "退款",
		Type:     ir.TypeSend,
		Metadata: map[string]string{"status": "退款成功", "orderId": orderID},
	}
}

func originalOrder(orderID string, money float64) ir.Order {
	return ir.Order{
		Money:    money,
		Metadata: map[string]string{"status": "交易成功", "orderId": orderID},
	}
}
