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
	OrderType       OrderType
	Peer            string
	Item            string
	Category        string
	MerchantOrderID *string
	OrderID         *string
	Money           float64
	Note            string
	PayTime         time.Time
	Type            Type // 方向，一般为 收/支
	TypeOriginal    string
	TxTypeOriginal  string // 交易类型
	Method          string
	Amount          float64
	Price           float64
	Commission      float64 // 手续费/服务费
	Units           map[Unit]string
	ExtraAccounts   map[Account]string
	MinusAccount    string
	PlusAccount     string
	Metadata        map[string]string
	Tags            []string
}

// Unit is the key commodity names
type Unit string

const (
	BaseUnit       Unit = "BaseUnit"
	TargetUnit     Unit = "TargetUnit"
	CommissionUnit Unit = "CommissionUnit"
)

// Account is the key for account names
type Account string

const (
	CashAccount       Account = "CashAccount"
	PositionAccount   Account = "PositionAccount"
	CommissionAccount Account = "CommissionAccount"
	PnlAccount        Account = "PnlAccount"
	PlusAccount       Account = "PlusAccount"
	MinusAccount      Account = "MinusAccount"
)

// Type is transaction type defined by alipay.
type Type string

const (
	TypeSend    Type = "Send"
	TypeRecv    Type = "Recv"
	TypeUnknown Type = "Unknwon"
)

type OrderType string // 为 IR 设置的交易类别

const (
	OrderTypeNormal          OrderType = "Normal"          // 流水交易
	OrderTypeHuobiTrade      OrderType = "HuobiTrade"      //  火币交易
	OrderTypeSecuritiesTrade OrderType = "SecuritiesTrade" // 证券交易
)

// New creates a new IR.
func New() *IR {
	return &IR{
		Orders: make([]Order, 0),
	}
}
