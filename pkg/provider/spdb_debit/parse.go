package spdb_debit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// translateToOrders parses a row into an Order struct.
func (sd *SpdbDebit) translateToOrders(row []string) error {
	if len(row) < 6 {
		return nil
	}

	for i := range row {
		row[i] = strings.TrimSpace(row[i])
	}

	if row[0] == "" {
		return nil
	}

	dateTime := row[1]
	parts := strings.Fields(dateTime)

	dateStr := ""
	timeStr := ""

	if len(parts) >= 2 {
		dateStr = parts[0]
		timeStr = parts[1]
	} else if len(parts) == 1 {
		dateStr = parts[0]
		timeStr = ""
	} else {
		return nil
	}

	if dateStr == "" {
		return nil
	}

	var amountStr string
	deposit := row[3]
	withdraw := row[4]

	if deposit != "" && deposit != "0" {
		amountStr = "-" + deposit
	} else if withdraw != "" && withdraw != "0" {
		amountStr = withdraw
	} else {
		return nil
	}

	summary := row[2]
	balance := row[5]
	peer := ""
	account := ""
	postscript := ""

	if len(row) > 6 {
		peer = row[6]
	}
	if len(row) > 7 {
		account = row[7]
	}
	// 只有附言非空时才读取（第9列）
	if len(row) > 8 && strings.TrimSpace(row[8]) != "" {
		postscript = row[8]
	}

	order := Order{
		TradeDate:     dateStr,
		TradeTime:     timeStr,
		Summary:       summary,
		Amount:        amountStr,
		Balance:       balance,
		Peer:          peer,
		Account:       account,
		Channel:       "",
		Postscript:    postscript,
		TransactionID: row[0],
	}

	sd.Orders = append(sd.Orders, order)

	return nil
}

// parseTradeTime parses the trade time.
func parseTradeTime(dateStr, timeStr string) (time.Time, error) {
	if strings.Contains(dateStr, " ") {
		parts := strings.Fields(dateStr)
		if len(parts) >= 2 {
			parseStr := parts[0] + " " + parts[1]
			layout := dateLayout + " " + timeLayout
			return time.ParseInLocation(layout, parseStr, time.FixedZone("Asia/Shanghai", beijingOffset))
		}
	}

	layout := dateLayout
	if timeStr != "" {
		layout = dateLayout + " " + timeLayout
		parseStr := dateStr + " " + timeStr
		return time.ParseInLocation(layout, parseStr, time.FixedZone("Asia/Shanghai", beijingOffset))
	}
	return time.ParseInLocation(layout, dateStr, time.FixedZone("Asia/Shanghai", beijingOffset))
}

// parseMoneyAndType parses the amount and determines the transaction type.
func parseMoneyAndType(amountStr string) (float64, OrderType, error) {
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)

	if amountStr == "" {
		return 0, OrderTypeUnknown, fmt.Errorf("empty amount string")
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, OrderTypeUnknown, fmt.Errorf("failed to parse amount '%s': %v", amountStr, err)
	}

	var txType OrderType
	if amount > 0 {
		txType = OrderTypeSend
	} else if amount < 0 {
		txType = OrderTypeRecv
		amount = -amount
	} else {
		txType = OrderTypeUnknown
	}

	return amount, txType, nil
}

// normalizePeer normalizes the peer information.
func normalizePeer(peer string) string {
	peer = strings.TrimSpace(peer)
	if peer == "" {
		return providerPeer
	}
	return peer
}

// normalizeItem normalizes the item description.
func normalizeItem(summary, postscript string) string {
	item := strings.TrimSpace(summary)
	if postscript != "" {
		item += " " + strings.TrimSpace(postscript)
	}
	return strings.TrimSpace(item)
}
