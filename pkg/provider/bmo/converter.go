package bmo

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (bmo *Bmo) convertToIR() *ir.IR {
	itermediateRepresentation := ir.New()
	for _, order := range bmo.Orders {

		irO := ir.Order{
			Peer:         "BMO",
			Item:         order.TransactionDescription,
			PayTime:      order.PayTime,
			Type:         convertType(order.Type),
			TypeOriginal: string(order.Type),
			Money:        order.Money,
		}
		itermediateRepresentation.Orders = append(itermediateRepresentation.Orders, irO)
	}
	return itermediateRepresentation
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
