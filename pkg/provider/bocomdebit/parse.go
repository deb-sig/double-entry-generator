package bocomdebit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

// translateLine parses a single CSV row into an Order.
func (b *Bocom) translateLine(row []string) error {
	for idx, col := range row {
		row[idx] = strings.TrimSpace(col)
	}

	if len(row) < 11 {
		return fmt.Errorf("row length is less than expected(11)")
	}

	var (
		order Order
		err   error
	)

	order.Sequence = strings.TrimPrefix(row[0], "\ufeff")

	payTimeStr := strings.TrimSpace(row[1] + " " + row[2])
	if payTimeStr != "" {
		order.PayTime, err = time.Parse(timeLayout, payTimeStr+" +0800 CST")
		if err != nil {
			return fmt.Errorf("parse pay time %s error: %v", payTimeStr, err)
		}
	}

	order.TxTypeOriginal = row[3]
	order.TypeOriginal = row[4]

	amountStr := strings.ReplaceAll(row[5], ",", "")
	if amountStr != "" {
		order.Money, err = strconv.ParseFloat(amountStr, 64)
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

	order.PeerAccount = row[7]
	order.PeerName = row[8]
	order.Location = row[9]
	if len(row) > 10 {
		order.Summary = row[10]
	}

	order.Peer = buildPeer(order.PeerName, order.PeerAccount)
	order.Item = buildItem(order.Location, order.Summary)

	switch {
	case strings.Contains(order.TypeOriginal, "贷"):
		order.Type = OrderTypeRecv
	case strings.Contains(order.TypeOriginal, "借"):
		order.Type = OrderTypeSend
	default:
		order.Type = OrderTypeUnknown
	}

	b.updateStatistics(order)
	b.Orders = append(b.Orders, order)
	return nil
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

func buildItem(location, summary string) string {
	location = strings.TrimSpace(location)
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return location
	}
	if location == "" {
		return summary
	}
	return strings.TrimSpace(location + " " + summary)
}

// updateStatistics updates statistics with the parsed order.
func (b *Bocom) updateStatistics(order Order) {
	b.Statistics.ParsedItems++

	if order.Type == OrderTypeRecv {
		b.Statistics.TotalInRecords++
		b.Statistics.TotalInMoney += order.Money
	} else if order.Type == OrderTypeSend {
		b.Statistics.TotalOutRecords++
		b.Statistics.TotalOutMoney += order.Money
	}

	if order.PayTime.IsZero() {
		return
	}

	if b.Statistics.Start.IsZero() || order.PayTime.Before(b.Statistics.Start) {
		b.Statistics.Start = order.PayTime
	}
	if b.Statistics.End.IsZero() || order.PayTime.After(b.Statistics.End) {
		b.Statistics.End = order.PayTime
	}
}
