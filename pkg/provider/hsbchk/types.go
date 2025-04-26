package hsbchk

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
	PostDate        time.Time // 入账日期 (仅信用卡)
	Description     string    // 交易描述
	Merchant        string    // 商户名称 (仅信用卡)
	Country         string    // 国家/地区
	Money           float64   // 交易金额
	Currency        string    // 货币
	Balance         float64   // 余额 (仅借记卡)
	BalanceCurrency string    // 余额货币 (仅借记卡)
	Type            OrderType // 收/支
	StatusOriginal  string    // 交易状态 (仅信用卡)
	CreditDebit     string    // CREDIT/DEBIT (仅信用卡)
}

// TimeFormat 设置时间格式
const TimeFormat = "02/01/2006"

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
