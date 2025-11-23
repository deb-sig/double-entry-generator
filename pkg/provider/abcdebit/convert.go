package abcdebit

import "github.com/deb-sig/double-entry-generator/v2/pkg/ir"

func (ad *AbcDebit) convertToIR() *ir.IR {
	i := ir.New()
	for _, order := range ad.Orders {
		if order.Type == OrderTypeUnknown {
			continue
		}
		irOrder := ir.Order{
			OrderType:      ir.OrderTypeNormal,
			Peer:           order.Peer,
			Item:           normalizeItem(order.Summary, order.Postscript),
			PayTime:        order.PayTime,
			Type:           convertType(order.Type),
			TypeOriginal:   string(order.Type),
			TxTypeOriginal: order.Summary,
			Money:          order.Amount,
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
		if order.RawAmount != "" {
			metadata["amountRaw"] = order.RawAmount
		}
		irOrder.Metadata = metadata

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
