package bocomcredit

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func (bc *BocomCredit) convertToIR() *ir.IR {
	i := ir.New()
	for _, order := range bc.Orders {
		if order.Type == OrderTypeUnknown {
			continue
		}
		irOrder := ir.Order{
			OrderType:      ir.OrderTypeNormal,
			Peer:           providerPeer,
			Item:           order.Description,
			PayTime:        order.TradeDate,
			Type:           convertType(order.Type),
			TypeOriginal:   order.TypeOriginal,
			TxTypeOriginal: order.TxTypeOriginal,
			Money:          order.Amount,
			Currency:       order.Currency,
		}
		metadata := map[string]string{
			"source":     "交通银行信用卡",
			"recordDate": order.RecordDate.Format(dateLayout),
		}
		if original := originalTransactionAmount(order); original != "" {
			metadata["transactionAmount"] = original
		}
		irOrder.Metadata = metadata
		i.Orders = append(i.Orders, irOrder)
	}
	return i
}

func originalTransactionAmount(order Order) string {
	if order.TxnCurrency == "" {
		return ""
	}
	if order.TxnCurrency == order.Currency && order.TxnAmount == order.Amount {
		return ""
	}
	if order.TxnAmountRaw != "" {
		return order.TxnAmountRaw
	}
	return fmt.Sprintf("%s %.2f", order.TxnCurrency, order.TxnAmount)
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
