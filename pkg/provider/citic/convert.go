package citic

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (citic *Citic) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range citic.Orders {
		irO := ir.Order{
			Peer:    "CITIC",
			PayTime: o.TradeTime,
			Item:    o.TradeDesc,
			Method:  o.Method,
			Type:    convertType(o.Type),
			Money:   o.Money,
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
