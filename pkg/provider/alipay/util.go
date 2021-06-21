package alipay

func getTxType(str string) TxTypeType {
	switch str {
	case string(TxTypeSend):
		return TxTypeSend
	case string(TxTypeRecv):
		return TxTypeRecv
	case string(TxTypeOthers):
		return TxTypeOthers
	case string(TxTypeEmpty):
		return TxTypeEmpty
	}
	return TxTypeNil
}
