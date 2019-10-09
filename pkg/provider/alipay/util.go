package alipay

import (
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

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

func conevertType(t TxTypeType) ir.TxType {
	switch t {
	case TxTypeSend:
		return ir.TxTypeSend
	case TxTypeRecv:
		return ir.TxTypeRecv
	default:
		return ir.TxTypeUnknown
	}
}
