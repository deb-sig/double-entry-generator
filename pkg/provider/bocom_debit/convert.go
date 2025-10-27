package bocom_debit

import (
	"strconv"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertToIR converts parsed Bocom orders into the intermediate representation.
func (b *Bocom) convertToIR() *ir.IR {
	irOrders := ir.New()
	for _, o := range b.Orders {
		payTime, _ := parsePayTime(o)
		orderType := determineOrderType(o.DcFlg)
		irOrder := ir.Order{
			Peer:           buildPeer(o.PaymentReceiptAccountName, o.PaymentReceiptAccount),
			Item:           buildItem(o.TradingPlace, o.Abstract),
			Money:          o.TransAmt,
			PayTime:        payTime,
			Type:           convertType(orderType),
			TypeOriginal:   o.DcFlg,
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
		"dcFlg":                     o.DcFlg,
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

	if payTime, err := parsePayTime(o); err == nil && !payTime.IsZero() {
		metadata["payTime"] = payTime.Format(timeLayout)
	}

	return metadata
}
