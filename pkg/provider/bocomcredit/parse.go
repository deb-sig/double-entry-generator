package bocomcredit

import (
	"fmt"
	"strings"
)

func (bc *BocomCredit) translateToOrders(record []string) error {
	if len(record) < 5 {
		return fmt.Errorf("invalid record length: %d", len(record))
	}

	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	if record[0] == "" {
		return nil
	}

	tradeDate, err := parseDate(record[0])
	if err != nil {
		return fmt.Errorf("parse trade date error: %w", err)
	}
	recordDate, err := parseDate(record[1])
	if err != nil {
		return fmt.Errorf("parse record date error: %w", err)
	}

	currency, amount, err := splitCurrencyAmount(record[3])
	if err != nil {
		return err
	}

	description := record[2]
	orderType := inferOrderType(description)

	order := Order{
		TradeDate:    tradeDate,
		RecordDate:   recordDate,
		Description:  description,
		Amount:       amount,
		Currency:     currency,
		Type:         orderType,
		TypeOriginal: description,
	}

	bc.Orders = append(bc.Orders, order)
	bc.Statistics.ParsedItems++
	if bc.Statistics.Start.IsZero() || tradeDate.Before(bc.Statistics.Start) {
		bc.Statistics.Start = tradeDate
	}
	if bc.Statistics.End.IsZero() || tradeDate.After(bc.Statistics.End) {
		bc.Statistics.End = tradeDate
	}

	switch order.Type {
	case OrderTypeRecv:
		bc.Statistics.TotalInRecords++
		bc.Statistics.TotalInMoney += amount
	case OrderTypeSend:
		bc.Statistics.TotalOutRecords++
		bc.Statistics.TotalOutMoney += amount
	}

	return nil
}
