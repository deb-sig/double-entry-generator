package bocomcredit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var beijingLocation = time.FixedZone("CST", 8*3600)

// parseDate trims whitespace and parses a trade or record date (for example,
// "2025-01-10") in the statement-provided layout within the Beijing time
// zone.
func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date value")
	}
	return time.ParseInLocation(dateLayout, value, beijingLocation)
}

// splitCurrencyAmount splits a string like "CNY 12.34" or "CNY12.34" into its
// currency and absolute numeric amount components, e.g. ("CNY", 12.34).
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
	} else {
		// Handle compact forms like "CNY13.50" where the currency and amount are
		// concatenated without whitespace.
		idx := strings.IndexFunc(currency, func(r rune) bool { return !unicode.IsLetter(r) })
		if idx <= 0 || idx >= len(currency) {
			return "", 0, fmt.Errorf("invalid amount field: %s", value)
		}
		amountField = currency[idx:]
		currency = currency[:idx]
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

// splitDescription extracts the leading transaction type token and the
// remaining description from a statement description field, e.g. parsing
// "消费 京东商城平台商户" into ("消费", "京东商城平台商户").
func splitDescription(description string) (string, string) {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return "", ""
	}
	parts := strings.SplitN(desc, " ", 2)
	prefix := strings.TrimSpace(parts[0])
	if len(parts) == 1 {
		return prefix, ""
	}
	return prefix, strings.TrimSpace(parts[1])
}

func inferOrderType(typeOriginal string) (OrderType, error) {
	switch strings.TrimSpace(typeOriginal) {
	case "退货", "信用卡还款", "红包还款", "刷卡金返还":
		return OrderTypeRecv, nil
	case "消费", "刷卡金扣回":
		return OrderTypeSend, nil
	default:
		return OrderTypeUnknown, fmt.Errorf("unsupported transaction type: %s", typeOriginal)
	}
}
