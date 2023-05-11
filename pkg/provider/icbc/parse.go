package icbc

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv file to []Order.
func (icbc *Icbc) translateToOrders(array []string) error {
	for idx, a := range array {
		a = strings.TrimSpace(a)
		array[idx] = a
	}

	if len(array) < 13 {
		log.Printf("ignore the invalid csv line: %+v\n", array)
		return nil
	}

	var bill Order
	var err error
	bill.PayTime, err = time.Parse(localTimeFmt, array[1]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", array[1], err)
	}

	bill.TxTypeOriginal = array[2]
	bill.Peer = array[3]
	bill.Region = array[4]

	a8 := strings.ReplaceAll(array[8], ",", "")
	a9 := strings.ReplaceAll(array[9], ",", "")
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

	bill.Currency = array[10]
	bill.Balances, _ = strconv.ParseFloat(strings.ReplaceAll(array[11], ",", ""), 64)
	bill.PeerAccountName = array[12]

	icbc.Orders = append(icbc.Orders, bill)
	return nil
}
