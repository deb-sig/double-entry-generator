package cmb

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (cmb *Cmb) translateDebitToOrders(arr []string) error {
	// trim strings
	for idx, a := range arr {
		a = strings.TrimSpace(a)
		arr[idx] = a
	}
	var bill DebitOrder
	var err error

	bill.Date, err = time.Parse(localTimeFmt, arr[0]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse trade time %s error: %v", arr[0], err)
	}

	bill.Currency = arr[1]

	bill.TransactionAmount, err = strconv.ParseFloat(strings.TrimLeft(strings.ReplaceAll(arr[2], ",", ""), "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", arr[2], err)
	}

	bill.Balance, err = strconv.ParseFloat(strings.TrimLeft(strings.ReplaceAll(arr[3], ",", ""), "-"), 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", arr[3], err)
	}

	bill.TransactionType = safeAccessStrList(arr, 4)
	bill.CounterParty = safeAccessStrList(arr, 5)
	bill.CustomerType = safeAccessStrList(arr, 6)

	bill.Type = getDebitOrderTypeByTransactionAmount(arr[2])

	cmb.DebitOrders = append(cmb.DebitOrders, bill)
	return nil
}

func (cmb *Cmb) translateCreditToOrders(arr []string) error {
	// trim strings
	for idx, a := range arr {
		a = strings.TrimSpace(a)
		arr[idx] = a
	}
	var bill CreditOrder
	var err error

	if safeAccessStrList(arr, 0) != "" {
		t, err := convertCreditBillDate(safeAccessStrList(arr, 0), cmb.CreditBillYear, cmb.CreditBillMonth)
		bill.SoldDate = &t
		if err != nil {
			return fmt.Errorf("parse trade time %s error: %v", safeAccessStrList(arr, 0), err)
		}
	}

	if safeAccessStrList(arr, 1) != "" {
		bill.PostedDate, err = convertCreditBillDate(safeAccessStrList(arr, 1), cmb.CreditBillYear, cmb.CreditBillMonth)
		if err != nil {
			return fmt.Errorf("parse trade time %s error: %v", safeAccessStrList(arr, 1), err)
		}
	}

	bill.Description = safeAccessStrList(arr, 2)

	bill.RmbAmount, err = extractCreditAmount(safeAccessStrList(arr, 3))
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", safeAccessStrList(arr, 3), err)
	}

	bill.CardNo = safeAccessStrList(arr, 4)

	bill.OriginalTranAmount, err = extractCreditAmount(safeAccessStrList(arr, 5))
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", safeAccessStrList(arr, 5), err)
	}

	bill.Type = getCreditOrderTypeByTransactionAmount(safeAccessStrList(arr, 3))

	cmb.CreditOrders = append(cmb.CreditOrders, bill)
	return nil
}
