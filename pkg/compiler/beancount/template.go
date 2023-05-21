package beancount

import (
	"text/template"
	"time"
)

// 普通账单的模版（消费账）
var normalOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ EscapeString .Peer }}" {{- if .Item }} "{{ EscapeString .Item }}"{{ end }}{{ range .Tags }} #{{ . }}{{ end }}{{ if .Note }} ; {{ .Note }}{{ end }}
	{{- range $key, $value := .Metadata }}{{ if $value }}{{ printf "\n" }}	{{ $key }}: "{{ $value }}"{{end}}{{end}}
	{{ .PlusAccount }} {{ .Money | printf "%.2f" }} {{ .Currency }}
	{{ .MinusAccount }} -{{ .Money | printf "%.2f" }} {{ .Currency }}
	{{- if .CommissionAccount }}{{ printf "\n" }}	{{ .CommissionAccount }} {{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
	{{- if .CommissionAccount }}{{ printf "\n" }}	{{ .MinusAccount }} -{{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
	{{- if .PnlAccount }}{{ printf "\n" }}	{{ .PnlAccount }}{{ end }}

`

type NormalOrderVars struct {
	PayTime           time.Time
	Peer              string
	Item              string
	Note              string
	Money             float64
	Commission        float64
	PlusAccount       string
	MinusAccount      string
	PnlAccount        string
	CommissionAccount string
	Currency          string
	Metadata          map[string]string // unordered metadata map
	Tags              []string
}

// 火币买入模版（手续费单位为购买单位货币）
var huobiTradeBuyOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TxTypeOriginal }}" "{{ .TypeOriginal }}-{{ .Item }}"
	{{ .CashAccount }} -{{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} {{ .Amount | printf "%.8f" }} {{ .TargetUnit }} { {{- .Price | printf "%.8f" }} {{ .BaseUnit -}} } @@ {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} -{{ .Commission | printf "%.8f" }} {{ .TargetUnit }} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}

`

// 火币买入模版 2（手续费为特定货币）
var huobiTradeBuyOrderDiffCommissionUnit = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TxTypeOriginal }}" "{{ .TypeOriginal }}-{{ .Item }}"
	{{ .CashAccount }} -{{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} {{ .Amount | printf "%.8f" }} {{ .TargetUnit }} { {{- .Price | printf "%.4f" }} {{ .BaseUnit -}} } @@ {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .PositionAccount }} -{{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}

`

type HuobiTradeBuyOrderVars struct {
	PayTime           time.Time
	Peer              string
	TxTypeOriginal    string
	TypeOriginal      string
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

// 火币卖出模版
var huobiTradeSellOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}-{{ .TxTypeOriginal }}" "{{ .TypeOriginal }}-{{ .Item }}"
	{{ .PositionAccount }} -{{ .Amount | printf "%.8f" }} {{ .TargetUnit }} {} @ {{ .Price | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} {{ .Money | printf "%.8f" }} {{ .BaseUnit }}
	{{ .CashAccount }} -{{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.8f" }} {{ .CommissionUnit }}
	{{ .PnlAccount }}

`

type HuobiTradeSellOrderVars struct {
	PayTime           time.Time
	Peer              string
	TxTypeOriginal    string
	TypeOriginal      string
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

// 海通买入模版
var htsecTradeBuyOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}" "{{ .TypeOriginal }}-{{ .Item }}"
	{{ .CashAccount }} -{{ .Money | printf "%.2f" }} {{ .Currency }}
	{{ .PositionAccount }} {{ .Amount | printf "%.2f" }} {{ .TxTypeOriginal }} { {{- .Price | printf "%.3f" }} {{ .Currency }}} @@ {{ .Money | printf "%.2f" }} {{ .Currency }}
	{{ .CashAccount }} -{{ .Commission | printf "%.2f" }} {{ .Currency }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.2f" }} {{ .Currency }}

`

type HtsecTradeBuyOrderVars struct {
	PayTime           time.Time
	Peer              string
	TxTypeOriginal    string
	TypeOriginal      string
	Item              string
	CashAccount       string
	PositionAccount   string
	CommissionAccount string
	PnlAccount        string
	Amount            float64
	Money             float64
	Commission        float64
	Price             float64
	Currency          string
}

var htsecTradeSellOrder = `{{ .PayTime.Format "2006-01-02" }} * "{{ .Peer }}" "{{ .TypeOriginal }}-{{ .Item }}"
	{{ .PositionAccount }} -{{ .Amount | printf "%.2f" }} {{ .TxTypeOriginal }} {} @ {{ .Price | printf "%.3f" }} {{ .Currency }}
	{{ .CashAccount }} {{ .Money | printf "%.2f" }} {{ .Currency }}
	{{ .CashAccount }} -{{ .Commission | printf "%.2f" }} {{ .Currency }}
	{{ .CommissionAccount }} {{ .Commission | printf "%.2f" }} {{ .Currency }}
	{{ .PnlAccount }}

`

type HtsecTradeSellOrderVars struct {
	PayTime           time.Time
	Peer              string
	TxTypeOriginal    string
	TypeOriginal      string
	Item              string
	CashAccount       string
	PositionAccount   string
	CommissionAccount string
	PnlAccount        string
	Amount            float64
	Money             float64
	Commission        float64
	Price             float64
	Currency          string
}

var (
	normalOrderTemplate                          *template.Template
	huobiTradeBuyOrderTemplate                   *template.Template
	huobiTradeBuyOrderDiffCommissionUnitTemplate *template.Template
	huobiTradeSellOrderTemplate                  *template.Template
	htsecTradeBuyOrderTemplate                   *template.Template
	htsecTradeSellOrderTemplate                  *template.Template
)
