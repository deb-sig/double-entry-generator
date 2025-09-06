package ccb

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
	PayTime         time.Time // 交易日期
	TxTypeOriginal  string    // 摘要
	Peer            string    // 对方户名
	Item            string    // 交易详情
	Region          string    // 交易地点
	Money           float64   // 记账金额 (收入-支出)
	Type            OrderType // 收/支
	Currency        string    // 币种
	Balances        float64   // 账户余额
	PeerAccountName string    // 对方户名
	PeerAccountNum  string    // 对方账号
	Expense         float64   // 支出金额
	Income          float64   // 收入金额
	TradeTime       string    // 交易时间
	RecordDate      string    // 记账日
}

// localTimeFmt set time format to utc+8
const localTimeFmt = "20060102 +0800 CST"

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
) 