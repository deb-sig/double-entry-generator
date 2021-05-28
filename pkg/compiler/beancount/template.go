package beancount

import (
	"text/template"
	"time"
)

// 普通账单的模版（消费账）
var normalOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}" "{{ .Item }}"
	{{ .PlusAccount }} {{ .Money | printf "%.2f" }} {{ .Currency }}
	{{ .MinusAccount }} -{{ .Money | printf "%.2f" }} {{ .Currency }}

`

type NormalOrderVars struct {
	PayTime      time.Time
	Peer         string
	Item         string
	Money        float64
	PlusAccount  string
	MinusAccount string
	Currency     string
}

// 火币买入模版（手续费单位为购买单位货币）
var huobiTradeBuyOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TypeOriginal }}" "{{ .TxTypeOriginal }}-{{ .Item }}"
	{{ .CashAccount }} -{{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} {{ .Amount | printf "%.8f" }} {{ .TargetUnit }} { {{- .Price | printf "%.8f" }} {{ .BaseUnit -}} } @@ {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} -{{ .Commission | printf "%.8f" }} {{ .TargetUnit }} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}

`

// 火币买入模版 2（手续费为特定货币）
var huobiTradeBuyOrderDiffCommissionUnit = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TypeOriginal }}" "{{ .TxTypeOriginal }}-{{ .Item }}"
	{{ .CashAccount }} -{{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} {{ .Amount | printf "%.8f" }} {{ .TargetUnit }} { {{- .Price | printf "%.4f" }} {{ .BaseUnit -}} } @@ {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} -{{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}

`

type HuobiTradeBuyOrderVars struct {
	PayTime           time.Time
	Peer              string
	TypeOriginal      string
	TxTypeOriginal    string
	Item              string
	CashAccount       string
	PositionAccount   string
	CommissionAccount string
	PnlAccount        string
	Amount            float64
	Money             float64
	Commission        float64
	Price             float64
	BaseUnit          string
	TargetUnit        string
	CommissionUnit    string
}

var huobiTradeSellOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TypeOriginal }}" "{{ .TxTypeOriginal }}-{{ .Item }}"
	{{ .PositionAccount }} -{{ .Amount | printf "%.8f" }} {{ .TargetUnit }} {} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} -{{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .PnlAccount }}

`

type HuobiTradeSellOrderVars struct {
	PayTime           time.Time
	Peer              string
	TypeOriginal      string
	TxTypeOriginal    string
	Item              string
	CashAccount       string
	PositionAccount   string
	CommissionAccount string
	PnlAccount        string
	Amount            float64
	Money             float64
	Commission        float64
	Price             float64
	BaseUnit          string
	TargetUnit        string
	CommissionUnit    string
}

var (
	normalOrderTemplate                          *template.Template
	huobiTradeBuyOrderTemplate                   *template.Template
	huobiTradeBuyOrderDiffCommissionUnitTemplate *template.Template
	huobiTradeSellOrderTemplate                  *template.Template
)
