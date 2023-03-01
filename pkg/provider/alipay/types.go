// Sean at shanghai

package alipay

import "time"

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

// Order is the single order.
type Order struct {
	Type         Type      `json:"type,omitempty"` // 收/支
	TypeOriginal string    `json:"typeOriginal,omitempty"`
	Peer         string    `json:"peer,omitempty"`        // 交易对方
	PeerAccount  string    `json:"peerAccount,omitempty"` // 对方账号
	ItemName     string    `json:"itemName,omitempty"`    // 商品说明
	Method       string    `json:"method,omitempty"`      // 收/付款方式
	Money        float64   `json:"money,omitempty"`       // 金额
	Status       string    `json:"status,omitempty"`      // 交易状态
	Category     string    `json:"category,omitempty"`    // 交易分类
	DealNo       string    `json:"dealNo,omitempty"`      // 交易订单号
	MerchantId   string    `json:"merchantId,omitempty"`  // 商家订单号
	PayTime      time.Time `json:"payTime,omitempty"`     // 交易时间

	// below is filled at runtime
	TargetAccount string `json:"targetAccount,omitempty"`
	MethodAccount string `json:"methodAccount,omitempty"`
}

// Type is transanction type defined by alipay.
type Type string

const (
	TypeSend   Type = "支出"
	TypeRecv   Type = "收入"
	TypeOthers Type = "不计收支"
	TypeEmpty  Type = ""
	TypeNil    Type = "未知"
)
