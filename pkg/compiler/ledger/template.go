package ledger

import (
	"html/template"
	"time"
)

// 与 beancount相比, Ledger 格式简单许多, reference:
// - https://ledger-cli.org/doc/ledger3.html
// - https://devhints.io/ledger
/*
2013/01/03 * Rent for January (ledger 支持中文和空格，不需要给交易商品增加双引号)
  ; comment
  Expenses:Rent   $600.00
  Assets:Savings
*/

// 普通账单的模版（消费账）
var normalOrder = `{{ .PayTime.Format "2006-01-02" }} * {{ EscapeString .Peer }} {{- if .Item }} - {{ EscapeString .Item }} {{ end }}
    {{- if .Note}}; {{ .Note }}{{ end }}
    {{- range $key, $value := .Metadata }}{{ if $value }}{{ printf "\n" }}    ; {{ $key }}: "{{ $value }}"{{end}}{{end}}
    {{ .PlusAccount }}      {{ .Money | printf "%.2f" }} {{ .Currency }}
    {{ .MinusAccount }}   - {{ .Money | printf "%.2f" }} {{ .Currency }}
    {{- if .CommissionAccount }}{{ printf "\n" }}    {{ .CommissionAccount }}      {{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
    {{- if .CommissionAccount }}{{ printf "\n" }}    {{ .MinusAccount }}   - {{ .Commission | printf "%.2f" }} {{ .Currency }}{{ end }}
    {{- if .PnlAccount }}{{ printf "\n" }}	{{ .PnlAccount }}{{ end }}

`

type NormalOrderVars struct {
	PayTime           time.Time         // 支付时间
	Peer              string            // 交易对手
	Item              string            // 交易商品
	Note              string            // 说明
	Money             float64           // 金额
	Commission        float64           // 手续费
	PlusAccount       string            // 入账账户
	MinusAccount      string            // 出账账户
	PnlAccount        string            //
	CommissionAccount string            // 佣金账户
	Metadata          map[string]string // 元数据
	Currency          string            // 货币
}

// 火币买入模版（手续费单位为购买单位货币）

/**
ledger 支持单价 * 数量, 如
; cost per item
2010/05/31 * Market
  Assets:Fridge                35 apples @ $0.42
  Assets:Cash

或者总价与数量(自动算出单价)
; total cost
2010/05/31 * Market
  Assets:Fridge                35 apples @@ $14.70
  Assets:Cash

但不能像beancount 那样子支持同时指定单价与总价.
2021-02-23 * "Huobi-币币交易" "买入-BTC1S/USDT"
	Assets:Rule1:Positions     4.57600000 "BTC1S" {1.22520000 "USDT" } @@ 5.60652159 "USDT"
	Assets:Cash

因为浮点数精度的原因, 4.57600000(数量) * 1.22520000(单价) = 5.6065152000000000, 而非 5.60652159, 就会导致对账不平

因此以总价及数量为标, 单价作为注释参考
2021-02-23 * "Huobi-币币交易" "买入-BTC1S/USDT"
	Assets:Rule1:Positions     4.57600000 "BTC1S" @@ 5.60652159 "USDT"; {1.22520000 "USDT" }
	Assets:Cash
**/

// 火币的货币中可能包含数字, 如BTC1S, ledger 包含数字的货币解析成金额，然后报错，因此需要使用双引号 "BTC1S"
var huobiTradeBuyOrder = `{{ .PayTime.Format "2006-01-02" }} * {{ .Peer }}-{{ .TxTypeOriginal }}-{{ .TypeOriginal }}-{{ .Item }}
    {{ .CashAccount }}     -{{ .Money | printf "%.8f" }} "{{ .BaseUnit }}"
    {{ .PositionAccount }}     {{ .Amount | printf "%.8f" }} "{{ .TargetUnit }}" @@ {{ .Money | printf "%.8f" }} "{{ .BaseUnit }}" ; { {{- .Price | printf "%.8f" }} "{{ .BaseUnit -}}" } 
    {{ .CashAccount }}     -{{ .Commission | printf "%.8f" }} "{{ .TargetUnit }}" @ {{ .Price | printf "%.8f" }} "{{ .BaseUnit }}"
    {{ .CommissionAccount }}     {{ .Commission | printf "%.8f" }} "{{ .CommissionUnit }}" @ {{ .Price | printf "%.8f" }} "{{ .BaseUnit }}"

`

type HuobiTradeBuyOrderVars struct {
	PayTime           time.Time // 交易时间
	Peer              string    // 交易对手
	TxTypeOriginal    string    // 交易类型(币币交易)
	TypeOriginal      string    // 操作类型(买入/卖出)
	Item              string    // 交易商品
	CashAccount       string    // 现金账号
	PositionAccount   string
	CommissionAccount string // 手续费账号
	PnlAccount        string
	Amount            float64 // 数量
	Money             float64 // 金额
	Commission        float64 // 手续费
	Price             float64 // 单价
	BaseUnit          string  // 支出货币类型
	TargetUnit        string  // 目标货币类型
	CommissionUnit    string  // 手续费货币类型
}

// 火币买入模版 2（手续费为特定货币）
var huobiTradeBuyOrderDiffCommissionUnit = `{{ .PayTime.Format "2006-01-02" }} * {{ .Peer }}-{{ .TxTypeOriginal }}-{{ .TypeOriginal }}-{{ .Item }}
    {{ .CashAccount }}     -{{ .Money | printf "%.8f" }} "{{ .BaseUnit }}"
    {{ .PositionAccount }}     {{ .Amount | printf "%.8f" }} "{{ .TargetUnit }}" @@ {{ .Money | printf "%.8f" }} "{{ .BaseUnit }}"; { {{- .Price | printf "%.4f" }} {{ .BaseUnit -}} }
    {{ .PositionAccount }}     -{{ .Commission | printf "%.8f" }} "{{ .CommissionUnit }}"
    {{ .CommissionAccount }}     {{ .Commission | printf "%.8f" }} "{{ .CommissionUnit }}"

`

// 火币卖出模版
var huobiTradeSellOrder = `{{ .PayTime.Format "2006-01-02" }} * {{ .Peer }}-{{ .TxTypeOriginal }}-{{ .TypeOriginal }}-{{ .Item }}
    {{ .PositionAccount }}     -{{ .Amount | printf "%.8f" }} "{{ .TargetUnit }}" @ {{ .Price | printf "%.8f" }} "{{ .BaseUnit }}"
    {{ .CashAccount }}     {{ .Money | printf "%.8f" }} "{{ .BaseUnit }}"
    {{ .CashAccount }}     -{{ .Commission | printf "%.8f" }} "{{ .CommissionUnit }}"
    {{ .CommissionAccount }}     {{ .Commission | printf "%.8f" }} "{{ .CommissionUnit }}"
    {{ .PnlAccount }}

`

type HuobiTradeSellOrderVars struct {
	PayTime           time.Time // 交易时间
	Peer              string    // 交易对手
	TxTypeOriginal    string    // 交易类型(币币交易)
	TypeOriginal      string    // 操作类型(买入/卖出)
	Item              string    // 交易商品
	CashAccount       string    // 现金账号
	PositionAccount   string
	CommissionAccount string // 手续费账号
	PnlAccount        string
	Amount            float64 // 数量
	Money             float64 // 金额
	Commission        float64 // 手续费
	Price             float64 // 单价
	BaseUnit          string  // 支出货币类型
	TargetUnit        string  // 目标货币类型
	CommissionUnit    string  // 手续费货币类型
}

// 海通买入模版(TODO)

var (
	normalOrderTemplate                          *template.Template
	huobiTradeBuyOrderTemplate                   *template.Template
	huobiTradeBuyOrderDiffCommissionUnitTemplate *template.Template
	huobiTradeSellOrderTemplate                  *template.Template
)
