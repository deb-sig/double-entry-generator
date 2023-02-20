package htsec

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (h *Htsec) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType:      ir.OrderTypeSecuritiesTrade,
			Peer:           "htsec",
			PayTime:        o.TransactionTime,
			TxTypeOriginal: o.TxTypeOriginal,
			Type:           convertType(o.Type),
			TypeOriginal:   string(o.Type),
			Item:           o.SecuritiesName,
			Money:          o.TransactionAmount,
			Amount:         float64(o.Volume),
			Price:          o.Price,
			Commission:     o.Commission,
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func convertType(t OrderType) ir.Type {
	switch t {
	case TxTypeBuy:
		return ir.TypeSend
	case TxTypeSell:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}
