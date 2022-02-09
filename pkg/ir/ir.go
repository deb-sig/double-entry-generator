/*
Copyright © 2019 Ce Gao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ir

import (
	"time"
)

// IR is the intermediate representation for the double-entry bookkeeping.
type IR struct {
	// TODO(gaocegege): Refactor it to be general.
	Orders []Order
}

// Order is the intermediate representation for the order.
type Order struct {
	OrderType OrderType
	Peer      string
	Item      string
	Category  string
	// 是否未入账
	Pending         bool
	MerchantOrderID *string
	OrderID         *string
	Money           float64
	Note            string
	PayTime         time.Time
	TxType          TxType // 方向，一般为 收/支
	TxTypeOriginal  string
	TypeOriginal    string
	Method          string
	Amount          float64
	Price           float64
	Commission      float64 // 手续费/服务费
	Units           map[Unit]string
	ExtraAccounts   map[Account]string
	MinusAccount    string
	PlusAccount     string
	Metadata        map[string]string
}

// Unit is the key commodity names
type Unit string

const (
	BaseUnit       Unit = "BaseUnit"
	TargetUnit          = "TargetUnit"
	CommissionUnit      = "CommissionUnit"
)

// Account is the key for account names
type Account string

const (
	CashAccount       Account = "CashAccount"
	PositionAccount           = "PositionAccount"
	CommissionAccount         = "CommissionAccount"
	PnlAccount                = "PnlAccount"
	PlusAccount               = "PlusAccount"
	MinusAccount              = "MinusAccount"
)

// TxType is transanction type defined by alipay.
type TxType string

const (
	TxTypeSend    TxType = "Send"
	TxTypeRecv           = "Recv"
	TxTypeUnknown        = "Unknwon"
)

type OrderType string // 为 IR 设置的交易类别

const (
	OrderTypeNormal     OrderType = "Normal"     // 流水交易
	OrderTypeHuobiTrade           = "HuobiTrade" //  火币交易
)

// New creates a new IR.
func New() *IR {
	return &IR{
		Orders: make([]Order, 0),
	}
}
