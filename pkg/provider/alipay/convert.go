package alipay

import (
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// convertToIR convert alipay bills to IR.
func (a *Alipay) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range a.Orders {

		irO := ir.Order{
			Peer:         o.Peer,
			Item:         o.ItemName,
			Category:     o.Category,
			Method:       o.Method,
			PayTime:      o.PayTime,
			Money:        o.Money,
			OrderID:      &o.DealNo,
			Type:         convertType(o.Type),
			TypeOriginal: o.TypeOriginal,
		}
		irO.Metadata = getMetadata(o)
		if o.MerchantId != "" {
			irO.MerchantOrderID = &o.MerchantId
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func convertType(t Type) ir.Type {
	switch t {
	case TypeSend:
		return ir.TypeSend
	case TypeRecv:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}

// getMetadata get the metadata (e.g. status, method, category and so on.)
//  from order.
func getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "支付宝"
	var status, method, category, typeOriginal, orderId, merchantId, paytime string

	paytime = o.PayTime.Format(localTimeFmt)

	if o.DealNo != "" {
		orderId = o.DealNo
	}

	if o.MerchantId != "" {
		merchantId = o.MerchantId
	}

	if o.Category != "" {
		category = o.Category
	}

	if o.TypeOriginal != "" {
		typeOriginal = o.TypeOriginal
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
		"type":       typeOriginal,
		"category":   category,
		"method":     method,
		"status":     status,
	}
}
