package icbc

import (
	"strconv"

	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// convertToIR convert ICBC bills to IR.
func (icbc *Icbc) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range icbc.Orders {
		irO := ir.Order{
			Peer:           o.Peer,
			Money:          o.Money,
			PayTime:        o.PayTime,
			Type:           convertType(o.Type),
			TypeOriginal:   string(o.Type),
			TxTypeOriginal: o.TxTypeOriginal,
		}
		irO.Metadata = icbc.getMetadata(o)
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

// getMetadata get the metadata (e.g. status, method, category and so on.)
//  from order.
func (icbc *Icbc) getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "中国工商银行"
	var txTypeOriginal, guessedType, currency, balances, peerAccount string

	if o.TxTypeOriginal != "" {
		txTypeOriginal = o.TxTypeOriginal
	}

	if o.Type != "" {
		guessedType = string(o.Type)
	}

	if o.Currency != "" {
		currency = o.Currency
	}

	if o.Balances != 0 {
		balances = strconv.FormatFloat(o.Balances, 'G', -1, 64)
	}

	if o.PeerAccountName != "" {
		peerAccount = o.PeerAccountName
	}

	metadata := map[string]string{
		"source":      source,
		"txType":      txTypeOriginal,
		"type":        guessedType,
		"currency":    currency,
		"balances":    balances,
		"peerAccount": peerAccount,
	}

	if icbc.CardName != "" {
		metadata["cardName"] = icbc.CardName
	}

	return metadata
}
