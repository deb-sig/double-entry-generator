package wechat

import "time"

// Statistics is the Statistics of the bill file.
type Statistics struct {
	UserID          string    `json:"user_id,omitempty"`
	Username        string    `json:"username,omitempty"`
	ParsedItems     int       `json:"parsed_items,omitempty"`
	Start           time.Time `json:"start,omitempty"`
	End             time.Time `json:"end,omitempty"`
	TotalInRecords  int       `json:"total_in_records,omitempty"`
	TotalInMoney    float64   `json:"total_in_money,omitempty"`
	TotalOutRecords int       `json:"total_out_records,omitempty"`
	TotalOutMoney   float64   `json:"total_out_money,omitempty"`
}

// Order is the single order.
type Order struct {
	OrderID        string    // 交易号
	MechantOrderID string    // 商家订单号
	PayTime        time.Time // 付款时间
	Type           OrderType // 收/支
	TypeOriginal   string
	Peer           string  // 交易对方
	Item           string  // 商品名称
	Money          float64 // 金额
	TxType         TxType  // 交易类型
	TxTypeOriginal string
	Status         string  // 交易状态
	Method         string  // 支付方式
	Commission     float64 // 服务费
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// OrderType is the type of the order.
type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv              = "收入"
	OrderTypeNil               = "/"
	OrderTypeUnknown           = "Unknown"
)

type TxType string

const (
	TxTypeConsume      TxType = "商户消费"
	TxTypeLucky               = "微信红包"
	TxTypeTransfer            = "转账"
	TxTypeQRIncome            = "二维码收款"
	TxTypeQRSend              = "扫二维码付款"
	TxTypeGroup               = "群收款"
	TxTypeRefund              = "退款"
	TxTypeCash2Cash           = "转入零钱通-来自零钱"
	TxTypeIntoCash            = "转入零钱通"
	TxTypeCashWithdraw        = "零钱提现"
	TxTypeUnknown             = "Unknown"
)
