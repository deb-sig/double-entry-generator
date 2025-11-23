package abcdebit

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// AbcDebit is the provider for Agricultural Bank of China debit card statements.
type AbcDebit struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

// New creates a new AbcDebit provider.
func New() *AbcDebit {
	return &AbcDebit{
		Statistics: Statistics{},
		Orders:     make([]Order, 0),
	}
}

// Translate converts the CSV statement to IR.
func (ad *AbcDebit) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-ABCdebit] ")

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

		// Skip empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Skip header row
		if ad.LineNum == 0 {
			firstField := strings.TrimSpace(record[0])
			if strings.Contains(firstField, "交易日期") {
				ad.LineNum++
				continue
			}
		}

		ad.LineNum++

		if err := ad.translateToOrders(record); err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %w", ad.LineNum, err)
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return ad.convertToIR(), nil
}
