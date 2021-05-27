package huobi

import "github.com/gaocegege/double-entry-generator/pkg/ir"

func (h *Huobi) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType:      ir.OrderTypeHuobiTrade,
			Peer:           "Huobi",
			PayTime:        o.PayTime,
			TypeOriginal:   o.TypeOriginal,
			TxType:         convertType(o.TxType),
			TxTypeOriginal: string(o.TxType),
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

func convertType(t TxType) ir.TxType {
	switch t {
	case TxTypeBuy:
		return ir.TxTypeSend
	case TxTypeSell:
		return ir.TxTypeRecv
	default:
		return ir.TxTypeUnknown
	}
}
