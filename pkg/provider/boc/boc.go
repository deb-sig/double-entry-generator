package boc

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

type Boc struct {
	Statistics  Statistics `json:"statistics,omitempty"`
	LineNum     int        `json:"line_num,omitempty"`
	Orders      []Order    `json:"orders,omitempty"`
	HeaderFound bool       `json:"header_found,omitempty"`
	CardName    string     `json:"card_name,omitempty"`
	Mode        CardMode   `json:"mode,omitempty"`
	TitleParsed bool       `json:"title_parsed,omitempty"`
}

func New() *Boc {
	return &Boc{
		Statistics:  Statistics{},
		LineNum:     0,
		Orders:      make([]Order, 0),
		HeaderFound: false,
		CardName:    "",
		Mode:        DebitMode,
	}
}

// Translate the boc bill records to IR.
func (boc *Boc) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Boc] ")
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
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if !boc.HeaderFound {
			if len(line) == 0 || !strings.HasPrefix(line[0], "中国银行") {
				continue
			}
			boc.HeaderFound = true
			if strings.Contains(line[0], "信用卡") {
				boc.Mode = CreditMode
			}
		}
		boc.LineNum++

		if boc.LineNum <= 2 {
			continue
		}

		if boc.Mode == DebitMode {
			err = boc.TranslateToDebitOrders(line)
		} else {
			err = boc.TranslateToCreditOrders(line)
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: data line %d: error: %v", boc.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	ir := boc.convertToIR()
	return ir, nil

}
