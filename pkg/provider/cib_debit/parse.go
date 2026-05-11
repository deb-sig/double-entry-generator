package cib_debit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseCurrency(accountLine string) string {
	switch {
	case strings.Contains(accountLine, "美元"):
		return "USD"
	case strings.Contains(accountLine, "港币"):
		return "HKD"
	case strings.Contains(accountLine, "人民币"):
		return "CNY"
	default:
		return defaultCurrency
	}
}

func (c *CibDebit) translateRow(row []string) error {
	for idx := range row {
		row[idx] = strings.TrimSpace(row[idx])
	}
	if len(row) < 6 || row[0] == "" {
		return nil
	}
	if strings.Contains(row[0], "说明") || strings.Contains(row[0], "交易时间") {
		return nil
	}
	if len(row) < 12 {
		padded := make([]string, 12)
		copy(padded, row)
		row = padded
	}

	order := Order{
		TradeTime:     row[0],
		AccountingDay: row[1],
		Expense:       row[2],
		Income:        row[3],
		Balance:       row[4],
		Summary:       row[5],
		Peer:          row[6],
		PeerBank:      row[7],
		PeerAccount:   row[8],
		Purpose:       row[9],
		Channel:       row[10],
		Remark:        row[11],
		AccountName:   c.AccountName,
		AccountNum:    c.AccountNum,
		SubAccount:    c.SubAccount,
		Currency:      c.Currency,
	}

	money, orderType, err := parseMoneyAndType(order.Expense, order.Income)
	if err != nil {
		return nil
	}
	if orderType == OrderTypeUnknown || money == 0 {
		return nil
	}

	payTime, err := parseTradeTime(order.TradeTime, order.AccountingDay)
	if err != nil {
		return err
	}
	c.updateStatistics(money, orderType, payTime)
	c.Orders = append(c.Orders, order)
	return nil
}

func parseTradeTime(tradeTime, accountingDay string) (time.Time, error) {
	loc := time.FixedZone("Asia/Shanghai", beijingOffset)
	tradeTime = strings.TrimSpace(tradeTime)
	if tradeTime != "" {
		if t, err := time.ParseInLocation(dateTimeLayout, tradeTime, loc); err == nil {
			return t, nil
		}
		if t, err := time.ParseInLocation(dateLayout, tradeTime, loc); err == nil {
			return t, nil
		}
	}
	accountingDay = strings.TrimSpace(accountingDay)
	if accountingDay == "" {
		return time.Time{}, fmt.Errorf("empty trade time")
	}
	return time.ParseInLocation(dateLayout, accountingDay, loc)
}

func parseMoneyAndType(expense, income string) (float64, OrderType, error) {
	expense = normalizeMoney(expense)
	income = normalizeMoney(income)

	if expense != "" {
		amount, err := strconv.ParseFloat(expense, 64)
		if err != nil {
			return 0, OrderTypeUnknown, fmt.Errorf("parse expense %q error: %w", expense, err)
		}
		if amount > 0 {
			return amount, OrderTypeSend, nil
		}
		if amount < 0 {
			return -amount, OrderTypeRecv, nil
		}
	}

	if income != "" {
		amount, err := strconv.ParseFloat(income, 64)
		if err != nil {
			return 0, OrderTypeUnknown, fmt.Errorf("parse income %q error: %w", income, err)
		}
		if amount > 0 {
			return amount, OrderTypeRecv, nil
		}
		if amount < 0 {
			return -amount, OrderTypeSend, nil
		}
	}

	return 0, OrderTypeUnknown, nil
}

func normalizeMoney(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "")
	return value
}

func normalizePeer(peer, peerBank, peerAccount string) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{peer, peerBank, peerAccount} {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return providerPeer
	}
	return strings.Join(parts, " ")
}

func normalizeItem(summary, purpose, remark string) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{summary, purpose, remark} {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " ")
}

func (c *CibDebit) updateStatistics(money float64, orderType OrderType, payTime time.Time) {
	c.Statistics.ParsedItems++
	if orderType == OrderTypeRecv {
		c.Statistics.TotalInRecords++
		c.Statistics.TotalInMoney += money
	} else if orderType == OrderTypeSend {
		c.Statistics.TotalOutRecords++
		c.Statistics.TotalOutMoney += money
	}
	if payTime.IsZero() {
		return
	}
	if c.Statistics.Start.IsZero() || payTime.Before(c.Statistics.Start) {
		c.Statistics.Start = payTime
	}
	if c.Statistics.End.IsZero() || payTime.After(c.Statistics.End) {
		c.Statistics.End = payTime
	}
}
