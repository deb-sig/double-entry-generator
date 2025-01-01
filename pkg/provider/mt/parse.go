package mt

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv file to []Order.
func (mt *MT) translateToOrders(array []string) error {
	var err error

	for idx, a := range array {
		array[idx] = strings.TrimSpace(a)
	}
	var bill Order
	bill.Type = getTxType(array[4])
	if bill.Type == TypeNil {
		log.Println("get tx type error:", array[4], array)
		return fmt.Errorf("Failed to get the tx type %s", array[4])
	}
	bill.TypeOriginal = array[4]
	bill.ItemName = array[3]
	bill.Method = array[5]
	bill.DealNo = array[8]
	bill.MerchantId = array[9]
	bill.Money, err = strconv.ParseFloat(array[7][2:], 32) // 去除 ¥ 符号，解析字符串为浮点数

	if err != nil {
		log.Println("parse money error:", array[7], err)
		return err
	}
	bill.PayTime, err = time.Parse(localTimeFmt, array[0]+" +0800 CST")
	if err != nil {
		log.Println("parse create time error:", array[1], err)
		return err
	}
	mt.Orders = append(mt.Orders, bill)
	return nil
}
