package mt

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertToIR convert mt bills to IR.
func (mt *MT) convertToIR() *ir.IR {
	// ir 是 package
	i := ir.New()
	for _, o := range mt.Orders {
		irO := ir.Order{
			PayTime:         o.PayTime,
			TypeOriginal:    o.TypeOriginal,
			Item:            o.ItemName,
			Type:            convertType(o.Type),
			Method:          o.Method,
			Money:           o.Money,
			OrderID:         &o.DealNo,
			MerchantOrderID: &o.MerchantId,
		}
		irO.Metadata = getMetadata(o)
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
//
//	from order.

func getMetadata(o Order) map[string]string {
	source := "美团"
	var method, typeOriginal, orderId, merchantId, paytime string

	paytime = o.PayTime.Format(localTimeFmt)

	orderId = o.DealNo
	merchantId = o.MerchantId
	typeOriginal = o.TypeOriginal
	method = o.Method

	return map[string]string{
		"source":     source,
		"payTime":    paytime,
		"orderId":    orderId,
		"merchantId": merchantId,
		"type":       typeOriginal,
		"method":     method,
	}
}
