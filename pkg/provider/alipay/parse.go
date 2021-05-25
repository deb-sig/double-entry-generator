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
func (a *Alipay) translateToOrders(array []string) error {
	var err error
	a.LineNum++

	if len(array) != 17 {
		return fmt.Errorf("Length mismatch: Expected 17, got %d", len(array))
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
	bill.DealNo = array[0]
	bill.OrderNo = array[1]
	bill.CreateTime, err = time.Parse(LocalTimeFmt, array[2]+" +0800")
	if err != nil {
		log.Println("parse create time error:", array[2], err)
		return err
	}
	if array[3] != "" {
		bill.PayTime, err = time.Parse(LocalTimeFmt, array[3]+" +0800")
		if err != nil {
			log.Println("parse paytime error:", array[3], err, array)
			return err
		}
	}
	bill.LastUpdate, err = time.Parse(LocalTimeFmt, array[4]+" +0800")
	if err != nil {
		log.Println("parse last update error:", array[4], err)
		return err
	}
	bill.DealSrc = array[5]
	bill.Type = array[6]
	bill.Peer = array[7]
	bill.ItemName = array[8]
	bill.Money, err = strconv.ParseFloat(array[9], 32)
	if err != nil {
		log.Println("parse money error:", array[9], err)
		return err
	}
	bill.TxType = getTxType(array[10])
	if bill.TxType == TxTypeNil {
		log.Println("get tx type error:", array[10], array)
		return fmt.Errorf("Failed to get the tx type %s", array[10])
	}
	bill.Status = array[11]
	if bill.Status == "交易关闭" {
		log.Printf("Line %d: There is a mole, The tx is canceled.", a.LineNum)
	}
	if bill.Status == "退款成功" {
		log.Printf("Lind %d: There has a refund transaction.", a.LineNum)
	}
	bill.ServiceFee, err = strconv.ParseFloat(array[12], 32)
	if err != nil {
		log.Println("parse service fee error:", array[12], err)
		return err
	}
	bill.Refund, err = strconv.ParseFloat(array[13], 32)
	if err != nil {
		log.Println("parse refund error:", array[13], err)
		return err
	}
	bill.Comment = array[14]
	bill.MoneyStatus = getMoneyStatus(array[15])
	if bill.MoneyStatus == MoneyStatusNil {
		return fmt.Errorf("Failed to get the money status: %s", array[15])
	}
	a.Orders = append(a.Orders, bill)
	return nil
}
