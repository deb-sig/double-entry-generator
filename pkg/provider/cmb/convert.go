package cmb

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (cmb *Cmb) convertDebitToIR() *ir.IR {
	i := ir.New()
	for _, o := range cmb.DebitOrders {
		irO := ir.Order{
			Peer:           o.CounterParty,
			Item:           o.CustomerType,
			TxTypeOriginal: o.TransactionType,
			Money:          o.TransactionAmount,
			PayTime:        o.Date,
			Type:           convertType(o.Type),
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func (cmb *Cmb) convertCreditToIR() *ir.IR {
	i := ir.New()
	for _, o := range cmb.CreditOrders {
		irO := ir.Order{
			Peer:    "CMB",
			Item:    o.Description,
			Money:   o.RmbAmount,
			Method:  o.CardNo,
			PayTime: o.PostedDate,
			Type:    convertType(o.Type),
		}

		// 有交易日优先用交易日作为交易时间
		if o.SoldDate != nil {
			irO.PayTime = *o.SoldDate
		}

		i.Orders = append(i.Orders, irO)
	}
	return i
}

func convertType(t OrderType) ir.Type {
	switch t {
	case OrderTypeSend:
		return ir.TypeSend
	case OrderTypeRecv:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}
