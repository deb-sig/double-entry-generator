package bmo

import "strings"

func getOrderType(transactionType string) OrderType {
	if transactionType == string(TransactionTypeDebit) {
		return OrderTypeSend
	} else if transactionType == string(TransactionTypeCredit) {
		return OrderTypeRecv
	} else {
		return OrderTypeUnknown
	}
}

func getOrderTypeByTransactionAmount(amount string) OrderType {
	if strings.HasPrefix(amount, "-") {
		return OrderTypeRecv
	} else {
		return OrderTypeSend
	}
}
