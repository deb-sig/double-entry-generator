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

	for idx, a := range array {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		array[idx] = a
	}
	var bill Order
	bill.Type = getTxType(array[5])
	if bill.Type == TypeNil {
		log.Println("get tx type error:", array[5], array)
		return fmt.Errorf("failed to get the tx type %s", array[5])
	}
	bill.TypeOriginal = array[5]
	bill.Peer = array[2]
	bill.PeerAccount = array[3]
	bill.ItemName = array[4]
	bill.Method = array[7]
	bill.Category = array[1]
	bill.DealNo = array[9]
	bill.MerchantId = array[10]
	bill.Money, err = strconv.ParseFloat(array[6], 32)
	if err != nil {
		log.Println("parse money error:", array[6], err)
		return err
	}
	bill.Status = array[8]
	bill.PayTime, err = time.Parse(localTimeFmt, array[0]+" +0800 CST")
	if err != nil {
		log.Println("parse create time error:", array[0], err)
		return err
	}

	a.Orders = append(a.Orders, bill)
	return nil
}
