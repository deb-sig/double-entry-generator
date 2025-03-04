package cmb

import "time"

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

type DebitOrder struct {
	Date              time.Time // 记账日期
	Currency          string    // 货币
	TransactionAmount float64   // 交易金额
	Balance           float64   // 联机余额
	TransactionType   string    // 交易摘要
	CounterParty      string    // 对手信息（ir.peer）
	CustomerType      string    // 客户摘要（ir.item）
	Type              OrderType // 收/支 (数据中无该列，推测而来)
}

type CreditOrder struct {
	SoldDate           *time.Time // 交易日
	PostedDate         time.Time  // 记账日
	Description        string     // 交易摘要
	RmbAmount          float64    // 人民币金额
	CardNo             string     // 卡号末四位
	OriginalTranAmount float64    // 交易地金额
	Type               OrderType  // 收/支 (数据中无该列，推测而来)
}

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
)

type CardMode string

const (
	CardModeDebit   CardMode = "Debit"
	CardModeCredit  CardMode = "Credit"
	CardModeUnknown CardMode = "Unknown"
)

// localTimeFmt set time format to utc+8
const localTimeFmt = "2006-01-02 +0800 CST"

var allDebitHeaders = [...]string{"记账日期", "货币", "交易金额", "联机余额", "交易摘要", "对手信息", "客户摘要"}
var allCreditHeaders = [...]string{"交易日", "记账日", "交易摘要", "人民币金额", "卡号末四位", "交易地金额"}
