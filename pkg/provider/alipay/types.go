// Sean at shanghai

package alipay

import "time"

// Statistics is the Statistics of the bill file.
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

// Order is the single order.
type Order struct {
	DealNo      string          `json:"dealNo,omitempty"`      // 交易号
	OrderNo     string          `json:"orderNo,omitempty"`     // 商家订单号
	CreateTime  time.Time       `json:"createTime,omitempty"`  // 交易创建时间
	PayTime     time.Time       `json:"payTime,omitempty"`     // 付款时间
	LastUpdate  time.Time       `json:"lastUpdate,omitempty"`  // 最近修改时间
	DealSrc     string          `json:"dealSrc,omitempty"`     // 交易来源地
	Type        string          `json:"type,omitempty"`        // 类型
	Peer        string          `json:"peer,omitempty"`        // 交易对方
	ItemName    string          `json:"itemName,omitempty"`    // 商品名称
	Money       float64         `json:"money,omitempty"`       // 金额
	TxType      TxTypeType      `json:"txType,omitempty"`      // 收/支
	Status      string          `json:"status,omitempty"`      // 交易状态
	ServiceFee  float64         `json:"serviceFee,omitempty"`  // 服务费
	Refund      float64         `json:"refund,omitempty"`      // 成功退款
	Comment     string          `json:"comment,omitempty"`     // 备注
	MoneyStatus MoneyStatusType `json:"moneyStatus,omitempty"` // 资金状态
	// below is filled at runtime
	MinusAccount string `json:"minusAccount,omitempty"`
	PlusAccount  string `json:"plusAccount,omitempty"`
}

// LocalTimeFmt set time format to utc+8
const LocalTimeFmt = "2006-01-02 15:04:05 -0700"

// TxTypeType is transanction type defined by alipay.
type TxTypeType string

const (
	TxTypeSend  TxTypeType = "支出"
	TxTypeRecv  TxTypeType = "收入"
	TxTypeEmpty TxTypeType = ""
	TxTypeNil   TxTypeType = "未知"
)

// MoneyStatusType is the status for the transaction.
type MoneyStatusType string

const (
	MoneySend      MoneyStatusType = "已支出"
	MoneyRecv                      = "已收入"
	MoneyTransfer                  = "资金转移"
	MoneyFreeze                    = "冻结"
	MoneyUnfreeze                  = "解冻"
	MoneyStatusNil                 = "未知"
)
