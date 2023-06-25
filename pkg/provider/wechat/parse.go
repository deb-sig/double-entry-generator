package wechat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var commissionRegex *regexp.Regexp

func init() {
	commissionRegex, _ = regexp.Compile(`\d+\.\d{2}`)
}

// translateToOrders translates csv file to []Order.
func (w *Wechat) translateToOrders(array []string) error {
	for idx, a := range array {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		array[idx] = a
	}
	var bill Order
	var err error
	bill.PayTime, err = time.Parse(localTimeFmt, array[0]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", array[0], err)
	}

	bill.TxType = getTxType(array[1])
	switch bill.TxType {
	case TxTypeCash2Cash:
		fallthrough
	case TxTypeCash2CashLooseChange:
		log.Printf("Get an unusable tx type, ignore it: %s\n", bill.TxType)
		return nil
	case TxTypeUnknown:
		return fmt.Errorf("Failed to get the tx type %s: %v", array[1], err)
	}
	bill.TxTypeOriginal = array[1]
	bill.Peer = array[2]
	bill.Item = array[3]
	bill.Type = getOrderType(array[4])
	bill.TypeOriginal = array[4]
	if bill.Type == OrderTypeUnknown {
		return fmt.Errorf("Failed to get the order type %s: %v", array[4], err)
	}
	// deal with the withdraw cash type
	if bill.TxType == TxTypeCashWithdraw {
		bill.Type = OrderTypeRecv
	}

	bill.Money, err = strconv.ParseFloat(array[5][2:], 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", array[5], err)
	}
	bill.Method = array[6]
	bill.Status = array[7]
	bill.OrderID = array[8]
	bill.MechantOrderID = array[9]
	note := array[10]

	// deal with the commission
	if strings.Contains(note, "服务费") {
		commissionStr := commissionRegex.FindString(note)
		bill.Commission, err = strconv.ParseFloat(commissionStr, 64)
		if err != nil {
			return fmt.Errorf("parse commission %s error: %v", commissionStr, err)
		}
		// update money of this transaction here (exclude the commission)
		bill.Money = bill.Money - bill.Commission
	}

	w.Orders = append(w.Orders, bill)
	return nil
}
