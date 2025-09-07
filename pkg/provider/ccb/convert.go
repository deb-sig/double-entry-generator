package ccb

import (
	"strconv"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertToIR convert CCB bills to IR.
func (ccb *CCB) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range ccb.Orders {
		if o.Type == OrderTypeUnknown {
			continue
		}
		irO := ir.Order{
			Peer:           o.Peer,
			Item:           o.Item,
			PayTime:        o.PayTime,
			Money:          o.Money,
			Type:           convertType(o.Type),
			TypeOriginal:   string(o.Type),
			TxTypeOriginal: o.TxTypeOriginal,
			Currency:       ccb.Currency,
			Note:           o.TxTypeOriginal,
		}
		irO.Metadata = ccb.getMetadata(o)

		// send
		if o.Type == OrderTypeSend {
			irO.MinusAccount = "Assets:CCB:Card"
			irO.PlusAccount = "Expenses:Default"
		} else { // recv
			irO.MinusAccount = "Income:Default"
			irO.PlusAccount = "Assets:CCB:Card"
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

// getMetadata get the metadata (e.g. status, method, category and so on.)
//
//	from order.
func (ccb *CCB) getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "中国建设银行"
	var txTypeOriginal, guessedType, currency, balances, peerAccount, peerAccountNum, region, tradeTime, recordDate, expense, income string

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

	if o.PeerAccountNum != "" {
		peerAccountNum = o.PeerAccountNum
	}

	if o.Region != "" {
		region = o.Region
	}

	if o.TradeTime != "" {
		tradeTime = o.TradeTime
	}

	if o.RecordDate != "" {
		recordDate = o.RecordDate
	}

	if o.Expense != 0 {
		expense = strconv.FormatFloat(o.Expense, 'G', -1, 64)
	}

	if o.Income != 0 {
		income = strconv.FormatFloat(o.Income, 'G', -1, 64)
	}

	metadata := map[string]string{
		"source":         source,
		"txType":         txTypeOriginal,
		"type":           guessedType,
		"currency":       currency,
		"balances":       balances,
		"peerAccount":    peerAccount,
		"peerAccountNum": peerAccountNum,
		"region":         region,
		"tradeTime":      tradeTime,
		"recordDate":     recordDate,
		"expense":        expense,
		"income":         income,
	}

	if ccb.AccountNum != "" {
		metadata["accountNum"] = ccb.AccountNum
	}

	return metadata
}
