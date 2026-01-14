package boc

import (
	"time"
)

const (
	// localTimeFmt set time format to utc+8
	localTimeFmt = "2006-01-02 15:04:05 -0700 CST"
)

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

type Order struct {
	// 记账日期,记账时间,币别,金额,余额,交易名称,渠道,网点名称,附言,对方账户名,对方卡号/账号,对方开户行,借记卡号
	// 这里的 json 控制的是输出的 beancount 的字段
	// CreateTime   time.Time `json:"createTime,omitempty"`
	PayTime    time.Time `json:"payTime,omitempty"`
	Currency   string    `json:"currency,omitempty"`   // 币种
	Money      float64   `json:"money,omitempty"`      // 记账金额（收入/支出）
	Type       OrderType `json:"type,omitempty"`       // 收入/支出（由金额正负推断而来）
	ItemName   string    `json:"itemName,omitempty"`   // 交易名称/交易描述
	Channel    string    `json:"channel,omitempty"`    // 渠道
	Branch     string    `json:"branch,omitempty"`     // 网点
	Postscript string    `json:"postscript,omitempty"` // 附言
	PeerName   string    `json:"peerName,omitempty"`   // 对方账户名
	PeerCard   string    `json:"peerCard,omitempty"`   // 对方卡号
	PeerBank   string    `json:"peerBank,omitempty"`   // 对方开户行
	Method     string    `json:"method,omitempty"`     // 卡号末四位
}

// type CreditOrder struct {
// 	// 交易日,银行记账日,卡号后四位,交易描述,存入,支出,币种
// 	TradeDate time.Time `json:"tradeDate,omitempty"` // 交易日期
// 	PostDate  time.Time `json:"postDate,omitempty"`  // 入账日期
// 	Method    string    `json:"method,omitempty"`    // 卡号后四位
// 	TradeDesc string    `json:"tradeDesc,omitempty"` // 交易描述
// 	Expense   float64   `json:"expense,omitempty"`   // 支出
// 	Income    float64   `json:"income,omitempty"`    // 存入
// 	Currency  string    `json:"currency,omitempty"`  // 币种
// }

type CardMode string

const (
	DebitMode  CardMode = "Debit"
	CreditMode CardMode = "Credit"
)

type OrderType string

const (
	TypeSend   OrderType = "支出"
	TypeRecv   OrderType = "收入"
	TypeOthers OrderType = "不计收支"
	TypeEmpty  OrderType = ""
	TypeNil    OrderType = "未知"
)
