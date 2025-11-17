package bocomcredit

import "time"

const (
	dateLayout   = "2006-01-02"
	providerPeer = "BOCOM Credit"
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
	TradeDate      time.Time
	RecordDate     time.Time
	Description    string
	Amount         float64
	Currency       string
	TxnAmount      float64
	TxnCurrency    string
	TxnAmountRaw   string
	Type           OrderType
	TypeOriginal   string
	TxTypeOriginal string
}
