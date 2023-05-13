package icbc

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
	PayTime         time.Time // 记账日期
	TxTypeOriginal  string    // 摘要
	Peer            string    // 交易场所
	Region          string    // 交易国家或地区简称
	Money           float64   // 记账金额 (收入/支出)
	Type            OrderType // 收/支 (数据中无该列，推测而来)
	Currency        string    // 记账币种
	Balances        float64   // 余额
	PeerAccountName string    // 对方户名
}

// localTimeFmt set time format to utc+8
const localTimeFmt = "2006-01-02 +0800 CST"

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
)

type CardMode string

const (
	DebitMode  CardMode = "Debit"
	CreditMode CardMode = "Credit"
)
