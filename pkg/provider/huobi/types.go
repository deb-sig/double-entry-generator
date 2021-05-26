package huobi

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

type Order struct {
	PayTime        time.Time // 付款时间
	Type           OrderType // 类型
	TypeOriginal   string    // 原始类型
	Item           string    // 交易对
	TxType         TxType    // 方向
	Price          float64   // 价格
	Amount         float64   // 数量
	Money          float64   // 成交额
	Commission     float64   // 手续费
	CommissionUnit string    // 手续费单位
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// OrderType is the type of the order.
type OrderType string

const (
	OrderTypeCoin    OrderType = "币币交易"
	OrderTypeUnknown           = "未知"
	// TODO(TripleZ): add more order types.
)

type TxType string

const (
	TxTypeBuy  TxType = "买入"
	TxTypeSell TxType = "卖出"
	TxTypeNil  TxType = "未知"
)
