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
	SerialNum                 string
	TransDate                 string
	TransTime                 string
	TradingType               string
	DrCr                      string
	TransAmount               float64
	Balance                   float64
	PaymentReceiptAccount     string
	PaymentReceiptAccountName string
	TradingPlace              string
	Abstract                  string
	PayTime                   time.Time
	Type                      OrderType
	Peer                      string
	Item                      string
}

type OrderType string

const (
	OrderTypeSend    OrderType = "支出"
	OrderTypeRecv    OrderType = "收入"
	OrderTypeUnknown OrderType = "Unknown"
)
