package citic

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (citic *Citic) TranslateToOrders(arr []string) error {
	// trim strings
	for idx, a := range arr {
		a = strings.TrimSpace(a)
		arr[idx] = a
	}
	var bill Order
	var err error

	bill.TradeTime, err = time.Parse(localTimeFmt, arr[0]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse trade time %s error: %v", arr[0], err)
	}

	bill.PostTime, err = time.Parse(localTimeFmt, arr[1]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse trade time %s error: %v", arr[0], err)
	}

	bill.TradeDesc = arr[2]
	bill.Method = arr[3]
	bill.Currency = arr[5]

	bill.Type = getOrderTypeByTransactionAmount(arr[6])

	bill.Money, err = strconv.ParseFloat(strings.TrimLeft(arr[6], "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", arr[6], err)
	}

	citic.Orders = append(citic.Orders, bill)
	return nil
}
