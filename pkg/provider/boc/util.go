package boc

import "strings"

// 仅有信用卡区分了收支
// 储蓄卡直接按照金额正负判断

func getOrderTypeByTransactionAmount(amount string) OrderType {
	if strings.HasPrefix(amount, "-") {
		return TypeSend
	} else {
		return TypeRecv
	}
}
