package boc

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func (boc *Boc) TranslateToDebitOrders(arr []string) error {
	var bill Order
	var err error
	for idx, a := range arr {
		arr[idx] = strings.TrimSpace(a)
	}
	bill.PayTime, err = time.Parse(localTimeFmt, arr[0]+" "+arr[1]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse pay time error: %v", err)
	}
	bill.Currency = "CNY"
	if arr[2] != "人民币" {
		bill.Currency = "USD"
	}
	var moneyNum = strings.ReplaceAll(arr[3], ",", "")
	bill.Money, err = strconv.ParseFloat(strings.TrimLeft(moneyNum, "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", moneyNum, err)
	}
	bill.Type = getOrderTypeByTransactionAmount(arr[3])
	bill.ItemName = arr[5]
	bill.Channel = arr[6]
	bill.Branch = arr[7]
	bill.Postscript = arr[8]
	bill.PeerName = arr[9]
	bill.PeerCard = arr[10]
	bill.PeerBank = arr[11]
	bill.Method = arr[12]
	boc.Orders = append(boc.Orders, bill)

	return nil
}

func (boc *Boc) TranslateToCreditOrders(arr []string) error {
	var bill Order
	var err error
	for idx, a := range arr {
		arr[idx] = strings.TrimSpace(a)
	}
	bill.PayTime, err = time.Parse(localTimeFmt, arr[1]+" "+"20:00:00"+" +0800 CST")
	if err != nil {
		log.Println("parse pay time error:", arr[1], err)
		return err
	}
	bill.Method = arr[2]
	bill.ItemName = arr[3]
	if arr[4] == "" { 
		bill.Type = TypeSend
		bill.Money, err = strconv.ParseFloat(strings.TrimLeft(arr[5], "-"), 64)
	} else {
		bill.Type = TypeRecv
		bill.Money, err = strconv.ParseFloat(strings.TrimLeft(arr[4], "-"), 64)
	}

	if err != nil {
		return fmt.Errorf("parse money %s error: %v", arr[3], err)
	}
	bill.Currency = arr[6]
	boc.Orders = append(boc.Orders, bill)
	return nil
}
