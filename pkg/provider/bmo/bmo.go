package bmo

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

type Bmo struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	CardName   string     `json:"card_name,omitempty"`
	Mode       CardMode   `json:"mode,omitempty"`
}

func New() *Bmo {
	return &Bmo{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		CardName:   "",
		Mode:       DebitMode,
	}
}

// Translate the bmo bill records to IR.
func (bmo *Bmo) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-BMO] ")

	billReader, err := reader.GetReader(filename)
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

		bmo.LineNum++

		if bmo.LineNum == 2 {
			// 借记卡(default) or 信用卡
			if line[0] == "Item #" {
				bmo.Mode = CreditMode
				continue
			}
		}

		if bmo.Mode == DebitMode && bmo.LineNum <= 2 {
			// bypass the useless
			continue
		}

		if bmo.Mode == DebitMode {
			err = bmo.translateDebitToOrders(line)
			if err != nil {
				return nil, fmt.Errorf("Failed to translate debit bill: line %d: %v", bmo.LineNum, err)
			}
		} else {
			err = bmo.translateCreditToOrders(line)
			if err != nil {
				return nil, fmt.Errorf("Failed to translate credit bill: line %d: %v", bmo.LineNum, err)
			}
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return bmo.convertToIR(), nil
}
