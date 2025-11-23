package abcdebit

import (
	"fmt"
	"strings"
)

func (ad *AbcDebit) translateToOrders(record []string) error {
	if len(record) < 4 {
		return fmt.Errorf("invalid record length: %d", len(record))
	}

	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	if record[0] == "" {
		return nil
	}

	payTime, err := parseTradeTime(record[0], safeAccess(record, 1))
	if err != nil {
		return err
	}

	amount, txType, err := parseAmountAndType(record[3])
	if err != nil {
		return err
	}

	order := Order{
		PayTime:    payTime,
		Summary:    safeAccess(record, 2),
		Postscript: safeAccess(record, 8),
		Amount:     amount,
		Balance:    safeAccess(record, 4),
		Peer:       normalizePeer(safeAccess(record, 5)),
		Channel:    safeAccess(record, 7),
		LogNumber:  safeAccess(record, 6),
		Type:       txType,
		RawAmount:  record[3],
	}

	ad.Orders = append(ad.Orders, order)
	ad.Statistics.ParsedItems++
	if ad.Statistics.Start.IsZero() || payTime.Before(ad.Statistics.Start) {
		ad.Statistics.Start = payTime
	}
	if ad.Statistics.End.IsZero() || payTime.After(ad.Statistics.End) {
		ad.Statistics.End = payTime
	}

	switch txType {
	case OrderTypeRecv:
		ad.Statistics.TotalInRecords++
		ad.Statistics.TotalInMoney += amount
	case OrderTypeSend:
		ad.Statistics.TotalOutRecords++
		ad.Statistics.TotalOutMoney += amount
	}

	return nil
}

func safeAccess(arr []string, idx int) string {
	if idx >= 0 && idx < len(arr) {
		return arr[idx]
	}
	return ""
}
