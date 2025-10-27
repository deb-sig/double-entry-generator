package bocom_debit

import (
	"strconv"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertToIR converts parsed Bocom orders into the intermediate representation.
func (b *Bocom) convertToIR() *ir.IR {
	irOrders := ir.New()
	for _, o := range b.Orders {
		irOrder := ir.Order{
			Peer:           o.Peer,
			Item:           o.Item,
			Money:          o.TransAmount,
			PayTime:        o.PayTime,
			Type:           convertType(o.Type),
			TypeOriginal:   o.DrCr,
			TxTypeOriginal: o.TradingType,
			Currency:       b.Currency,
			Note:           o.TradingType,
		}
		irOrder.Metadata = b.getMetadata(o)
		irOrders.Orders = append(irOrders.Orders, irOrder)
	}
	return irOrders
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

func (b *Bocom) getMetadata(o Order) map[string]string {
	metadata := map[string]string{
		"source":                    "交通银行",
		"serialNum":                 o.SerialNum,
		"drCr":                      o.DrCr,
		"tradingType":               o.TradingType,
		"tradingPlace":              o.TradingPlace,
		"abstract":                  o.Abstract,
		"paymentReceiptAccount":     o.PaymentReceiptAccount,
		"paymentReceiptAccountName": o.PaymentReceiptAccountName,
		"currency":                  b.Currency,
	}

	if o.Balance != 0 {
		metadata["balance"] = strconv.FormatFloat(o.Balance, 'f', -1, 64)
	}

	if o.TransDate != "" {
		metadata["transDate"] = o.TransDate
	}
	if o.TransTime != "" {
		metadata["transTime"] = o.TransTime
	}

	return metadata
}
