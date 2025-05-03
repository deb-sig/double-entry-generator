package mt

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// MT is the provider for Meituan
type MT struct {
	Statistics  Statistics `json:"statistics,omitempty"`
	LineNum     int        `json:"line_num,omitempty"`
	HeaderFound bool       `json:"header_found,omitempty"`
	Orders      []Order    `json:"orders,omitempty"`

	// TitleParsed is a workaround to ignore the title row.
	TitleParsed bool `json:"title_parsed,omitempty"`
}

func New() *MT {
	return &MT{
		Statistics:  Statistics{},
		LineNum:     0,
		Orders:      make([]Order, 0),
		TitleParsed: false,
	}
}

func (mt *MT) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-MT] ")
	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true // 可以处理不规范的引号
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	csvReader.FieldsPerRecord = -1

	for {
		line, err := csvReader.Read()
		if err == io.EOF { // 文件读取结束
			break
		} else if err != nil {
			return nil, err
		}
		if !mt.HeaderFound {
			if len(line) == 0 || !strings.HasPrefix(line[0], "交易创建时间") {
				continue
			}
			mt.HeaderFound = true
		}
		mt.LineNum++
		if mt.LineNum < 2 { // 跳过以 "交易创建时间" 开头的行以及之前的内容
			continue
		}

		err = mt.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: data line %d: %v", mt.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	ir := mt.convertToIR()
	return ir, nil
}
