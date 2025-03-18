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

package alipay

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// Alipay is the provider for alipay.
type Alipay struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`

	// TitleParsed is a workaround to ignore the title row.
	TitleParsed bool `json:"title_parsed,omitempty"`
}

// New creates a new Alipay provider.
func New() *Alipay {
	return &Alipay{
		Statistics:  Statistics{},
		LineNum:     0,
		Orders:      make([]Order, 0),
		TitleParsed: false,
	}
}

// Translate translates the alipay bill records to IR.
func (a *Alipay) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Alipay] ")

	billReader, err := reader.GetGBKReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	csvReader.FieldsPerRecord = -1

	for {
		line, err := csvReader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if a.LineNum == 0 && strings.Contains(line[0], "支付宝") {
			return nil, fmt.Errorf("可能为支付宝老版本 csv 账单，请使用 1.7.0 及之前的版本尝试转换")
		}

		a.LineNum++

		if a.LineNum <= 23 {
			// bypass the useless
			continue
		}

		err = a.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				a.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	ir := a.convertToIR()
	return a.postProcess(ir), nil
}

func (a *Alipay) postProcess(ir_ *ir.IR) *ir.IR {
	var orders []ir.Order
	for i := 0; i < len(ir_.Orders); i++ {
		var order = ir_.Orders[i]
		// found alipay refund tx
		// “退款成功”状态的交易记录的category未必都是“退款”，“退款成功”的饿了么订单记录的category是“餐饮美食”
		if order.Metadata["status"] == "退款成功" {
			for j := 0; j < len(ir_.Orders); j++ {
				// find the order corresponding to the refund
				// (different tx) && (prefix match) && (money equal)
				if i != j &&
					strings.HasPrefix(
						ir_.Orders[i].Metadata["orderId"],
						ir_.Orders[j].Metadata["orderId"]) &&
					ir_.Orders[i].Money == ir_.Orders[j].Money {
					log.Printf("[orderId %s] Refund for [orderId %s].",
						ir_.Orders[i].Metadata["orderId"],
						ir_.Orders[j].Metadata["orderId"])
					ir_.Orders[i].Metadata["useless"] = "true"
					ir_.Orders[j].Metadata["useless"] = "true"
				}
			}
		}
		// found alipay closed tx
		// “交易关闭”状态的交易记录的type未必都是“不计收支”，“交易关闭”的闲鱼卖出记录的type是“收入”
		if order.Metadata["status"] == "交易关闭" {
			ir_.Orders[i].Metadata["useless"] = "true"
			log.Printf("[orderId %s] canceled.",
				ir_.Orders[i].Metadata["orderId"])
		}
	}

	for _, v := range ir_.Orders {
		if v.Metadata["useless"] != "true" {
			if v.Metadata["status"] == "交易关闭" {
				log.Printf("[orderId %s] canceled tx left unprocessed.", v.Metadata["orderId"])
			}
			if v.Metadata["status"] == "退款成功" {
				log.Printf("[orderId %s] refund tx left unprocessed.", v.Metadata["orderId"])
			}
			orders = append(orders, v)
		}
	}
	ir_.Orders = orders
	// 超时
	return ir_
}
