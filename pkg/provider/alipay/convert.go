package alipay

import (
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// convertToIR convert alipay bills to IR.
func (a *Alipay) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range a.Orders {

		irO := ir.Order{
			Peer:           o.Peer,
			Item:           o.ItemName,
			Category:       o.Category,
			Method:         o.Method,
			PayTime:        o.PayTime,
			Money:          o.Money,
			OrderID:        &o.DealNo,
			TxType:         conevertType(o.TxType),
			TxTypeOriginal: o.TxTypeOriginal,
		}
		irO.Metadata = getMetadata(o)
		if o.MerchantId != "" {
			irO.MerchantOrderID = &o.MerchantId
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func conevertType(t TxTypeType) ir.TxType {
	switch t {
	case TxTypeSend:
		return ir.TxTypeSend
	case TxTypeRecv:
		return ir.TxTypeRecv
	default:
		return ir.TxTypeUnknown
	}
}

func getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "支付宝"
	var status, method, category, txType, orderId, merchantId, paytime string

	paytime = o.PayTime.String()

	if o.DealNo != "" {
		orderId = o.DealNo
	}

	if o.MerchantId != "" {
		merchantId = o.MerchantId
	}

	if o.Category != "" {
		category = o.Category
	}

	if o.TxTypeOriginal != "" {
		txType = o.TxTypeOriginal
	}

	if o.Method != "" {
		method = o.Method
	}

	if o.Status != "" {
		status = o.Status
	}

	return map[string]string{
		"source":     source,
		"payTime":    paytime,
		"orderId":    orderId,
		"merchantId": merchantId,
		"txType":     txType,
		"category":   category,
		"method":     method,
		"status":     status,
	}
}
