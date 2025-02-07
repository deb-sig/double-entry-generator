package citic

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (h *Citic) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
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
