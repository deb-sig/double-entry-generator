package alipay

import (
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

// convertToIR convert alipay bills to IR.
func (a *Alipay) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range a.Orders {

		irO := ir.Order{
			Peer:           o.Peer,
			Item:           o.ItemName,
			Method:         o.Method,
			PayTime:        o.PayTime,
			Money:          o.Money,
			OrderID:        &o.DealNo,
			TxType:         conevertType(o.TxType),
			TxTypeOriginal: o.TxTypeOriginal,
		}
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
