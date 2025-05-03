package huobi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (h *Huobi) translateToOrders(arr []string) error {
	// trim strings
	for idx, a := range arr {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		arr[idx] = a
	}
	var bill Order
	var err error
	bill.PayTime, err = time.Parse(LocalTimeFmt, arr[0]+" +0800") // UTC+8 by default
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", arr[0], err)
	}

	bill.TxType = getTxType(arr[1])
	if bill.TxType == TxTypeUnknown {
		return fmt.Errorf("Failed to get the order type %s: %v", arr[1], err)
	}
	bill.TxTypeOriginal = arr[1]

	bill.Item = arr[2]
	units := strings.Split(arr[2], "/")
	if len(units) != 2 {
		return fmt.Errorf("Failed to get the base & target units from %s", arr[2])
	}
	bill.BaseUnit = units[1]
	bill.TargetUnit = units[0]

	bill.Type = getOrderType(arr[3])
	if bill.Type == TypeNil {
		return fmt.Errorf("Failed to get the tx type: %s: %v", arr[3], err)
	}
	bill.Price, err = strconv.ParseFloat(arr[4], 64)
	if err != nil {
		return fmt.Errorf("parse price %s error: %v", arr[4], err)
	}
	bill.Amount, err = strconv.ParseFloat(arr[5], 64)
	if err != nil {
		return fmt.Errorf("parse amount %s error: %v", arr[5], err)
	}
	bill.Money, err = strconv.ParseFloat(arr[6], 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", arr[6], err)
	}

	co, err := regexp.Compile(`([.\d]*)(\w+)`)
	if err != nil {
		return fmt.Errorf("Failed to compile the regex")
	}
	co_res := co.FindStringSubmatch(arr[7])
	bill.Commission, err = strconv.ParseFloat(co_res[1], 64)
	if err != nil {
		return fmt.Errorf("parse commission %s error: %v", co_res[1], err)
	}
	bill.CommissionUnit = co_res[2]

	h.Orders = append(h.Orders, bill)
	return nil
}
