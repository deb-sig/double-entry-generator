package cgb_credit

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// CgbCredit 是广发银行信用卡账单 provider。
type CgbCredit struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

// New 创建广发银行信用卡 provider。
func New() *CgbCredit {
	return &CgbCredit{
		Statistics: Statistics{},
		Orders:     make([]Order, 0),
	}
}

// Translate 将广发信用卡 CSV 账单转换为 IR。
func (cc *CgbCredit) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-CGBCredit] ")

	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(record) == 0 {
			continue
		}

		cc.LineNum++
		raw, err := cc.parseRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record: line %d: %w", cc.LineNum, err)
		}
		if raw == nil {
			continue
		}

		order, err := cc.convertRawRecord(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record: line %d: %w", cc.LineNum, err)
		}

		cc.Orders = append(cc.Orders, *order)
		cc.updateStatistics(order)
	}

	log.Printf("Finished to parse the file %s", filename)
	return cc.convertToIR(), nil
}
