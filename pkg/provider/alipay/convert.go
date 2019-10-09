package alipay

import (
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

// convertToIR convert alipay bills to IR.
func (a *Alipay) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range a.Orders {
		// Do not convert the freeze tx.
		if o.MoneyStatus == MoneyUnfreeze || o.MoneyStatus == MoneyFreeze {
			continue
		}

		irO := ir.Order{
			Peer:    o.Peer,
			Item:    o.ItemName,
			PayTime: o.CreateTime,
			Money:   o.Money,
			OrderID: &o.DealNo,
			Type:    conevertType(o.TxType),
		}
		if o.OrderNo != "" {
			irO.MerchantOrderID = &o.OrderNo
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
