package huobi

func getTxType(s string) TxType {
	switch s {
	case string(TxTypeCoin):
		return TxTypeCoin
	default:
		return TxTypeUnknown
	}
}

func getOrderType(s string) OrderType {
	switch s {
	case string(TypeBuy):
		return TypeBuy
	case string(TypeSell):
		return TypeSell
	default:
		return TypeNil
	}
}
