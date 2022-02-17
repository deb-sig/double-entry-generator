package htsec

import "github.com/deb-sig/double-entry-generator/pkg/ir"

func (h *Htsec) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType:      ir.OrderTypeSecurityTrade,
			Peer:           "htsec",
			PayTime:        o.TransactionTime,
			TypeOriginal:   o.TypeOriginal,
			TxType:         convertType(o.TxType),
			TxTypeOriginal: string(o.TxType),
			Item:           o.SecurityName,
			Money:          o.TransactionAmount,
			Amount:         float64(o.Volume),
			Price:          o.Price,
			Commission:     o.Commission,
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
