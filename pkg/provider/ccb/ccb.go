package ccb

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/shakinm/xlsReader/xls"
)

// CCB is the provider for China Construction Bank.
type CCB struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
	CardName   string     `json:"card_name,omitempty"`
	AccountNum string     `json:"account_num,omitempty"`
	Currency   string     `json:"currency,omitempty"`
}

// New creates a new CCB provider.
func New() *CCB {
	return &CCB{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
		CardName:   "",
		AccountNum: "",
		Currency:   "CNY", // 默认使用CNY而不是人民币
	}
}

// SetCurrency sets the currency for the CCB provider
func (ccb *CCB) SetCurrency(currency string) {
	if currency != "" {
		ccb.Currency = currency
	}
}

// Translate translates the CCB bill records to IR.
func (ccb *CCB) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-CCB] ")

	// Check if it's an Excel file
	if strings.HasSuffix(strings.ToLower(filename), ".xls") || strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return ccb.translateExcel(filename)
	}

	// Handle CSV file
	return ccb.translateCSV(filename)
}

// translateCSV handles CSV file parsing
func (ccb *CCB) translateCSV(filename string) (*ir.IR, error) {
	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	for {
		line, err := csvReader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		ccb.LineNum++

		// Skip empty lines
		if len(line) == 0 || (len(line) == 1 && strings.TrimSpace(line[0]) == "") {
			continue
		}

		// Parse header information - 适配新的文件格式
		if ccb.LineNum == 1 && len(line) > 0 && strings.Contains(line[0], "China Construction Bank") {
			continue
		} else if ccb.LineNum == 2 && len(line) > 1 && strings.Contains(line[0], "开户机构") {
			continue
		} else if ccb.LineNum == 3 && len(line) > 1 && strings.Contains(line[0], "币") {
			// 不从文件中读取货币信息，保持使用配置中的默认货币
			continue
		} else if ccb.LineNum == 4 && len(line) > 1 && strings.Contains(line[0], "账") {
			if len(line) > 1 {
				ccb.AccountNum = strings.TrimSpace(line[1])
			}
			continue
		} else if ccb.LineNum <= 6 {
			// Skip empty lines and header
			continue
		}

		// Check if it's the end of data - 适配新的结束标记
		if len(line) > 0 && (strings.Contains(line[0], "以上数据仅供参考") ||
			strings.Contains(line[0], "具体内容请以柜台为准") ||
			strings.Contains(strings.Join(line, ""), "以上数据仅供参考")) {
			break
		}

		// Check if it's the header row
		if len(line) > 0 && strings.Contains(line[0], "记账日") {
			continue
		}

		err = ccb.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %v",
				ccb.LineNum, err)
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return ccb.convertToIR(), nil
}

// translateExcel handles Excel file parsing
func (ccb *CCB) translateExcel(filename string) (*ir.IR, error) {
	log.Printf("Attempting to open Excel file with xlsReader: %s", filename)

	xlFile, err := xls.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件，请检查文件路径或文件是否已损坏。原始错误: %v", err)
	}

	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}

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

		ccb.LineNum = i + 1

		// Skip empty rows
		if len(rowData) == 0 || (len(rowData) == 1 && strings.TrimSpace(rowData[0]) == "") {
			continue
		}

		// Parse header information - 适配新的文件格式
		if ccb.LineNum == 1 && len(rowData) > 0 && strings.Contains(rowData[0], "China Construction Bank") {
			continue
		} else if ccb.LineNum == 2 && len(rowData) > 1 && strings.Contains(rowData[0], "开户机构") {
			continue
		} else if ccb.LineNum == 3 && len(rowData) > 1 && strings.Contains(rowData[0], "币") {
			// 不从文件中读取货币信息，保持使用配置中的默认货币
			continue
		} else if ccb.LineNum == 4 && len(rowData) > 1 && strings.Contains(rowData[0], "账") {
			if len(rowData) > 1 {
				ccb.AccountNum = strings.TrimSpace(rowData[1])
			}
			continue
		} else if ccb.LineNum <= 6 {
			// Skip empty lines and header
			continue
		}

		// Check if it's the end of data - 适配新的结束标记
		if len(rowData) > 0 && (strings.Contains(rowData[0], "以上数据仅供参考") ||
			strings.Contains(rowData[0], "具体内容请以柜台为准") ||
			strings.Contains(strings.Join(rowData, ""), "以上数据仅供参考")) {
			break
		}

		// Check if it's the header row
		if len(rowData) > 0 && strings.Contains(rowData[0], "记账日") {
			continue
		}

		err = ccb.translateToOrders(rowData)
		if err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %v",
				ccb.LineNum, err)
		}
	}

	log.Printf("Finished to parse the excel file %s", filename)
	return ccb.convertToIR(), nil
}

// TranslateFromExcelBytes 从字节数组解析 XLS 文件（用于 WASM）
func (ccb *CCB) TranslateFromExcelBytes(fileData []byte) (*ir.IR, error) {
	log.SetPrefix("[Provider-CCB] ")
	log.Printf("TranslateFromExcelBytes called with %d bytes", len(fileData))
	
	// 使用 xls.OpenReader 从字节流读取
	xlFile, err := xls.OpenReader(strings.NewReader(string(fileData)))
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件。原始错误: %v", err)
	}
	
	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}
	
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
		
		ccb.LineNum = i + 1
		
		// Skip empty rows
		if len(rowData) == 0 || (len(rowData) == 1 && strings.TrimSpace(rowData[0]) == "") {
			continue
		}
		
		// Parse header information - 适配新的文件格式
		if ccb.LineNum == 1 && len(rowData) > 0 && strings.Contains(rowData[0], "China Construction Bank") {
			continue
		} else if ccb.LineNum == 2 && len(rowData) > 1 && strings.Contains(rowData[0], "开户机构") {
			continue
		} else if ccb.LineNum == 3 && len(rowData) > 1 && strings.Contains(rowData[0], "币") {
			// 不从文件中读取货币信息，保持使用配置中的默认货币
			continue
		} else if ccb.LineNum == 4 && len(rowData) > 1 && strings.Contains(rowData[0], "账") {
			if len(rowData) > 1 {
				ccb.AccountNum = strings.TrimSpace(rowData[1])
			}
			continue
		} else if ccb.LineNum <= 6 {
			// Skip empty lines and header
			continue
		}
		
		// Check if it's the end of data
		if len(rowData) > 0 && (strings.Contains(rowData[0], "以上数据仅供参考") ||
			strings.Contains(rowData[0], "具体内容请以柜台为准") ||
			strings.Contains(strings.Join(rowData, ""), "以上数据仅供参考")) {
			break
		}
		
		// Check if it's the header row
		if len(rowData) > 0 && strings.Contains(rowData[0], "记账日") {
			continue
		}
		
		err = ccb.translateToOrders(rowData)
		if err != nil {
			return nil, fmt.Errorf("failed to translate bill: line %d: %v",
				ccb.LineNum, err)
		}
	}
	
	log.Printf("Finished to parse the Excel file from bytes")
	return ccb.convertToIR(), nil
}