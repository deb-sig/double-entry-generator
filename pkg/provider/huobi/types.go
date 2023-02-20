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
	TxType         TxType    // 类型
	TxTypeOriginal string    // 原始类型
	Item           string    // 交易对
	Type           OrderType // 方向
	Price          float64   // 价格
	Amount         float64   // 数量
	Money          float64   // 成交额
	Commission     float64   // 手续费
	BaseUnit       string    // 基准单位
	TargetUnit     string    // 目标单位
	CommissionUnit string    // 手续费单位
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// TxType is the type of the order.
type TxType string

const (
	TxTypeCoin    TxType = "币币交易"
	TxTypeUnknown TxType = "未知"
	// TODO(TripleZ): add more order types.
)

type OrderType string

const (
	TypeBuy  OrderType = "买入"
	TypeSell OrderType = "卖出"
	TypeNil  OrderType = "未知"
)
