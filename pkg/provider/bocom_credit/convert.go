package bocomcredit

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (bc *BocomCredit) convertToIR() *ir.IR {
	i := ir.New()
	for _, order := range bc.Orders {
		if order.Type == OrderTypeUnknown {
			continue
		}
		irOrder := ir.Order{
			OrderType:      ir.OrderTypeNormal,
			Peer:           providerPeer,
			Item:           order.Description,
			PayTime:        order.TradeDate,
			Type:           convertType(order.Type),
			TypeOriginal:   order.TypeOriginal,
			TxTypeOriginal: order.TxTypeOriginal,
			Money:          order.Amount,
			Currency:       order.Currency,
		}
		irOrder.Metadata = map[string]string{
			"source":     "交通银行信用卡",
			"recordDate": order.RecordDate.Format(dateLayout),
		}
		i.Orders = append(i.Orders, irOrder)
	}
	return i
}

func convertType(t OrderType) ir.Type {
	switch t {
	case OrderTypeRecv:
		return ir.TypeRecv
	case OrderTypeSend:
		return ir.TypeSend
	default:
		return ir.TypeUnknown
	}
}
