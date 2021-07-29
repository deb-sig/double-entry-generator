package wechat

import (
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// convertToIR convert wechat bills to IR.
func (w *Wechat) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range w.Orders {
		irO := ir.Order{
			Peer:           o.Peer,
			Item:           o.Item,
			PayTime:        o.PayTime,
			Money:          o.Money,
			OrderID:        &o.OrderID,
			TxType:         conevertType(o.Type),
			TxTypeOriginal: o.TypeOriginal,
			TypeOriginal:   o.TxTypeOriginal,
			Method:         o.Method,
			Commission:     o.Commission,
		}
		irO.Metadata = getMetadata(o)
		if o.MechantOrderID != "" {
			irO.MerchantOrderID = &o.MechantOrderID
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func conevertType(t OrderType) ir.TxType {
	switch t {
	case OrderTypeSend:
		return ir.TxTypeSend
	case OrderTypeRecv:
		return ir.TxTypeRecv
	default:
		return ir.TxTypeUnknown
	}
}

func getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "微信支付"
	var status, method, category, txType, typeOriginal, orderId, merchantId, paytime string

	paytime = o.PayTime.String()

	if o.OrderID != "" {
		orderId = o.OrderID
	}

	if o.MechantOrderID != "" {
		merchantId = o.MechantOrderID
	}

	if o.TypeOriginal != "" {
		txType = o.TypeOriginal
	}

	if o.TxTypeOriginal != "" {
		typeOriginal = o.TxTypeOriginal
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
		"type":       typeOriginal,
		"category":   category,
		"method":     method,
		"status":     status,
	}
}
