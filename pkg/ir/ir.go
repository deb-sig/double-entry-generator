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
	Peer            string
	Item            string
	MerchantOrderID *string
	OrderID         *string
	Money           float64
	PayTime         time.Time
	Type            TxType
	TypeOriginal    string
	Method          string

	MinusAccount string
	PlusAccount  string
}

// TxType is transanction type defined by alipay.
type TxType string

const (
	TxTypeSend    TxType = "Send"
	TxTypeRecv           = "Recv"
	TxTypeUnknown        = "Unknwon"
)

// New creates a new IR.
func New() *IR {
	return &IR{
		Orders: make([]Order, 0),
	}
}
