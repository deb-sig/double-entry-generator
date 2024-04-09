/*
Copyright © 2024 CNLHC

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
package jd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

const (
	TypeSend    Type = "支出"
	TypeRecv    Type = "收入"
	TypeOthers  Type = "不计收支"
	TypeUnknown Type = "未知"
)

var (
	timeFormat = "2006-01-02 15:04:05 -0700 CST"
)

type (
	JD struct {
		LineNum int     `json:"line_num,omitempty"`
		Orders  []Order `json:"orders,omitempty"`

		// TitleParsed is a workaround to ignore the title row.
		TitleParsed bool `json:"title_parsed,omitempty"`
	}
	Type string

	Order struct {
		PayTime    time.Time `json:"payTime,omitempty"`  // 交易时间
		Category   string    `json:"category,omitempty"` // 交易分类
		Peer       string    `json:"peer,omitempty"`     // 商户名称
		PeerType   string    `json:"peerType,omitempty"`
		ItemName   string    `json:"itemName,omitempty"`   // 交易说明
		Type       Type      `json:"type,omitempty"`       // 收/支
		Money      int64     `json:"money,omitempty"`      // 金额
		Method     string    `json:"method,omitempty"`     // 收/付款方式
		Status     string    `json:"status,omitempty"`     // 交易状态
		DealNo     string    `json:"dealNo,omitempty"`     // 交易订单号
		MerchantId string    `json:"merchantId,omitempty"` // 商家订单号
		Notes      string    `json:"notes,omitempty"`      // 交易备注

		// below is filled at runtime
		TargetAccount string `json:"targetAccount,omitempty"`
		MethodAccount string `json:"methodAccount,omitempty"`
	}
)

func New() *JD {
	return &JD{}

}

func (c *JD) Translate(fn string) (*ir.IR, error) {

	log.SetPrefix("[Provider-Alipay] ")
	r, err := reader.GetReader(fn)
	if err != nil {
		return nil, fmt.Errorf("can not get bill reader. %w", err)
	}
	csvReader := csv.NewReader(r)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1
	res := ir.New()

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("io error: %w", err)
		}
		c.LineNum++
		if c.LineNum < 21 {
			// skip header
			continue
		}
		err = c.translateLine(row)
		if err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %w ", c.LineNum, err)
		}
	}
	for _, o := range c.Orders {
		res.Orders = append(res.Orders, c.convertToIR(o))
	}
	return res, nil

}

func (c *JD) translateLine(row []string) error {
	var (
		bill Order
		err  error
	)

	if len(row) < 11 {
		return fmt.Errorf("row length is less than expected(11)")
	}
	for idx, a := range row {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		row[idx] = a
	}

	bill.PayTime, err = time.Parse(timeFormat, row[0]+" +0800 CST")
	if err != nil {
		return err
	}
	bill.Category = row[1]
	bill.PeerType = row[2]
	bill.ItemName = row[3]

	bill.Type = c.translateType(row[4])
	bill.Money, err = c.translateValue(row[5])
	if err != nil {
		return err
	}
	bill.Method = row[6]
	bill.Status = row[7]
	bill.DealNo = row[8]
	bill.MerchantId = row[9]
	bill.Notes = row[10]

	if bill.PeerType == "京东平台商户" {
		realPeer := strings.Split(bill.ItemName, " ")
		if !strings.HasPrefix(bill.ItemName, "退款") &&
			len(realPeer) > 0 &&
			len(realPeer[0]) < 15 {
			bill.Peer = realPeer[0]
		} else {
			bill.Peer = bill.PeerType
		}
	} else {
		bill.Peer = bill.PeerType
	}

	c.Orders = append(c.Orders, bill)

	return nil
}

func (c *JD) translateType(s string) Type {
	switch Type(s) {
	case TypeRecv:
		return TypeRecv
	case TypeSend:
		return TypeSend
	case TypeOthers:
		return TypeOthers
	default:
		return TypeUnknown
	}
}

func (c *JD) convertToIRType(s Type) ir.Type {
	switch s {
	case TypeRecv:
		return ir.TypeRecv
	case TypeSend:
		return ir.TypeSend
	default:
		return ir.TypeUnknown
	}
}

func (c *JD) translateValue(s string) (int64, error) {
	s = strings.ReplaceAll(s, ".", "")
	return strconv.ParseInt(s, 10, 64)
}

func (c *JD) convertToIR(s Order) ir.Order {
	return ir.Order{
		OrderType:       ir.OrderTypeNormal,
		Peer:            s.Peer,
		Item:            s.ItemName,
		Category:        s.Category,
		MerchantOrderID: &s.MerchantId,
		OrderID:         &s.DealNo,
		Money:           float64(s.Money) / 100.0,
		Note:            s.Notes,
		PayTime:         s.PayTime,
		Type:            c.convertToIRType(s.Type),
		Method:          s.Method,
		Metadata:        c.getMetadata(s),
		TypeOriginal:    string(s.Type),
	}

}

func (*JD) getMetadata(s Order) map[string]string {
	return map[string]string{
		"source":     s.Peer,
		"category":   s.Category,
		"payTime":    s.PayTime.Format(timeFormat),
		"orderId":    s.DealNo,
		"merchantId": s.MerchantId,
		"type":       string(s.Type),
		"method":     s.Method,
		"status":     s.Status,
	}
}
