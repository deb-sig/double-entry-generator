package td

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translate csv file to []Order.
func (td *Td) translateToOrders(columns []string) error {
	for idx, record := range columns {
		record = strings.Trim(record, " ")
		record = strings.Trim(record, "\t")
		columns[idx] = record
	}

	var bill Order
	var err error
	bill.PayTime, err = time.Parse(LocalTimeFmt, columns[0]+" -0700")
	if err != nil {
		return fmt.Errorf("parse Pay time %s error: %v", columns[0], err)
	}

	bill.TransactionDescription = columns[1]
	var withdrawal = columns[2]
	var deposit = columns[3]
	bill.Type = getOrderType(withdrawal, deposit)
	var money string
	if len(withdrawal) > 0 {
		money = withdrawal
	} else {
		money = deposit
	}
	bill.Money, err = strconv.ParseFloat(money, 64)
	if err != nil {
		return fmt.Errorf("parse money %s error: %v", columns[4], err)
	}

	bill.Balance, err = strconv.ParseFloat(columns[4], 64)
	if err != nil {
		return fmt.Errorf("parse balance %s error: %v", columns[4], err)
	}

	td.Orders = append(td.Orders, bill)
	return nil
}
