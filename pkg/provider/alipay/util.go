package alipay

func getMoneyStatus(str string) MoneyStatusType {
	switch str {
	case string(MoneySend):
		return MoneySend
	case string(MoneyRecv):
		return MoneyRecv
	case string(MoneyTransfer):
		return MoneyTransfer
	case string(MoneyFreeze):
		return MoneyFreeze
	case string(MoneyUnfreeze):
		return MoneyUnfreeze
	default:
		return MoneyStatusNil
	}
}

func getTxType(str string) TxTypeType {
	switch str {
	case string(TxTypeSend):
		return TxTypeSend
	case string(TxTypeRecv):
		return TxTypeRecv
	case string(TxTypeEmpty):
		return TxTypeEmpty
	}
	return TxTypeNil
}
