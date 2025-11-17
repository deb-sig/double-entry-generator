package bocomcredit

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// BocomCredit is the provider for Bank of Communications credit card statements.
type BocomCredit struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

// New creates a new BocomCredit provider.
func New() *BocomCredit {
	return &BocomCredit{
		Statistics: Statistics{},
		Orders:     make([]Order, 0),
	}
}

// Translate converts the CSV statement to IR.
func (bc *BocomCredit) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-BOCOMCredit] ")

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

		// Skip header row
		if bc.LineNum == 0 {
			firstField := strings.TrimSpace(record[0])
			if strings.EqualFold(firstField, "交易日期") {
				bc.LineNum++
				continue
			}
		}

		bc.LineNum++

		if err := bc.translateToOrders(record); err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %w", bc.LineNum, err)
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return bc.convertToIR(), nil
}
