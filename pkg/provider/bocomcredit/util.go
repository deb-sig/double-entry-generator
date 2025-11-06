package bocomcredit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var beijingLocation = time.FixedZone("CST", 8*3600)

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date value")
	}
	return time.ParseInLocation(dateLayout, value, beijingLocation)
}

func splitCurrencyAmount(value string) (string, float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0, fmt.Errorf("empty amount value")
	}
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", 0, fmt.Errorf("invalid amount field: %s", value)
	}
	currency := fields[0]
	amountField := ""
	if len(fields) > 1 {
		amountField = fields[1]
	}
	if amountField == "" {
		return currency, 0, fmt.Errorf("invalid amount field: %s", value)
	}
	amountField = strings.ReplaceAll(amountField, ",", "")
	amount, err := strconv.ParseFloat(amountField, 64)
	if err != nil {
		return currency, 0, fmt.Errorf("parse amount error: %w", err)
	}
	if amount < 0 {
		amount = -amount
	}
	return currency, amount, nil
}

func inferOrderType(description string) OrderType {
	desc := strings.TrimSpace(description)
	switch {
	case strings.HasPrefix(desc, "退货"):
		return OrderTypeRecv
	case strings.HasPrefix(desc, "信用卡还款"):
		return OrderTypeRecv
	case strings.HasPrefix(desc, "消费"):
		return OrderTypeSend
	default:
		return OrderTypeUnknown
	}
}
