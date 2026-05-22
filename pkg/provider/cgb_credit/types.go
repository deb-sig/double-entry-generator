package cgb_credit

import "time"

const (
	dateLayout   = "2006-01-02"
	inputLayout  = "2006/01/02"
	providerPeer = "CGB Credit"
)

// Statistics 记录广发信用卡账单解析统计信息。
type Statistics struct {
	ParsedItems     int       `json:"parsed_items,omitempty"`
	Start           time.Time `json:"start,omitempty"`
	End             time.Time `json:"end,omitempty"`
	TotalInRecords  int       `json:"total_in_records,omitempty"`
	TotalInMoney    float64   `json:"total_in_money,omitempty"`
	TotalOutRecords int       `json:"total_out_records,omitempty"`
	TotalOutMoney   float64   `json:"total_out_money,omitempty"`
}

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "未知"
)

// RawRecord 表示 bill-file-converter 从广发信用卡 PDF 转出的 CSV 原始字段。
type RawRecord struct {
	TradeDate      string // 交易日期
	RecordDate     string // 入账日期
	Description    string // 交易摘要
	TradeAmount    string // 交易金额
	TradeCurrency  string // 交易货币
	SettleAmount   string // 入账金额
	SettleCurrency string // 入账货币
}

// Order 表示已解析并标准化后的广发信用卡交易。
type Order struct {
	TradeDate       time.Time
	RecordDate      time.Time
	Description     string
	Amount          float64
	Currency        string
	TradeAmount     float64
	TradeCurrency   string
	TradeAmountRaw  string
	SettleAmountRaw string
	Type            OrderType
	TypeOriginal    string
	TxTypeOriginal  string
}
