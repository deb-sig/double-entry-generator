package wechat

import (
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// convertToIR convert wechat bills to IR.
func (w *Wechat) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range w.Orders {
		irO := ir.Order{
			Peer:         o.Peer,
			Item:         o.Item,
			PayTime:      o.PayTime,
			Money:        o.Money,
			OrderID:      &o.OrderID,
			TxType:       conevertType(o.Type),
			TypeOriginal: o.TypeOriginal,
			Method:       o.Method,
		}
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
