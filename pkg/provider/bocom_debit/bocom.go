package bocom_debit

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// BocomDebit is the provider for Bank of Communications debit card statements.
type BocomDebit struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	Currency   string     `json:"currency,omitempty"`
}

// New creates a new BocomDebit provider.
func New() *BocomDebit {
	return &BocomDebit{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		Currency:   "CNY",
	}
}

// Translate converts the BocomDebit CSV statement into IR orders.
func (b *BocomDebit) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-BOCOM-DEBIT] ")

	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		b.LineNum++

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		// Skip header row(s)
		if len(row) > 0 {
			firstCell := strings.TrimSpace(row[0])
			// CSV files exported from Office suites may prefix the first cell with a UTF-8 BOM.
			firstCell = strings.TrimPrefix(firstCell, "\ufeff")
			if firstCell == "序号" {
				continue
			}
		}

		if err := b.translateLine(row); err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %v", b.LineNum, err)
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return b.convertToIR(), nil
}
