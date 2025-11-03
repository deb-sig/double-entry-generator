package bocom_debit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

// translateLine parses a single CSV row into an Order.
func (b *BocomDebit) translateLine(row []string) error {
	for idx, col := range row {
		row[idx] = strings.TrimSpace(col)
	}

	// abstract field is optional
	if len(row) < 10 {
		return fmt.Errorf("row length is less than expected(10)")
	}

	var (
		order Order
		err   error
	)

	order.SerialNum = strings.TrimPrefix(row[0], "\ufeff")
	order.TransDate = row[1]
	order.TransTime = row[2]

	order.TradingType = row[3]
	order.DcFlg = row[4]

	amountStr := strings.ReplaceAll(row[5], ",", "")
	if amountStr != "" {
		order.TransAmt, err = strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("parse amount %s error: %v", row[5], err)
		}
	}

	balanceStr := strings.ReplaceAll(row[6], ",", "")
	if balanceStr != "" {
		order.Balance, err = strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return fmt.Errorf("parse balance %s error: %v", row[6], err)
		}
	}

	order.PaymentReceiptAccount = row[7]
	order.PaymentReceiptAccountName = row[8]
	order.TradingPlace = row[9]
	if len(row) > 10 {
		order.Abstract = row[10]
	}

	orderType := determineOrderType(order.DcFlg)
	payTime, err := parsePayTime(order)
	if err != nil {
		return err
	}

	b.updateStatistics(order, orderType, payTime)
	b.Orders = append(b.Orders, order)
	return nil
}

func determineOrderType(dcFlg string) OrderType {
	dcFlg = strings.TrimSpace(dcFlg)
	switch {
	case strings.Contains(dcFlg, "贷"):
		return OrderTypeRecv
	case strings.Contains(dcFlg, "借"):
		return OrderTypeSend
	default:
		return OrderTypeUnknown
	}
}

func parsePayTime(order Order) (time.Time, error) {
	payTimeStr := strings.TrimSpace(order.TransDate + " " + order.TransTime)
	if payTimeStr == "" {
		return time.Time{}, nil
	}
	payTime, err := time.Parse(timeLayout, payTimeStr+" +0800 CST")
	if err != nil {
		return time.Time{}, fmt.Errorf("parse pay time %s error: %v", payTimeStr, err)
	}
	return payTime, nil
}

func buildPeer(name, account string) string {
	name = strings.TrimSpace(name)
	account = strings.TrimSpace(account)
	switch {
	case name == "":
		return account
	case account == "":
		return name
	default:
		return strings.TrimSpace(name + " " + account)
	}
}

func buildItem(tradingPlace, abstract string) string {
	tradingPlace = strings.TrimSpace(tradingPlace)
	abstract = strings.TrimSpace(abstract)
	if abstract == "" {
		return tradingPlace
	}
	if tradingPlace == "" {
		return abstract
	}
	return strings.TrimSpace(tradingPlace + " " + abstract)
}

// updateStatistics updates statistics with the parsed order.
func (b *BocomDebit) updateStatistics(order Order, orderType OrderType, payTime time.Time) {
	b.Statistics.ParsedItems++

	if orderType == OrderTypeRecv {
		b.Statistics.TotalInRecords++
		b.Statistics.TotalInMoney += order.TransAmt
	} else if orderType == OrderTypeSend {
		b.Statistics.TotalOutRecords++
		b.Statistics.TotalOutMoney += order.TransAmt
	}

	if payTime.IsZero() {
		return
	}

	if b.Statistics.Start.IsZero() || payTime.Before(b.Statistics.Start) {
		b.Statistics.Start = payTime
	}
	if b.Statistics.End.IsZero() || payTime.After(b.Statistics.End) {
		b.Statistics.End = payTime
	}
}
