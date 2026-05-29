package cgb_credit

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	beijingLocation = time.FixedZone("CST", 8*3600)
	typePattern     = regexp.MustCompile(`^[（(]([^）)]+)[）)]`)
)

// parseDate 将广发账单中的 YYYY/MM/DD 日期转换为北京时间。
func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date value")
	}
	return time.ParseInLocation(inputLayout, value, beijingLocation)
}

// parseSignedAmount 解析账单金额，保留正负号用于判断交易方向。
func parseSignedAmount(value string) (float64, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, ",", ""))
	if value == "" {
		return 0, fmt.Errorf("empty amount value")
	}
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse amount error: %w", err)
	}
	return amount, nil
}

func normalizeCurrency(value string) string {
	switch strings.TrimSpace(value) {
	case "人民币", "CNY", "RMB":
		return "CNY"
	case "美元", "USD":
		return "USD"
	case "日元", "JPY":
		return "JPY"
	case "欧元", "EUR":
		return "EUR"
	default:
		return strings.TrimSpace(value)
	}
}

func extractType(description string) string {
	match := typePattern.FindStringSubmatch(strings.TrimSpace(description))
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func inferOrderType(amount float64, typeOriginal string) (OrderType, error) {
	if amount < 0 {
		return OrderTypeRecv, nil
	}
	if amount > 0 {
		return OrderTypeSend, nil
	}

	switch strings.TrimSpace(typeOriginal) {
	case "退款", "还款", "入账":
		return OrderTypeRecv, nil
	case "消费", "分期", "取现":
		return OrderTypeSend, nil
	default:
		return OrderTypeUnknown, fmt.Errorf("unsupported transaction type: %s", typeOriginal)
	}
}

func absAmount(amount float64) float64 {
	return math.Abs(amount)
}
