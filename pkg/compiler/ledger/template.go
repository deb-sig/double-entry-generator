package ledger

import (
	"html/template"
	"time"
)

// 与 beancount相比, Ledger 格式简单许多, reference:
// - https://ledger-cli.org/doc/ledger3.html
// - https://devhints.io/ledger
/*
2013/01/03 * Rent for January
  ; comment
  Expenses:Rent   $600.00
  Assets:Savings
*/

// 普通账单的模版（消费账）
var normalOrder = `{{ .PayTime.Format "2006-01-02" }} * {{ EscapeString .Peer }} {{- if .Item }} - {{ EscapeString .Item }} {{ end }}
    {{- if .Note}}; {{ .Note }}{{ end }}
    {{- range $key, $value := .Metadata }}{{ if $value }}{{ printf "\n" }}    ; {{ $key }}: "{{ $value }}"{{end}}{{end}}
    {{ .PlusAccount }}      {{ .Amount | printf "%.2f" }} {{ .Currency }}
    {{ .MinusAccount }}   - {{ .Amount | printf "%.2f" }} {{ .Currency }}
    {{- if .CommissionAccount }}{{ printf "\n" }}    {{ .CommissionAccount }}      {{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
    {{- if .CommissionAccount }}{{ printf "\n" }}    {{ .MinusAccount }}   - {{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
    {{- if .PnlAccount }}{{ printf "\n" }}	{{ .PnlAccount }}{{ end }}

`

type NormalOrderVars struct {
	PayTime           time.Time         // 支付时间
	Peer              string            // 交易对手
	Item              string            // 交易商品
	Note              string            // 说明
	Amount            float64           // 金额
	Commission        float64           // 手续费
	PlusAccount       string            // 入账账户
	MinusAccount      string            // 支出账户
	PnlAccount        string            //
	CommissionAccount string            // 佣金账户
	Metadata          map[string]string // 元数据
	Currency          string            // 货币
}

var (
	normalOrderTemplate *template.Template
)

// 火币买入模板(TODO)

// 火币买入模板2(TODO)

// 火币卖出模版(TODO)

// 海通买入模版(TODO)
