package bocom_credit

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertRawRecord converts a RawRecord to an Order with full business logic:
// date parsing, amount parsing, type inference, etc.
func (bc *BocomCredit) convertRawRecord(raw *RawRecord) (*Order, error) {
	tradeDate, err := parseDate(raw.TradeDate)
	if err != nil {
		return nil, fmt.Errorf("parse trade date error: %w", err)
	}

	recordDate, err := parseDate(raw.RecordDate)
	if err != nil {
		return nil, fmt.Errorf("parse record date error: %w", err)
	}

	txnCurrency, txnAmount, err := splitCurrencyAmount(raw.TxnCurrencyAmount)
	if err != nil {
		return nil, err
	}

	settleCurrency, settleAmount, err := splitCurrencyAmount(raw.SettleCurrencyAmount)
	if err != nil {
		return nil, err
	}

	typeOriginal, _ := splitDescription(raw.TradeDescription)
	if typeOriginal == "" {
		return nil, fmt.Errorf("missing transaction type in description: %s", raw.TradeDescription)
	}

	orderType, err := inferOrderType(typeOriginal)
	if err != nil {
		return nil, fmt.Errorf("infer order type from %q: %w", raw.TradeDescription, err)
	}

	return &Order{
		TradeDate:      tradeDate,
		RecordDate:     recordDate,
		Description:    raw.TradeDescription,
		Amount:         settleAmount,
		Currency:       settleCurrency,
		TxnAmount:      txnAmount,
		TxnCurrency:    txnCurrency,
		TxnAmountRaw:   raw.TxnCurrencyAmount,
		Type:           orderType,
		TypeOriginal:   typeOriginal,
		TxTypeOriginal: typeOriginal,
	}, nil
}

// updateStatistics updates the statistics based on a converted order.
func (bc *BocomCredit) updateStatistics(order *Order) {
	bc.Statistics.ParsedItems++

	if bc.Statistics.Start.IsZero() || order.TradeDate.Before(bc.Statistics.Start) {
		bc.Statistics.Start = order.TradeDate
	}
	if bc.Statistics.End.IsZero() || order.TradeDate.After(bc.Statistics.End) {
		bc.Statistics.End = order.TradeDate
	}

	switch order.Type {
	case OrderTypeRecv:
		bc.Statistics.TotalInRecords++
		bc.Statistics.TotalInMoney += order.Amount
	case OrderTypeSend:
		bc.Statistics.TotalOutRecords++
		bc.Statistics.TotalOutMoney += order.Amount
	}
}

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
