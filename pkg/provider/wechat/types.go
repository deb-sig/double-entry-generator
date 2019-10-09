package wechat

import "time"

// Statistics is the Statistics of the bill file.
type Statistics struct {
	UserID           string    `json:"user_id,omitempty"`
	Username         string    `json:"username,omitempty"`
	ParsedItems      int       `json:"parsed_items,omitempty"`
	Start            time.Time `json:"start,omitempty"`
	End              time.Time `json:"end,omitempty"`
	TotalInRecords   int       `json:"total_in_records,omitempty"`
	TotalInMoney     float64   `json:"total_in_money,omitempty"`
	TotoalOutRecords int       `json:"totoal_out_records,omitempty"`
	TotoalOutMoney   float64   `json:"totoal_out_money,omitempty"`
}

// Order is the single order.
type Order struct {
	OrderID        string    // 交易号
	MechantOrderID string    // 商家订单号
	PayTime        time.Time // 付款时间
	Type           OrderType // 类型
	TypeOriginal   string
	Peer           string  // 交易对方
	Item           string  // 商品名称
	Money          float64 // 金额
	TxType         TxType  // 收/支
	Status         string  // 交易状态
	Method         string  // 支付方式
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// OrderType is the type of the order.
type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv              = "收入"
	OrderTypeUnknown           = "Unknown"
)

type TxType string

const (
	TxTypeConsume  TxType = "商户消费"
	TxTypeLucky           = "微信红包"
	TxTypeTransfer        = "转账"
	TxTypeUnknown         = "Unknown"
)
