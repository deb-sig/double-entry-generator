package td

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (td *Td) convertToIR() *ir.IR {
	itermediateRepresentation := ir.New()
	for _, order := range td.Orders {

		irO := ir.Order{
			Peer:         "TD",
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
