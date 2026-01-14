package abc_debit

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (ad *AbcDebit) convertToIR() (*ir.IR, error) {
	i := ir.New()
	for _, order := range ad.Orders {
		payTime, err := parseTradeTime(order.TradeDate, order.TradeTime)
		if err != nil {
			return nil, err
		}

		money, txType, err := parseMoneyAndType(order.Amount)
		if err != nil {
			return nil, err
		}

		irOrder := ir.Order{
			OrderType:      ir.OrderTypeNormal,
			Peer:           normalizePeer(order.Peer),
			Item:           normalizeItem(order.Summary, order.Postscript),
			PayTime:        payTime,
			Type:           convertType(txType),
			TypeOriginal:   string(txType),
			TxTypeOriginal: order.Summary,
			Money:          money,
			Currency:       defaultCurrency,
		}
		metadata := map[string]string{
			"source": providerSource,
		}
		if order.Postscript != "" {
			metadata["postscript"] = order.Postscript
		}
		if order.Channel != "" {
			metadata["channel"] = order.Channel
		}
		if order.LogNumber != "" {
			metadata["logNumber"] = order.LogNumber
		}
		if order.Balance != "" {
			metadata["balance"] = order.Balance
		}
		if order.Amount != "" {
			metadata["amount"] = order.Amount
		}
		irOrder.Metadata = metadata

		i.Orders = append(i.Orders, irOrder)
	}
	return i, nil
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
