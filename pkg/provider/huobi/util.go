package huobi

func getOrderType(s string) OrderType {
	switch s {
	case string(OrderTypeCoin):
		return OrderTypeCoin
	default:
		return OrderTypeUnknown
	}
}

func getTxType(s string) TxType {
	switch s {
	case string(TxTypeBuy):
		return TxTypeBuy
	case string(TxTypeSell):
		return TxTypeSell
	default:
		return TxTypeNil
	}
}
