package alipay

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv file to []Order.
// Copyright © 2019 Sean at Shanghai
// Modified by TripleZ at Shenzhen(2021)
func (a *Alipay) translateToOrders(array []string) error {
	var err error
	a.LineNum++

	if len(array) != 12 {
		return fmt.Errorf("Length mismatch: Expected 12, got %d", len(array))
	}

	// Ignore the title row.
	if !a.TitleParsed {
		a.TitleParsed = true
		return nil
	}

	for idx, a := range array {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		array[idx] = a
	}
	var bill Order
	bill.TxType = getTxType(array[0])
	if bill.TxType == TxTypeNil {
		log.Println("get tx type error:", array[0], array)
		return fmt.Errorf("Failed to get the tx type %s", array[0])
	}
	bill.TxTypeOriginal = array[0]
	bill.Peer = array[1]
	bill.PeerAccount = array[2]
	bill.ItemName = array[3]
	bill.Method = array[4]
	bill.Money, err = strconv.ParseFloat(array[5], 32)
	if err != nil {
		log.Println("parse money error:", array[5], err)
		return err
	}
	bill.Status = array[6]
	if bill.Status == "交易关闭" {
		log.Printf("Line %d: There is a mole, The tx is canceled.", a.LineNum)
	}
	if bill.Status == "退款成功" {
		log.Printf("Lind %d: There has a refund transaction.", a.LineNum)
	}
	bill.Category = array[7]
	bill.DealNo = array[8]
	bill.MerchantId = array[9]
	bill.PayTime, err = time.Parse(LocalTimeFmt, array[10]+" +0800")
	if err != nil {
		log.Println("parse create time error:", array[10], err)
		return err
	}

	a.Orders = append(a.Orders, bill)
	return nil
}
