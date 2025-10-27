package bocom_debit

import "time"

// Statistics captures aggregated information for the statement file.
type Statistics struct {
	ParsedItems     int       `json:"parsed_items,omitempty"`
	Start           time.Time `json:"start,omitempty"`
	End             time.Time `json:"end,omitempty"`
	TotalInRecords  int       `json:"total_in_records,omitempty"`
	TotalInMoney    float64   `json:"total_in_money,omitempty"`
	TotalOutRecords int       `json:"total_out_records,omitempty"`
	TotalOutMoney   float64   `json:"total_out_money,omitempty"`
}

// Order represents a single Bank of Communications transaction.
type Order struct {
	SerialNum                 string  // 序号
	TransDate                 string  // 交易日期
	TransTime                 string  // 交易时间
	TradingType               string  // 交易类型
	DcFlg                     string  // 借贷
	TransAmt                  float64 // 交易金额
	Balance                   float64 // 余额
	PaymentReceiptAccount     string  // 对方账号
	PaymentReceiptAccountName string  // 对方户名
	TradingPlace              string  // 交易地点
	Abstract                  string  // 摘要
}

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
)
