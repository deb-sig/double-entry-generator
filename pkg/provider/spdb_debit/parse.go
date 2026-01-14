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
		// Skip rows that don't have enough columns (at least transaction number, date-time, summary, deposit, withdraw, balance)
		return nil
	}

	for i := range row {
		row[i] = strings.TrimSpace(row[i])
	}

	// Skip empty rows (check the first column which should be the transaction number)
	if row[0] == "" {
		return nil
	}

	// 解析交易时间，格式为 "2025/12/03  14:06:51"
	// 日期和时间中间用空格分隔
	dateTime := row[1]
	// 使用Fields函数处理多个空格，返回非空字段切片
	parts := strings.Fields(dateTime)
	
	dateStr := ""
	timeStr := ""
	
	if len(parts) >= 2 {
		dateStr = parts[0]
		timeStr = parts[1]
	} else if len(parts) == 1 {
		// 只有日期部分
		dateStr = parts[0]
		timeStr = ""
	} else {
		// 空字符串，跳过
		return nil
	}
	
	// 确保日期不为空
	if dateStr == "" {
		return nil
	}
	
	// 解析交易金额
	// 根据存入金额和取出金额确定交易类型和金额
	var amountStr string
	deposit := row[3]
	withdraw := row[4]
	
	// 支出为正数，收入为负数
	if deposit != "" && deposit != "0" {
		// 存入金额不为空，是收入
		amountStr = "-" + deposit
	} else if withdraw != "" && withdraw != "0" {
		// 取出金额不为空，是支出
		amountStr = withdraw
	} else {
		// 金额为0，跳过
		return nil
	}
	
	// 解析其他字段
	summary := row[2]
	balance := row[5]
	peer := ""
	account := ""
	
	if len(row) > 6 {
		peer = row[6] // 第7列是对方户名
	}
	if len(row) > 7 {
		account = row[7] // 第8列是对方账号
	}
	
	order := Order{
		TradeDate:     dateStr,   // 日期
		TradeTime:     timeStr,   // 时间
		Summary:       summary,   // 交易摘要
		Amount:        amountStr, // 交易金额（支出为正，收入为负）
		Balance:       balance,   // 借记卡余额
		Peer:          peer,      // 对方户名
		Account:       account,    // 对方账号
		Channel:       "",        // 交易渠道
		Postscript:    "",        // 附言
		TransactionID: row[0],    // 交易序号
	}

	sd.Orders = append(sd.Orders, order)

	return nil
}



// parseTradeTime parses the trade time.
func parseTradeTime(dateStr, timeStr string) (time.Time, error) {
	// 处理两种情况：
	// 1. dateStr 已经包含完整的日期和时间（如："2025/12/03  14:06:51"）
	// 2. dateStr 只有日期，timeStr 有时间
	
	// 检查 dateStr 是否包含时间（包含空格）
	if strings.Contains(dateStr, " ") {
		// 处理日期和时间在同一列的情况
		// 支持多种空格分隔（1个或多个空格）
		parts := strings.Fields(dateStr)
		if len(parts) >= 2 {
			parseStr := parts[0] + " " + parts[1]
			layout := dateLayout + " " + timeLayout
			return time.ParseInLocation(layout, parseStr, time.FixedZone("Asia/Shanghai", beijingOffset))
		}
	}
	
	// 常规情况：dateStr 只有日期
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
	// 去除金额字符串中的空格和逗号
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)

	if amountStr == "" {
		return 0, OrderTypeUnknown, fmt.Errorf("empty amount string")
	}

	// 浦发银行借记卡交易明细中，支出为正数，收入为负数
	// 这里需要根据金额的正负来判断交易类型
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, OrderTypeUnknown, fmt.Errorf("failed to parse amount '%s': %v", amountStr, err)
	}

	var txType OrderType
	if amount > 0 {
		txType = OrderTypeSend
		// 支出金额保持正数
	} else if amount < 0 {
		txType = OrderTypeRecv
		// 收入金额转换为正数
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
