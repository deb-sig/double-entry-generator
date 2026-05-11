package cib_debit

import "time"

const (
	dateTimeLayout  = "2006-01-02 15:04:05"
	dateLayout      = "2006-01-02"
	beijingOffset   = 8 * 60 * 60
	providerSource  = "兴业银行借记卡"
	providerPeer    = "CIB Debit"
	defaultCurrency = "CNY"
)

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

type Order struct {
	TradeTime     string
	AccountingDay string
	Expense       string
	Income        string
	Balance       string
	Summary       string
	Peer          string
	PeerBank      string
	PeerAccount   string
	Purpose       string
	Channel       string
	Remark        string
	AccountName   string
	AccountNum    string
	SubAccount    string
	Currency      string
}
