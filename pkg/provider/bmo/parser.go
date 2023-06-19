package bmo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translate Debit Card csv file to []Order.
func (bmo *Bmo) translateDebitToOrders(columns []string) error {
	for idx, record := range columns {
		record = strings.Trim(record, " ")
		record = strings.Trim(record, "\t")
		columns[idx] = record
	}

	var bill Order
	var err error
	bill.PayTime, err = time.Parse(LocalTimeFmt, columns[2]+" -0700")
	if err != nil {
		return fmt.Errorf("parse Pay time %s error: %v", columns[2], err)
	}

	bill.TransactionDescription = columns[4]
	var transactionType = columns[1]
	bill.Type = getOrderType(transactionType)
	var amount = columns[3]
	bill.Money, err = strconv.ParseFloat(strings.TrimLeft(amount, "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", columns[3], err)
	}

	bmo.Orders = append(bmo.Orders, bill)
	return nil
}

// translate Credit Card csv file to []Order.
func (bmo *Bmo) translateCreditToOrders(columns []string) error {
	for idx, record := range columns {
		record = strings.Trim(record, " ")
		record = strings.Trim(record, "\t")
		columns[idx] = record
	}

	var bill Order
	var err error
	bill.PayTime, err = time.Parse(LocalTimeFmt, columns[2]+" -0700")
	if err != nil {
		return fmt.Errorf("parse Pay time %s error: %v", columns[2], err)
	}

	bill.TransactionDescription = columns[5]
	var amount = columns[4]
	bill.Type = getOrderTypeByTransactionAmount(amount)
	bill.Money, err = strconv.ParseFloat(strings.TrimLeft(amount, "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", amount, err)
	}

	bmo.Orders = append(bmo.Orders, bill)
	return nil
}
