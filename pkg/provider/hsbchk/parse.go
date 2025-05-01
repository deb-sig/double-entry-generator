package hsbchk

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translateToOrders 将一行CSV数据解析为订单
func (h *HsbcHK) translateToOrders(line []string) error {
	// 根据模式判断处理方式
	if h.Mode == DebitMode {
		return h.translateToDebitOrders(line)
	} else {
		return h.translateToCreditOrders(line)
	}
}

// translateToDebitOrders 解析借记卡数据
// 列: Date,Description,Billing amount,Billing currency,Balance,Balance currency
func (h *HsbcHK) translateToDebitOrders(line []string) error {
	if len(line) < 5 {
		return fmt.Errorf("invalid line format for debit card record, got %d fields", len(line))
	}

	// 解析日期
	payTime, err := time.Parse(TimeFormat, line[0])
	if err != nil {
		return fmt.Errorf("failed to parse date: %v", err)
	}

	// 解析交易金额
	moneyStr := strings.Replace(strings.Replace(line[2], ",", "", -1), "\"", "", -1)
	money, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse billing amount: %v", err)
	}

	// 解析余额
	balanceStr := strings.Replace(strings.Replace(line[4], ",", "", -1), "\"", "", -1)
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %v", err)
	}

	// 根据金额判断收支类型
	orderType := OrderTypeUnknown
	if money < 0 {
		orderType = OrderTypeSend
		money = -money
	} else if money > 0 {
		orderType = OrderTypeRecv
	}

	// 记录统计
	h.Statistics.ParsedItems++
	if orderType == OrderTypeSend {
		h.Statistics.TotalOutRecords++
		h.Statistics.TotalOutMoney += -money
	} else if orderType == OrderTypeRecv {
		h.Statistics.TotalInRecords++
		h.Statistics.TotalInMoney += money
	}

	if h.Statistics.Start.IsZero() || payTime.Before(h.Statistics.Start) {
		h.Statistics.Start = payTime
	}
	if payTime.After(h.Statistics.End) {
		h.Statistics.End = payTime
	}

	// 创建订单
	order := Order{
		PayTime:         payTime,
		Description:     strings.TrimSpace(line[1]),
		Money:           money,
		Currency:        strings.TrimSpace(line[3]),
		Balance:         balance,
		BalanceCurrency: strings.TrimSpace(line[5]),
		Type:            orderType,
		Country:         "", // 借记卡账单没有国家信息
	}

	h.Orders = append(h.Orders, order)
	return nil
}

// translateToCreditOrders 解析信用卡数据
// 列: Transaction date,Post date,Description,Billing amount,Billing currency,Transaction status,Merchant name,Country / region,Area / district,Credit / Debit
func (h *HsbcHK) translateToCreditOrders(line []string) error {
	if len(line) < 10 {
		return fmt.Errorf("invalid line format for credit card record, got %d fields", len(line))
	}

	// 解析交易日期
	payTime, err := time.Parse(TimeFormat, line[0])
	if err != nil {
		return fmt.Errorf("failed to parse transaction date: %v", err)
	}

	// 解析入账日期
	postDate, err := time.Parse(TimeFormat, line[1])
	if err != nil {
		return fmt.Errorf("failed to parse post date: %v", err)
	}

	// 解析交易金额
	moneyStr := strings.Replace(strings.Replace(line[3], ",", "", -1), "\"", "", -1)
	money, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse billing amount: %v", err)
	}

	// 信用卡中 CREDIT 是收入(还款等), DEBIT 是支出(消费)
	orderType := OrderTypeUnknown
	creditOrDebit := strings.TrimSpace(line[9])
	if creditOrDebit == "DEBIT" {
		orderType = OrderTypeSend
		money = -money
	} else if creditOrDebit == "CREDIT" {
		orderType = OrderTypeRecv
	}

	// 记录统计
	h.Statistics.ParsedItems++
	if orderType == OrderTypeSend {
		h.Statistics.TotalOutRecords++
		h.Statistics.TotalOutMoney += -money
	} else if orderType == OrderTypeRecv {
		h.Statistics.TotalInRecords++
		h.Statistics.TotalInMoney += money
	}

	if h.Statistics.Start.IsZero() || payTime.Before(h.Statistics.Start) {
		h.Statistics.Start = payTime
	}
	if payTime.After(h.Statistics.End) {
		h.Statistics.End = payTime
	}

	// 创建订单
	order := Order{
		PayTime:        payTime,
		PostDate:       postDate, // m
		Description:    line[2],
		Money:          money,
		Currency:       line[4],
		StatusOriginal: line[5],
		Merchant:       line[6],
		Country:        line[7], // m
		Type:           orderType,
		CreditDebit:    creditOrDebit, // m
	}

	h.Orders = append(h.Orders, order)
	return nil
}
