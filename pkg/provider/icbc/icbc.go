package icbc

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// Icbc is the provider for Icbc.
type Icbc struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	Mode       CardMode   `json:"mode,omitempty"`
	CardName   string     `json:"card_name,omitempty"`
}

// New creates a new ICBC provider.
func New() *Icbc {
	return &Icbc{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		Mode:       CreditMode,
		CardName:   "",
	}
}

// Translate translates the icbc bill records to IR.
func (icbc *Icbc) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-ICBC] ")

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

		icbc.LineNum++
		if icbc.LineNum == 2 && len(line) > 1 {
			// 卡别名
			icbc.CardName = strings.TrimLeft(line[1], "卡别名: ")
			continue
		} else if icbc.LineNum == 3 {
			// 借记卡 or 信用卡(default)
			for _, col := range line {
				if strings.TrimLeft(col, "子账户类别: ") == "活期" {
					icbc.Mode = DebitMode
				}
			}
			log.Printf("Now the ICBC provider is in `%s` mode", icbc.Mode)
			continue
		} else if icbc.LineNum <= 5 {
			// The first 5 non-empty lines are useless for us.
			continue
		}

		if line[0] == "合计金额" || line[0] == "人民币合计" {
			// ignore the last line
			break
		}

		err = icbc.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				icbc.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return icbc.convertToIR(), nil
}
