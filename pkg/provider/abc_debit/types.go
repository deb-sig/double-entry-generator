package abc_debit

import "time"

const (
	dateLayout      = "20060102"
	shortTimeLayout = "150405"
	beijingOffset   = 8 * 60 * 60
	providerPeer    = "ABC Debit"
	providerSource  = "中国农业银行储蓄卡"
	defaultCurrency = "CNY"
)

// Statistics keeps parsed summary information.
type Statistics struct {
	ParsedItems     int       `json:"parsed_items,omitempty"`
	Start           time.Time `json:"start,omitempty"`
	End             time.Time `json:"end,omitempty"`
	TotalInRecords  int       `json:"total_in_records,omitempty"`
	TotalInMoney    float64   `json:"total_in_money,omitempty"`
	TotalOutRecords int       `json:"total_out_records,omitempty"`
	TotalOutMoney   float64   `json:"total_out_money,omitempty"`
}

// OrderType indicates the direction of the transaction.
type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "未知"
)

// Order represents a parsed CSV row.
type Order struct {
	TradeDate  string // 交易日期
	TradeTime  string // 交易时间
	Summary    string // 交易摘要
	Amount     string // 交易金额原始值
	Balance    string // 本次余额
	Peer       string // 对手信息
	LogNumber  string // 日志号
	Channel    string // 交易渠道
	Postscript string // 交易附言
}
