package cmb

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getDebitOrderTypeByTransactionAmount(amount string) OrderType {
	if strings.HasPrefix(amount, "-") {
		return OrderTypeSend
	} else {
		return OrderTypeRecv
	}
}

func getCreditOrderTypeByTransactionAmount(amount string) OrderType {
	if strings.HasPrefix(amount, "-") {
		return OrderTypeRecv
	} else {
		return OrderTypeSend
	}
}

func indexOfStrList(strList []string, str string) int {
	for k, v := range strList {
		if str == v {
			return k
		}
	}

	return -1
}

func safeAccessStrList(strList []string, index int) string {
	if index >= 0 && index < len(strList) {
		return strList[index]
	}

	return ""
}

// 储蓄卡可选导出列所以字段可能不全，将没有的字段填充为空字符串，便于统一处理
func fillDebitRow(realHeaderTitles []string, row []string) []string {
	filledRow := make([]string, 7)
	for i, h := range allDebitHeaders {
		originIdx := indexOfStrList(realHeaderTitles, h)
		if originIdx >= 0 && originIdx < len(row) {
			filledRow[i] = row[originIdx]
		} else {
			filledRow[i] = ""
		}
	}
	return filledRow
}

// 从`招商银行信用卡对账单（个人消费卡账户 2025年01月）`标题中提取年月
func extractYearAndMonthFromCreditTitle(s string) (year, month string) {
	re := regexp.MustCompile(`((?:19|20)\d{2})年(0[1-9]|1[0-2])月`)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		return "", ""
	}
	year = matches[1]
	month = matches[2]
	return
}

// 储蓄卡交易日期 YYYY-MM-DD
func isValidDebitDateFormat(s string) bool {
	pattern := `^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// 信用卡卡末四位数字
func isValidCreditCardNoFormat(s string) bool {
	pattern := `^\d{4}$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func updateCardMode(row []string) CardMode {
	// `记账日期` -> 储蓄卡
	if row[0] == allDebitHeaders[0] {
		return CardModeDebit
	}

	// `交易日` -> 信用卡
	if row[0] == allCreditHeaders[0] {
		return CardModeCredit
	}

	// `银行信用卡对账单（个人消费卡账户 YYYY年MM月）` -> 信用卡
	year, month := extractYearAndMonthFromCreditTitle(row[0])
	if year != "" && month != "" {
		return CardModeCredit
	}

	return CardModeUnknown
}

// 将信用卡交易记录 MM/DD 格式的日期，结合出帐年月，算出来实际的交易日期
func convertCreditBillDate(billDate, yearStr, monthStr string) (time.Time, error) {
	// 分割账单日期字符串
	parts := strings.Split(billDate, "/")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("无效的账单日期格式: %s", billDate)
	}

	// 解析账单日期中的月份和日期
	billMonthStr, billDayStr := parts[0], parts[1]
	billMonth, err := strconv.Atoi(billMonthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法解析账单日期的月份: %s", billDate)
	}
	billDay, err := strconv.Atoi(billDayStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法解析账单日期的日期: %s", billDate)
	}

	// 解析出账年月
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法解析出账年份: %s", yearStr)
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法解析出账月份: %s", monthStr)
	}

	// 处理可能的跨年情况
	if billMonth > month {
		year--
	}

	// 获取 UTC+8 时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, fmt.Errorf("无法加载时区: %v", err)
	}

	// 构造 time.Time 对象，使用 UTC+8 时区
	return time.Date(year, time.Month(billMonth), billDay, 0, 0, 0, 0, loc), nil
}

func extractCreditAmount(s string) (float64, error) {
	// 先移除金额中的逗号，以便兼容带千分位分隔符的数字（例如 "-1,234.56"）
	s = strings.ReplaceAll(s, ",", "")

	// 定义正则表达式，用于匹配可选负号和小数部分的金额（例如 "-1234.56" 或 "123"）
	re := regexp.MustCompile(`^-?\d+(?:\.\d+)?`)
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("未找到有效的金额数字")
	}

	// 将匹配到的字符串转换为 float64 类型
	amount, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, fmt.Errorf("将金额字符串转换为浮点数时出错: %v", err)
	}

	return math.Abs(amount), nil
}
