package wechat

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/xuri/excelize/v2"
)

// Wechat is the provider for Wechat.
type Wechat struct {
	Statistics           Statistics `json:"statistics,omitempty"`
	LineNum              int        `json:"line_num,omitempty"`
	Orders               []Order    `json:"orders,omitempty"`
	IgnoreInvalidTxTypes bool       `json:"ignore_invalid_tx_types,omitempty"`
}

// New creates a new wechat provider.
func New() *Wechat {
	return &Wechat{
		Statistics:           Statistics{},
		LineNum:              0,
		Orders:               make([]Order, 0),
		IgnoreInvalidTxTypes: false,
	}
}

// Translate translates the wechat bill records to IR.
func (w *Wechat) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Wechat] ")

	// Check if it's an Excel file
	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return w.translateExcel(filename)
	}

	// Handle CSV file
	return w.translateCSV(filename)
}

// translateCSV handles CSV file parsing
func (w *Wechat) translateCSV(filename string) (*ir.IR, error) {
	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	// Read the entire file content and replace tabs
	content, err := io.ReadAll(billReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}
	cleanedContent := strings.ReplaceAll(string(content), "\t", "")

	// Create a new CSV reader with the cleaned content
	csvReader := csv.NewReader(strings.NewReader(cleanedContent))
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

		w.LineNum++
		if w.LineNum <= 17 {
			// The first 17 lines are useless for us.
			continue
		}

		err = w.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				w.LineNum, err)
		}
	}
	log.Printf("Finished to parse the CSV file %s", filename)
	return w.convertToIR(), nil
}

// translateExcel handles Excel file parsing
func (w *Wechat) translateExcel(filename string) (*ir.IR, error) {
	log.Printf("Attempting to open Excel file: %s", filename)

	xlsxFile, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件，请检查文件路径或文件是否已损坏。原始错误: %v", err)
	}

	rows, err := xlsxFile.GetRows("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}

	for _, row := range rows {
		w.LineNum++
		if w.LineNum <= 17 {
			// The first 17 lines are useless for us.
			continue
		}

		err = w.translateToOrders(row)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				w.LineNum, err)
		}
	}

	log.Printf("Finished to parse the Excel file %s", filename)
	return w.convertToIR(), nil
}
