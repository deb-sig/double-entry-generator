package citic

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
	TradeTime time.Time // 交易日期
	PostTime  time.Time // 入账日期
	TradeDesc string    // 交易描述
	Method    string    // 卡末四位
	Currency  string    // 记账币种
	Money     float64   // 记账金额 (收入/支出)
	Type      OrderType // 收/支 (数据中无该列，推测而来)
}

// localTimeFmt set time format to utc+8
const localTimeFmt = "2006-01-02 +0800 CST"

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
)
