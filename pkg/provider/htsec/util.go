package htsec

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
