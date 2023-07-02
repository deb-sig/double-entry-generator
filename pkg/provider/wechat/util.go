package wechat

import "strings"

func getOrderType(ot string) OrderType {
	switch ot {
	case string(OrderTypeRecv):
		return OrderTypeRecv
	case string(OrderTypeSend):
		return OrderTypeSend
	case string(OrderTypeNil):
		return OrderTypeNil
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
	} else if strings.Contains(tt, string(TxTypeQRIncome)) {
		return TxTypeQRIncome
	} else if strings.Contains(tt, string(TxTypeQRSend)) {
		return TxTypeQRSend
	} else if strings.Contains(tt, string(TxTypeGroup)) {
		return TxTypeGroup
	} else if strings.Contains(tt, string(TxTypeRefund)) {
		return TxTypeRefund
	} else if strings.Contains(tt, string(TxTypeCash2Cash)) {
		return TxTypeCash2Cash
	} else if strings.Contains(tt, string(TxTypeIntoCash)) {
		return TxTypeIntoCash
	} else if strings.Contains(tt, string(TxTypeCashIn)) {
		return TxTypeCashIn
	} else if strings.Contains(tt, string(TxTypeCashWithdraw)) {
		return TxTypeCashWithdraw
	} else if strings.Contains(tt, string(TxTypeCreditCardRefund)) {
		return TxTypeCreditCardRefund
	} else if strings.Contains(tt, string(TxTypeBuyLiCaiTong)) {
		return TxTypeBuyLiCaiTong
	} else if strings.Contains(tt, string(TxTypeCash2CashLooseChange)) {
		return TxTypeCash2CashLooseChange
	} else if strings.Contains(tt, string(TxTypeCash2Others)) {
		return TxTypeCash2Others
	} else if strings.Contains(tt, string(TxTypeFamilyCard)) {
		return TxTypeFamilyCard
	} else if strings.Contains(tt, string(TxTypeSponsorCode)) {
		return TxTypeSponsorCode
	} else if strings.Contains(tt, string(TxTypeOther)){
		return TxTypeOther
	} else {
		return TxTypeUnknown
	}
}
