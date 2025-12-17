package abc_debit

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

	order := Order{
		TradeDate:  safeAccess(record, 0),
		TradeTime:  safeAccess(record, 1),
		Summary:    safeAccess(record, 2),
		Amount:     safeAccess(record, 3),
		Balance:    safeAccess(record, 4),
		Peer:       safeAccess(record, 5),
		LogNumber:  safeAccess(record, 6),
		Channel:    safeAccess(record, 7),
		Postscript: safeAccess(record, 8),
	}

	ad.Orders = append(ad.Orders, order)

	return nil
}

func safeAccess(arr []string, idx int) string {
	if idx >= 0 && idx < len(arr) {
		return arr[idx]
	}
	return ""
}
