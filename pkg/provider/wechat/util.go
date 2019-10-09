package wechat

import "strings"

func getOrderType(ot string) OrderType {
	switch ot {
	case string(OrderTypeRecv):
		return OrderTypeRecv
	case string(OrderTypeSend):
		return OrderTypeSend
	default:
		return OrderTypeUnknown
	}
}

func getTxType(tt string) TxType {
	if strings.Contains(tt, string(TxTypeLucky)) {
		return TxTypeLucky
	} else if strings.Contains(tt, string(TxTypeConsume)) {
		return TxTypeConsume
	} else if strings.Contains(tt, string(TxTypeTransfer)) {
		return TxTypeTransfer
	} else {
		return TxTypeUnknown
	}
}
