package cgb_credit

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertRawRecord 将广发信用卡原始字段转换为标准交易。
func (cc *CgbCredit) convertRawRecord(raw *RawRecord) (*Order, error) {
	tradeDate, err := parseDate(raw.TradeDate)
	if err != nil {
		return nil, fmt.Errorf("parse trade date error: %w", err)
	}

	recordDate, err := parseDate(raw.RecordDate)
	if err != nil {
		return nil, fmt.Errorf("parse record date error: %w", err)
	}

	tradeAmount, err := parseSignedAmount(raw.TradeAmount)
	if err != nil {
		return nil, fmt.Errorf("parse trade amount error: %w", err)
	}

	settleAmount, err := parseSignedAmount(raw.SettleAmount)
	if err != nil {
		return nil, fmt.Errorf("parse settle amount error: %w", err)
	}

	typeOriginal := extractType(raw.Description)
	if typeOriginal == "" {
		return nil, fmt.Errorf("missing transaction type in description: %s", raw.Description)
	}

	orderType, err := inferOrderType(settleAmount, typeOriginal)
	if err != nil {
		return nil, fmt.Errorf("infer order type from %q: %w", raw.Description, err)
	}

	return &Order{
		TradeDate:       tradeDate,
		RecordDate:      recordDate,
		Description:     raw.Description,
		Amount:          absAmount(settleAmount),
		Currency:        normalizeCurrency(raw.SettleCurrency),
		TradeAmount:     absAmount(tradeAmount),
		TradeCurrency:   normalizeCurrency(raw.TradeCurrency),
		TradeAmountRaw:  raw.TradeAmount,
		SettleAmountRaw: raw.SettleAmount,
		Type:            orderType,
		TypeOriginal:    typeOriginal,
		TxTypeOriginal:  typeOriginal,
	}, nil
}

func (cc *CgbCredit) updateStatistics(order *Order) {
	cc.Statistics.ParsedItems++

	if cc.Statistics.Start.IsZero() || order.TradeDate.Before(cc.Statistics.Start) {
		cc.Statistics.Start = order.TradeDate
	}
	if cc.Statistics.End.IsZero() || order.TradeDate.After(cc.Statistics.End) {
		cc.Statistics.End = order.TradeDate
	}

	switch order.Type {
	case OrderTypeRecv:
		cc.Statistics.TotalInRecords++
		cc.Statistics.TotalInMoney += order.Amount
	case OrderTypeSend:
		cc.Statistics.TotalOutRecords++
		cc.Statistics.TotalOutMoney += order.Amount
	}
}

func (cc *CgbCredit) convertToIR() *ir.IR {
	result := ir.New()
	for _, order := range cc.Orders {
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
			"source":     "广发银行信用卡",
			"recordDate": order.RecordDate.Format(dateLayout),
		}
		if original := originalTransactionAmount(order); original != "" {
			metadata["transactionAmount"] = original
		}
		irOrder.Metadata = metadata
		result.Orders = append(result.Orders, irOrder)
	}
	return result
}

func originalTransactionAmount(order Order) string {
	if order.TradeCurrency == "" {
		return ""
	}
	if order.TradeCurrency == order.Currency && order.TradeAmount == order.Amount {
		return ""
	}
	return fmt.Sprintf("%s %.2f", order.TradeCurrency, order.TradeAmount)
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
