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

	switch icbc.Mode {
	case CreditMode:
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
			return fmt.Errorf("parse money [%s,%s] error: %v", array[8], array[9], err)
		}

		bill.Currency = array[10]
		bill.Balances, _ = strconv.ParseFloat(strings.ReplaceAll(array[11], ",", ""), 64)
		bill.PeerAccountName = array[12]
	case DebitMode:
		var debitBillVersion int32
		if len(array) == 14 {
			debitBillVersion = 1
		} else if len(array) == 16 {
			debitBillVersion = 2
		} else {
			return fmt.Errorf("cannot recognize this debit bill format, len()=%v", len(array))
		}

		bill.PayTime, err = time.Parse(localTimeFmt, array[0]+" +0800 CST")
		if err != nil {
			return fmt.Errorf("parse create time %s error: %v", array[0], err)
		}
		bill.TxTypeOriginal = array[1]

		switch debitBillVersion {
		case 1:
			bill.Peer = array[2]
			bill.Region = array[3]

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
				return fmt.Errorf("parse money [%s,%s] error: %v", array[8], array[9], err)
			}
			bill.Currency = array[10]
			bill.Balances, _ = strconv.ParseFloat(strings.ReplaceAll(array[11], ",", ""), 64)
			bill.PeerAccountName = array[12]
		case 2:
			bill.Item = array[2]
			bill.Peer = array[3]
			bill.Region = array[4]

			_income := strings.ReplaceAll(array[9], ",", "")
			_expense := strings.ReplaceAll(array[10], ",", "")
			if _income == "" && _expense == "" {
				bill.Type = OrderTypeUnknown
			} else if _expense == "" {
				bill.Type = OrderTypeRecv
				bill.Money, err = strconv.ParseFloat(_income, 64)
			} else {
				bill.Type = OrderTypeSend
				bill.Money, err = strconv.ParseFloat(_expense, 64)
			}
			if err != nil {
				return fmt.Errorf("parse money [%s,%s] error: %v", array[8], array[9], err)
			}
			bill.Currency = array[11]
			bill.Balances, _ = strconv.ParseFloat(strings.ReplaceAll(array[12], ",", ""), 64)
			bill.PeerAccountName = array[13]
			bill.PeerAccountNum = array[14]
		}

	}

	if bill.Peer == "" {
		bill.Peer = bill.PeerAccountName
	} else if bill.PeerAccountName != "" {
		// both Peer & PeerAccountName are not empty
		bill.Peer = bill.Peer + " " + bill.PeerAccountName
	}

	icbc.Orders = append(icbc.Orders, bill)
	return nil
}
