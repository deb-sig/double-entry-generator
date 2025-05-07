package mt

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

type Order struct {
	// CreateTime   time.Time `json:"createTime,omitempty"`
	PayTime      time.Time `json:"payTime,omitempty"`
	TypeOriginal string    `json:"typeOriginal,omitempty"` // 交易类型（退款/支付）
	ItemName     string    `json:"itemName,omitempty"`     // 订单标题
	Type         Type      `json:"type,omitempty"`         // 收/支
	Method       string    `json:"method,omitempty"`       // 收/付款方式
	// BillMoney    float64   `json:"billMoney,omitempty"`    // 订单金额
	Money      float64 `json:"money,omitempty"`      // 实付金额
	DealNo     string  `json:"dealNo,omitempty"`     // 交易订单号
	MerchantId string  `json:"merchantId,omitempty"` // 商家订单号
	Note       string  `json:"note,omitempty"`       // 备注

	// below is filled at runtime
	TargetAccount string `json:"targetAccount,omitempty"` // 复式记账中的另一个帐号，日常消费中通常为 expenses 对应帐号
	MethodAccount string `json:"methodAccount,omitempty"` // 付款帐号，退款帐号
}

type Type string

const (
	TypeSend   Type = "支出"
	TypeRecv   Type = "收入"
	TypeOthers Type = "不计收支"
	TypeEmpty  Type = ""
	TypeNil    Type = "未知"
)
