package htsec

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (h *Htsec) translateToOrders(arr []string) error {
	// trim strings
	for idx, a := range arr {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		arr[idx] = a
	}
	var bill Order
	var err error

	var code string

	code = fmt.Sprintf("%06s", arr[0])

	if arr[1] == "新增证券" {
		return nil
	}

	if strings.HasPrefix(arr[17], "A") {
		bill.TypeOriginal = "SH" + code
	} else {
		bill.TypeOriginal = "SZ" + code
	}
	bill.SecuritiesName = code + "-" + arr[1]
	if len(arr[3]) == 0 {
		arr[3] = "00:00:00"
	}
	bill.TransactionTime, err = time.Parse(LocalTimeFmt, arr[2][0:4]+"-"+arr[2][4:6]+"-"+arr[2][6:8]+" "+arr[3]+" +0800") // UTC+8 by default
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", arr[0], err)
	}

	bill.Volume, err = strconv.ParseInt(arr[4], 10, 64)
	if err != nil {
		return fmt.Errorf("parse Volume %s error: %v", arr[4], err)
	}

	bill.Price, err = strconv.ParseFloat(arr[5], 64)
	if err != nil {
		return fmt.Errorf("parse Price %s error: %v", arr[5], err)
	}

	bill.TransactionAmount, err = strconv.ParseFloat(arr[6], 64)
	if err != nil {
		return fmt.Errorf("parse TransactionAmount %s error: %v", arr[6], err)
	}

	bill.OccurAmount, err = strconv.ParseFloat(arr[7], 64)
	if err != nil {
		return fmt.Errorf("parse OccurAmount %s error: %v", arr[7], err)
	}

	bill.TxType = getTxType(arr[8])
	if bill.TxType == TxTypeNil {
		return fmt.Errorf("Failed to get the tx type: %s: %v", arr[8], err)
	}

	bill.OrderID = arr[9]
	bill.TransactionID = arr[10]

	bill.Commission, err = strconv.ParseFloat(arr[11], 64)
	if err != nil {
		return fmt.Errorf("parse commission %s error: %v", arr[11], err)
	}

	bill.StampDuty, err = strconv.ParseFloat(arr[12], 64)
	if err != nil {
		return fmt.Errorf("parse stamp duty %s error: %v", arr[12], err)
	}

	bill.TransferFee, err = strconv.ParseFloat(arr[13], 64)
	if err != nil {
		return fmt.Errorf("parse transfer fee %s error: %v", arr[13], err)
	}

	bill.OtherFee, err = strconv.ParseFloat(arr[14], 64)
	if err != nil {
		return fmt.Errorf("parse other fee %s error: %v", arr[14], err)
	}

	// put all transaction fees together as commission
	bill.Commission = bill.Commission + bill.StampDuty + bill.TransferFee + bill.OtherFee

	bill.RemainAmount, err = strconv.ParseFloat(arr[15], 64)
	if err != nil {
		return fmt.Errorf("parse remain amount %s error: %v", arr[15], err)
	}

	bill.RemainShare, err = strconv.ParseInt(arr[16], 10, 64)
	if err != nil {
		return fmt.Errorf("parse remain share %s error: %v", arr[16], err)
	}

	h.Orders = append(h.Orders, bill)
	return nil
}
