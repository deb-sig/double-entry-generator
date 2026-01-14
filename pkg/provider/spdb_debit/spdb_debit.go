package spdb_debit

import (
	"fmt"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/shakinm/xlsReader/xls"
)

// SpdbDebit is the provider for Shanghai Pudong Development Bank debit card statements.
type SpdbDebit struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	CardName   string     `json:"card_name,omitempty"`
	AccountNum string     `json:"account_num,omitempty"`
	Currency   string     `json:"currency,omitempty"`
}

// New creates a new SpdbDebit provider.
func New() *SpdbDebit {
	return &SpdbDebit{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		CardName:   "",
		AccountNum: "",
		Currency:   defaultCurrency,
	}
}

// SetCurrency sets the currency for the SpdbDebit provider.
func (sd *SpdbDebit) SetCurrency(currency string) {
	if currency != "" {
		sd.Currency = currency
	}
}

// Translate converts the XLS statement to IR.
func (sd *SpdbDebit) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-SPDB_debit] ")

	// Check if it's an Excel file
	if strings.HasSuffix(strings.ToLower(filename), ".xls") {
		return sd.translateExcel(filename)
	}

	return nil, fmt.Errorf("unsupported file format, only .xls files are supported")
}

// translateExcel handles Excel file parsing.
func (sd *SpdbDebit) translateExcel(filename string) (*ir.IR, error) {
	log.Printf("Attempting to open Excel file with xlsReader: %s", filename)

	xlFile, err := xls.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件，请检查文件路径或文件是否已损坏。原始错误: %v", err)
	}

	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}

	// Skip rows until we find the actual transaction data
	// 浦发银行借记卡交易明细的结构可能包含：
	// - 标题行
	// - 账户信息行
	// - 空行
	// - 表头行
	// - 交易记录行
	// - 合计行
	isDataSection := false

	for i := 0; i <= int(sheet.GetNumberRows()); i++ {
		row, err := sheet.GetRow(i)
		if err != nil {
			log.Printf("跳过无法读取的行 %d: %v", i, err)
			continue
		}
		if row == nil {
			continue
		}

		var rowData []string
		for _, col := range row.GetCols() {
			rowData = append(rowData, col.GetString())
		}

		sd.LineNum = i + 1

		// Skip empty rows
		if len(rowData) == 0 || (len(rowData) == 1 && strings.TrimSpace(rowData[0]) == "") {
			continue
		}

		// Join all columns for easier checking
		rowStr := strings.Join(rowData, "")

		// Check if we've reached the end of data
		if strings.Contains(rowStr, "合计") {
			break
		}

		// Check if this is the header row (start of data section)
		if !isDataSection && len(rowData) > 3 {
			// Look for header keywords in any column
			for _, col := range rowData {
				if strings.Contains(col, "交易日期") || strings.Contains(col, "交易时间") || 
				   strings.Contains(col, "交易摘要") || strings.Contains(col, "交易金额") {
					isDataSection = true
					log.Printf("Found header row at line %d, starting data parsing", sd.LineNum)
					break
				}
			}
			// Skip the header row itself
			continue
		}

		// Only process rows after the header
		if isDataSection {
			// 处理数据行，translateToOrders 函数内部会进行验证
			err = sd.translateToOrders(rowData)
			if err != nil {
				return nil, fmt.Errorf("failed to translate bill: line %d: %v",
					sd.LineNum, err)
			}
		}
	}

	log.Printf("Finished to parse the excel file %s, parsed %d orders", filename, len(sd.Orders))
	return sd.convertToIR()
}

// TranslateFromExcelBytes parses XLS file from byte array (for WASM).
func (sd *SpdbDebit) TranslateFromExcelBytes(fileData []byte) (*ir.IR, error) {
	log.SetPrefix("[Provider-SPDB_debit] ")
	log.Printf("TranslateFromExcelBytes called with %d bytes", len(fileData))
	
	// Use xls.OpenReader to read from byte stream
	xlFile, err := xls.OpenReader(strings.NewReader(string(fileData)))
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件。原始错误: %v", err)
	}
	
	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}

	// Skip rows until we find the actual transaction data
	isDataSection := false
	
	for i := 0; i <= int(sheet.GetNumberRows()); i++ {
		row, err := sheet.GetRow(i)
		if err != nil {
			log.Printf("跳过无法读取的行 %d: %v", i, err)
			continue
		}
		
		if row == nil {
			continue
		}
		
		var rowData []string
		for _, col := range row.GetCols() {
			rowData = append(rowData, col.GetString())
		}
		
		sd.LineNum = i + 1
		
		// Skip empty rows
		if len(rowData) == 0 || (len(rowData) == 1 && strings.TrimSpace(rowData[0]) == "") {
			continue
		}

		// Join all columns for easier checking
		rowStr := strings.Join(rowData, "")

		// Check if we've reached the end of data
		if strings.Contains(rowStr, "合计") {
			break
		}

		// Check if this is the header row (start of data section)
		if !isDataSection && len(rowData) > 3 {
			// Look for header keywords in any column
			for _, col := range rowData {
				if strings.Contains(col, "交易日期") || strings.Contains(col, "交易时间") || 
				   strings.Contains(col, "交易摘要") || strings.Contains(col, "交易金额") {
					isDataSection = true
					log.Printf("Found header row at line %d, starting data parsing", sd.LineNum)
					break
				}
			}
			// Skip the header row itself
			continue
		}

		// Only process rows after the header
		if isDataSection {
			// 处理数据行，translateToOrders 函数内部会进行验证
			err = sd.translateToOrders(rowData)
			if err != nil {
				return nil, fmt.Errorf("failed to translate bill: line %d: %v",
					sd.LineNum, err)
			}
		}
	}
	
	log.Printf("Finished to parse the Excel file from bytes, parsed %d orders", len(sd.Orders))
	return sd.convertToIR()
}
