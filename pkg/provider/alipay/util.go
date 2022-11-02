package alipay

func getTxType(str string) Type {
	switch str {
	case string(TypeSend):
		return TypeSend
	case string(TypeRecv):
		return TypeRecv
	case string(TypeOthers):
		return TypeOthers
	case string(TypeEmpty):
		return TypeEmpty
	}
	return TypeNil
}
