package td

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

type Td struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	CardName   string     `json:"card_name,omitempty"`
}

func New() *Td {
	return &Td{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		CardName:   "",
	}
}

// Translate the td bill records to IR.
func (td *Td) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-TD] ")

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

		td.LineNum++

		err = td.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", td.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return td.convertToIR(), nil
}
