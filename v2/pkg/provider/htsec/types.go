package htsec

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
	SecuritiesName    string    // 证券代码+证券名称
	TransactionTime   time.Time // 成交时间
	Volume            int64     // 成交数量
	Price             float64   // 成交价格
	TransactionAmount float64   // 成交金额
	OccurAmount       float64   // 发生金额
	Type              OrderType // 方向
	OrderID           string    // 合同号
	TransactionID     string    // 成交号
	Commission        float64   // 手续费
	StampDuty         float64   // 印花税
	TransferFee       float64   // 过户费
	OtherFee          float64   // 其他费
	RemainAmount      float64   // 资金余额
	RemainShare       int64     // 份额余额
	TxTypeOriginal    string    // 市场+证券代码
	Useless           bool      // 需删除数据
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// OrderType is the type of the order.
type OrderType string

const (
	TxTypeBuy  OrderType = "买"
	TxTypeSell OrderType = "卖"
	TxTypeNil  OrderType = "未知"
)
