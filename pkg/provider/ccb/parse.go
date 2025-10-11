package ccb

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// translateToOrders translates csv/xlsx line to []Order.
func (ccb *CCB) translateToOrders(array []string) error {
	for idx, a := range array {
		a = strings.TrimSpace(a)
		array[idx] = a
	}

	if len(array) < 5 { // A valid row must have at least date, date, expense, income, currency
		log.Printf("ignore the invalid line, too short: %+v\n", array)
		return nil
	}

	var bill Order
	var err error

	// 1. Determine offset for the optional time column
	isTimePresent := false
	if len(array) > 2 {
		_, timeParseErr := time.Parse("15:04:05", array[2])
		if timeParseErr == nil {
			isTimePresent = true
		}
	}
	timeOffset := 0
	if !isTimePresent {
		timeOffset = -1
	}

	// 2. Parse fixed-position fields
	bill.RecordDate = array[0]
	bill.PayTime, err = time.Parse(localTimeFmt, array[1]+" +0800 CST")
	if err != nil {
		return fmt.Errorf("parse trade date %s error: %v", array[1], err)
	}
	if isTimePresent {
		bill.TradeTime = array[2]
	}

	// Expense and Income are mandatory for a valid transaction row.
	expenseIndex := 3 + timeOffset
	incomeIndex := 4 + timeOffset
	if incomeIndex >= len(array) {
		log.Printf("ignore invalid line, no expense/income fields: %+v\n", array)
		return nil
	}
	// Parse Expense
	expenseStr := strings.TrimSpace(strings.ReplaceAll(array[expenseIndex], ",", ""))
	if expenseStr != "" {
		bill.Expense, err = strconv.ParseFloat(expenseStr, 64)
		if err != nil {
			return fmt.Errorf("parse expense %s error: %v", array[expenseIndex], err)
		}
	} else {
		bill.Expense = 0
	}
	// Parse Income
	incomeStr := strings.TrimSpace(strings.ReplaceAll(array[incomeIndex], ",", ""))
	if incomeStr != "" {
		bill.Income, err = strconv.ParseFloat(incomeStr, 64)
		if err != nil {
			return fmt.Errorf("parse income %s error: %v", array[incomeIndex], err)
		}
	} else {
		bill.Income = 0
	}

	// 3. Determine offset for the optional balance column
	balanceMissingOffset := 0
	balanceIndex := 5 + timeOffset
	if balanceIndex >= len(array) {
		balanceMissingOffset = -1
		bill.Balances = 0
	} else {
		balanceStr := strings.TrimSpace(strings.ReplaceAll(array[balanceIndex], ",", ""))
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			// Can't parse, so balance column is missing
			balanceMissingOffset = -1
			bill.Balances = 0
		} else {
			bill.Balances = balance
		}
	}

	// 4. Parse all subsequent fields using combined offsets, with length checks
	finalOffset := timeOffset + balanceMissingOffset

	baseIndex := 5
	if len(array) > baseIndex+1+finalOffset {
		bill.Currency = array[baseIndex+1+finalOffset]
	}
	if len(array) > baseIndex+2+finalOffset {
		bill.TxTypeOriginal = array[baseIndex+2+finalOffset]
	}
	if len(array) > baseIndex+3+finalOffset {
		bill.PeerAccountNum = array[baseIndex+3+finalOffset]
	}
	if len(array) > baseIndex+4+finalOffset {
		bill.PeerAccountName = array[baseIndex+4+finalOffset]
	}
	if len(array) > baseIndex+5+finalOffset {
		bill.Region = array[baseIndex+5+finalOffset]
	}

	// 计算交易金额和类型
	if bill.Expense > 0 && bill.Income == 0 {
		bill.Money = bill.Expense // 支出
		bill.Type = OrderTypeSend
	} else if bill.Income > 0 && bill.Expense == 0 {
		bill.Money = bill.Income // 收入
		bill.Type = OrderTypeRecv
	} else if bill.Income > 0 && bill.Expense > 0 {
		// 如果收入和支出都有，可能是转账或者有手续费的交易
		// 根据净额判断类型
		bill.Money = bill.Income - bill.Expense
		if bill.Money > 0 {
			bill.Type = OrderTypeRecv
		} else if bill.Money < 0 {
			bill.Type = OrderTypeSend
			bill.Money = -bill.Money
		} else {
			// 收入=支出，可能是转账，根据交易类型判断
			if strings.Contains(bill.TxTypeOriginal, "转账") || strings.Contains(bill.TxTypeOriginal, "转出") {
				bill.Type = OrderTypeSend
				bill.Money = bill.Expense
			} else if strings.Contains(bill.TxTypeOriginal, "转入") {
				bill.Type = OrderTypeRecv
				bill.Money = bill.Income
			} else {
				// 默认作为支出处理
				bill.Type = OrderTypeSend
				bill.Money = bill.Expense
			}
		}
	} else {
		// 收入和支出都为0，根据交易类型和余额变化判断
		// 检查是否是余额查询或者无效交易
		if strings.Contains(bill.TxTypeOriginal, "查询") || strings.Contains(bill.TxTypeOriginal, "余额") {
			bill.Type = OrderTypeUnknown
		} else {
			// 其他情况，根据交易类型判断
			if strings.Contains(bill.TxTypeOriginal, "消费") || strings.Contains(bill.TxTypeOriginal, "支出") {
				bill.Type = OrderTypeSend
				bill.Money = 0 // 金额为0的消费
			} else if strings.Contains(bill.TxTypeOriginal, "收入") || strings.Contains(bill.TxTypeOriginal, "转入") {
				bill.Type = OrderTypeRecv
				bill.Money = 0 // 金额为0的收入
			} else {
				// 默认作为支出处理
				bill.Type = OrderTypeSend
				bill.Money = 0
			}
		}
	}

	// 设置对方信息
	if bill.PeerAccountName != "" {
		bill.Peer = bill.PeerAccountName
	} else {
		bill.Peer = bill.PeerAccountNum
	}

	// 设置交易详情
	bill.Item = bill.Region

	// 更新统计信息
	ccb.updateStatistics(bill)

	ccb.Orders = append(ccb.Orders, bill)
	return nil
}

// updateStatistics updates the statistics of the bill.
func (ccb *CCB) updateStatistics(bill Order) {
	ccb.Statistics.ParsedItems++

	if bill.Type == OrderTypeRecv {
		ccb.Statistics.TotalInRecords++
		ccb.Statistics.TotalInMoney += bill.Money
	} else if bill.Type == OrderTypeSend {
		ccb.Statistics.TotalOutRecords++
		ccb.Statistics.TotalOutMoney += -bill.Money // 支出金额为正数
	}

	// 更新开始和结束时间
	if ccb.Statistics.Start.IsZero() || bill.PayTime.Before(ccb.Statistics.Start) {
		ccb.Statistics.Start = bill.PayTime
	}
	if ccb.Statistics.End.IsZero() || bill.PayTime.After(ccb.Statistics.End) {
		ccb.Statistics.End = bill.PayTime
	}
} 