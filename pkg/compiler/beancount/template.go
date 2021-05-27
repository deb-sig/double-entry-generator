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

// 金融投资模版（投资账）
var tradeBuyOrder = ``
var tradeSellOrder = ``

var (
	normalOrderTemplate    *template.Template
	tradeBuyOrderTemplate  *template.Template
	tradeSellOrderTemplate *template.Template
)
