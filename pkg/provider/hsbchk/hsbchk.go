package hsbchk

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// HsbcHK is the provider for HSBC HK.
type HsbcHK struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	Mode       CardMode   `json:"mode,omitempty"`
}

// New creates a new HSBC HK provider.
func New() *HsbcHK {
	return &HsbcHK{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		Mode:       CreditMode, // 默认使用信用卡模式
	}
}

// Translate translates the HSBC HK bill records to IR.
func (h *HsbcHK) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-HSBCHK] ")

	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	csvReader.FieldsPerRecord = -1

	var headers []string
	for {
		line, err := csvReader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		h.LineNum++
		if h.LineNum == 1 {
			// 保存标题行用于判断账单类型
			headers = line
			h.detectCardMode(headers)
			continue
		}

		// 跳过空行
		if len(line) == 0 || (len(line) == 1 && line[0] == "") {
			continue
		}

		err = h.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %v",
				h.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return h.convertToIR(), nil
}
