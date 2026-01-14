package boc

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (boc *Boc) convertToIR() *ir.IR {
	itermediateRepresentation := ir.New()
	for _, order := range boc.Orders {

		irO := ir.Order{
			PayTime:      order.PayTime,
			Currency:     order.Currency,
			Money:        order.Money,
			Type:         convertType(order.Type),
			TypeOriginal: string(order.Type),
			Method:       order.Method,
			Item:         order.ItemName,
			Peer:         order.PeerName + order.PeerCard,
		}
		itermediateRepresentation.Orders = append(itermediateRepresentation.Orders, irO)
	}
	return itermediateRepresentation
}

func convertType(t OrderType) ir.Type {
	switch t {
	case TypeSend:
		return ir.TypeSend
	case TypeRecv:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}
