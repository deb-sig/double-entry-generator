package cmb

import (
	"encoding/csv"
	"io"
	"time"

	"fmt"
	"log"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

type Cmb struct {
	Statistics       Statistics    `json:"statistics,omitempty"`
	LineNum          int           `json:"line_num,omitempty"`
	Mode             CardMode      `json:"mode,omitempty"`
	DebitOrders      []DebitOrder  `json:"debit_orders,omitempty"`
	CreditOrders     []CreditOrder `json:"credit_orders,omitempty"`
	DebitRealHeaders []string      `json:"debit_real_headers,omitempty"`
	CreditBillYear   string        `json:"credit_bill_year,omitempty"`
	CreditBillMonth  string        `json:"credit_bill_month,omitempty"`
}

func New() *Cmb {
	return &Cmb{
		Statistics:       Statistics{},
		LineNum:          0,
		Mode:             CardModeUnknown,
		DebitOrders:      make([]DebitOrder, 0),
		CreditOrders:     make([]CreditOrder, 0),
		DebitRealHeaders: make([]string, 0),
		CreditBillYear:   fmt.Sprintf("%04d", int(time.Now().Year())), // 默认使用当前年月兜底
		CreditBillMonth:  fmt.Sprintf("%02d", int(time.Now().Month())),
	}
}

// Translate the cmb bill records to IR.
func (cmb *Cmb) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-cmb] ")

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
		}
		if err != nil {
			return nil, err
		}

		cmb.LineNum++

		// update card mode
		if cmb.Mode == CardModeUnknown {
			cmb.Mode = updateCardMode(row)
		}

		// update credit bill year&month
		year, month := extractYearAndMonthFromCreditTitle(safeAccessStrList(row, 0))
		if year != "" && month != "" {
			cmb.CreditBillYear = year
			cmb.CreditBillMonth = month
		}

		// update debit headers
		if safeAccessStrList(row, 0) == allDebitHeaders[0] {
			cmb.DebitRealHeaders = row
		}

		if isValidDebitDateFormat(safeAccessStrList(row, 0)) {
			err = cmb.translateDebitToOrders(fillDebitRow(cmb.DebitRealHeaders, row))
			if err != nil {
				return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
					cmb.LineNum, err)
			}
		} else if isValidCreditCardNoFormat(safeAccessStrList(row, 4)) {
			err = cmb.translateCreditToOrders(row)
			if err != nil {
				return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
					cmb.LineNum, err)
			}
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	if cmb.Mode == CardModeDebit {
		return cmb.convertDebitToIR(), nil
	}
	if cmb.Mode == CardModeCredit {
		return cmb.convertCreditToIR(), nil
	}
	return nil, fmt.Errorf("cmb card mode is unknown")
}
