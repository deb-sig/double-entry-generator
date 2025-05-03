package citic

import "strings"

func getOrderTypeByTransactionAmount(amount string) OrderType {
	if strings.HasPrefix(amount, "-") {
		return OrderTypeRecv
	} else {
		return OrderTypeSend
	}
}
