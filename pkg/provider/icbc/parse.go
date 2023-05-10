package icbc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv file to []Order.
func (icbc *Icbc) translateToOrders(array []string) error {
	for idx, a := range array {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		array[idx] = a
	}
	var bill Order
	var err error
	bill.PayTime, err = time.Parse(localTimeFmt, strings.TrimSpace(array[1])+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", array[1], err)
	}

	bill.TxTypeOriginal = strings.TrimSpace(array[2])
	bill.Peer = strings.TrimSpace(array[3])
	bill.Region = strings.TrimSpace(array[4])

	a8 := strings.ReplaceAll(strings.TrimSpace(array[8]), ",", "")
	a9 := strings.ReplaceAll(strings.TrimSpace(array[9]), ",", "")
	if a8 == "" && a9 == "" {
		bill.Type = OrderTypeUnknown
	} else if a9 == "" {
		bill.Type = OrderTypeRecv
		bill.Money, err = strconv.ParseFloat(a8, 64)
	} else {
		bill.Type = OrderTypeSend
		bill.Money, err = strconv.ParseFloat(a9, 64)
	}
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", array[5], err)
	}

	bill.Currency = strings.TrimSpace(array[10])
	bill.Balances, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(array[11]), ",", ""), 64)
	bill.PeerAccountName = strings.TrimSpace(array[12])

	icbc.Orders = append(icbc.Orders, bill)
	return nil
}
