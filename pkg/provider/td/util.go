package td

func getOrderType(withdrawal string, deposit string) OrderType {
	if len(withdrawal) > 0 {
		return OrderTypeSend
	} else if len(deposit) > 0 {
		return OrderTypeRecv
	} else {
		return OrderTypeUnknown
	}
}
