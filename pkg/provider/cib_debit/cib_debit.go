package cib_debit

import (
	"fmt"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
)

// CibDebit is the provider for Industrial Bank debit card statements.
type CibDebit struct {
	Statistics  Statistics `json:"statistics,omitempty"`
	LineNum     int        `json:"line_num,omitempty"`
	Orders      []Order    `json:"orders,omitempty"`
	AccountName string     `json:"account_name,omitempty"`
	AccountNum  string     `json:"account_num,omitempty"`
	SubAccount  string     `json:"sub_account,omitempty"`
	Currency    string     `json:"currency,omitempty"`
}

// New creates a new CibDebit provider.
func New() *CibDebit {
	return &CibDebit{
		Statistics: Statistics{},
		Orders:     make([]Order, 0),
		Currency:   defaultCurrency,
	}
}

// Translate converts a single CIB XLS statement to IR.
func (c *CibDebit) Translate(filename string) (*ir.IR, error) {
	return c.TranslateFiles([]string{filename})
}

// TranslateFiles converts multiple CIB XLS statements into one IR.
func (c *CibDebit) TranslateFiles(filenames []string) (*ir.IR, error) {
	log.SetPrefix("[Provider-CIB_debit] ")
	for _, filename := range filenames {
		switch {
		case strings.HasSuffix(strings.ToLower(filename), ".xls"):
			if err := c.parseExcel(filename); err != nil {
				return nil, err
			}
		case strings.HasSuffix(strings.ToLower(filename), ".xlsx"):
			if err := c.parseXLSX(filename); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported file format %s, only .xls and .xlsx files are supported", filename)
		}
	}
	log.Printf("Finished to parse %d CIB debit file(s), parsed %d orders", len(filenames), len(c.Orders))
	return c.convertToIR()
}

func (c *CibDebit) parseExcel(filename string) error {
	xlFile, err := xls.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("无法打开Excel文件 %s。原始错误: %w", filename, err)
	}

	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return fmt.Errorf("无法获取Excel的第一个工作表 %s。原始错误: %w", filename, err)
	}

	accountName := ""
	accountNum := ""
	subAccount := ""
	currency := defaultCurrency
	isDataSection := false

	for rowIdx := 0; rowIdx <= int(sheet.GetNumberRows()); rowIdx++ {
		row, err := sheet.GetRow(rowIdx)
		if err != nil || row == nil {
			continue
		}

		rowData := make([]string, 0, len(row.GetCols()))
		for _, col := range row.GetCols() {
			rowData = append(rowData, strings.TrimSpace(col.GetString()))
		}
		if len(rowData) == 0 || strings.TrimSpace(strings.Join(rowData, "")) == "" {
			continue
		}

		c.LineNum = rowIdx + 1
		first := strings.TrimSpace(rowData[0])
		rowText := strings.Join(rowData, "")

		if first == "账户户名" && len(rowData) > 1 {
			accountName = rowData[1]
			continue
		}
		if first == "账户账号" && len(rowData) > 1 {
			accountNum = rowData[1]
			continue
		}
		if first == "卡内账户" && len(rowData) > 1 {
			subAccount = rowData[1]
			currency = parseCurrency(subAccount)
			continue
		}
		if strings.Contains(rowText, "说明") {
			break
		}
		if !isDataSection {
			if first == "交易时间" {
				isDataSection = true
			}
			continue
		}

		c.AccountName = accountName
		c.AccountNum = accountNum
		c.SubAccount = subAccount
		c.Currency = currency
		if err := c.translateRow(rowData); err != nil {
			return fmt.Errorf("failed to translate bill %s: line %d: %w", filename, c.LineNum, err)
		}
	}

	return nil
}

func (c *CibDebit) parseXLSX(filename string) error {
	xlsxFile, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("无法打开Excel文件 %s。原始错误: %w", filename, err)
	}
	defer func() {
		if err := xlsxFile.Close(); err != nil {
			log.Printf("close xlsx file error: %v", err)
		}
	}()

	sheets := xlsxFile.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("Excel文件 %s 没有工作表", filename)
	}
	rows, err := xlsxFile.GetRows(sheets[0])
	if err != nil {
		return fmt.Errorf("无法读取Excel文件 %s 的第一个工作表。原始错误: %w", filename, err)
	}

	return c.parseRows(filename, rows)
}

func (c *CibDebit) parseRows(filename string, rows [][]string) error {
	accountName := ""
	accountNum := ""
	subAccount := ""
	currency := defaultCurrency
	isDataSection := false

	for _, rowData := range rows {
		for idx := range rowData {
			rowData[idx] = strings.TrimSpace(rowData[idx])
		}
		if len(rowData) == 0 || strings.TrimSpace(strings.Join(rowData, "")) == "" {
			continue
		}

		c.LineNum++
		first := strings.TrimPrefix(strings.TrimSpace(rowData[0]), "\ufeff")
		rowText := strings.Join(rowData, "")

		if first == "账户户名" && len(rowData) > 1 {
			accountName = rowData[1]
			continue
		}
		if first == "账户账号" && len(rowData) > 1 {
			accountNum = rowData[1]
			continue
		}
		if first == "卡内账户" && len(rowData) > 1 {
			subAccount = rowData[1]
			currency = parseCurrency(subAccount)
			continue
		}
		if strings.Contains(rowText, "说明") {
			break
		}
		if !isDataSection {
			if first == "交易时间" {
				isDataSection = true
			}
			continue
		}

		c.AccountName = accountName
		c.AccountNum = accountNum
		c.SubAccount = subAccount
		c.Currency = currency
		if err := c.translateRow(rowData); err != nil {
			return fmt.Errorf("failed to translate bill %s: line %d: %w", filename, c.LineNum, err)
		}
	}

	return nil
}
