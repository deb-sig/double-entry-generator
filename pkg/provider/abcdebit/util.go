package abcdebit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var beijingLocation = time.FixedZone("CST", beijingOffset)

// Examples:
//
//	parseTradeTime("2024-01-02", "") -> 2024-01-02 00:00:00 +0800 CST
//	parseTradeTime("2024-01-02", "083015") -> 2024-01-02 08:30:15 +0800 CST
func parseTradeTime(dateStr, timeStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)

	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty trade date")
	}

	date, err := time.ParseInLocation(dateLayout, dateStr, beijingLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date error: %w", err)
	}

	if timeStr == "" {
		return date, nil
	}

	if len(timeStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}

	hour, err := strconv.Atoi(timeStr[:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}
	minute, err := strconv.Atoi(timeStr[2:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}
	second, err := strconv.Atoi(timeStr[4:6])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}

	if hour >= 24 || minute >= 60 || second >= 60 {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}

	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, second, 0, beijingLocation), nil
}

// Examples:
//
//	parseAmountAndType("+100.00") -> 100.00, OrderTypeRecv, nil
//	parseAmountAndType("-200.00") -> 200.00, OrderTypeSend, nil
//	parseAmountAndType("+1200.00") -> 1200.00, OrderTypeRecv, nil
func parseMoneyAndType(raw string) (float64, OrderType, error) {
	value := strings.TrimSpace(raw)
	if len(value) < 2 {
		return 0, OrderTypeUnknown, fmt.Errorf("invalid amount value: %s", value)
	}

	sign := value[0]
	if sign != '+' && sign != '-' {
		return 0, OrderTypeUnknown, fmt.Errorf("invalid amount value: %s", value)
	}

	digits := strings.ReplaceAll(value[1:], ",", "")
	if digits == "" {
		return 0, OrderTypeUnknown, fmt.Errorf("invalid amount value: %s", value)
	}

	amount, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		return 0, OrderTypeUnknown, fmt.Errorf("parse amount error: %w", err)
	}

	txType := OrderTypeRecv
	if sign == '-' {
		txType = OrderTypeSend
	}

	return amount, txType, nil
}

// Examples:
//
//	normalizePeer(" Alice ") -> "Alice"
//	normalizePeer("") -> "ABC Debit"
//	normalizePeer("--") -> "ABC Debit"
func normalizePeer(peer string) string {
	p := strings.TrimSpace(peer)
	if p == "" || p == "--" {
		return providerPeer
	}
	return p
}

// Examples:
//
//	normalizeItem(" summary ", " postscript ") -> "summary - postscript"
//	normalizeItem(" summary ", "") -> "summary"
//	normalizeItem("", " postscript ") -> "postscript"
func normalizeItem(summary, postscript string) string {
	s := strings.TrimSpace(summary)
	p := strings.TrimSpace(postscript)

	if s != "" && p != "" {
		return s + " - " + p
	}
	if s != "" {
		return s
	}
	return p
}
