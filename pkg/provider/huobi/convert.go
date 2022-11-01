package huobi

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (h *Huobi) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType:      ir.OrderTypeHuobiTrade,
			Peer:           "Huobi",
			PayTime:        o.PayTime,
			TxTypeOriginal: o.TxTypeOriginal,
			Type:           convertType(o.Type),
			TypeOriginal:   string(o.Type),
			Item:           o.Item,
			Money:          o.Money,
			Amount:         o.Amount,
			Price:          o.Price,
			Commission:     o.Commission,
			Units: map[ir.Unit]string{
				ir.BaseUnit:       o.BaseUnit,
				ir.TargetUnit:     o.TargetUnit,
				ir.CommissionUnit: o.CommissionUnit,
			},
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func convertType(t OrderType) ir.Type {
	switch t {
	case TypeBuy:
		return ir.TypeSend
	case TypeSell:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}
