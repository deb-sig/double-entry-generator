package wechat

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv file to []Order.
func (w *Wechat) translateToOrders(array []string) error {
	for idx, a := range array {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		array[idx] = a
	}
	var bill Order
	var err error
	bill.PayTime, err = time.Parse(LocalTimeFmt, array[0]+" +0800")
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", array[0], err)
	}

	bill.TxType = getTxType(array[1])
	if bill.TxType == TxTypeUnknown {
		return fmt.Errorf("Failed to get the tx type %s: %v", array[1], err)
	}
	bill.Peer = array[2]
	bill.Item = array[3]
	bill.Type = getOrderType(array[4])
	bill.TypeOriginal = array[4]
	if bill.Type == OrderTypeUnknown {
		return fmt.Errorf("Failed to get the order type %s: %v", array[4], err)
	}
	bill.Money, err = strconv.ParseFloat(array[5][2:], 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", array[5], err)
	}
	bill.Method = array[6]
	bill.OrderID = array[8]
	bill.MechantOrderID = array[9]

	w.Orders = append(w.Orders, bill)
	return nil
}
