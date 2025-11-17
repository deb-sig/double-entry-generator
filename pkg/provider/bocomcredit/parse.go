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

	txnCurrency, txnAmount, err := splitCurrencyAmount(record[3])
	if err != nil {
		return err
	}

	settleCurrency, settleAmount, err := splitCurrencyAmount(record[4])
	if err != nil {
		return err
	}

	description := record[2]
	typeOriginal, _ := splitDescription(description)
	if typeOriginal == "" {
		return fmt.Errorf("missing transaction type in description: %s", description)
	}

	orderType, err := inferOrderType(typeOriginal)
	if err != nil {
		return fmt.Errorf("infer order type from %q: %w", description, err)
	}

	order := Order{
		TradeDate:      tradeDate,
		RecordDate:     recordDate,
		Description:    description,
		Amount:         settleAmount,
		Currency:       settleCurrency,
		TxnAmount:      txnAmount,
		TxnCurrency:    txnCurrency,
		TxnAmountRaw:   record[3],
		Type:           orderType,
		TypeOriginal:   typeOriginal,
		TxTypeOriginal: typeOriginal,
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
		bc.Statistics.TotalInMoney += settleAmount
	case OrderTypeSend:
		bc.Statistics.TotalOutRecords++
		bc.Statistics.TotalOutMoney += settleAmount
	}

	return nil
}
