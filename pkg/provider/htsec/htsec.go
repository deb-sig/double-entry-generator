package htsec

import (
	"fmt"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/xuri/excelize/v2"
)

type Htsec struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

func New() *Htsec {
	return &Htsec{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
	}
}

func (h *Htsec) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Htsec] ")

	xlsxFile, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	rows, err := xlsxFile.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		h.LineNum++
		if h.LineNum <= 1 {
			// The first line is xlsx file header.
			continue
		}

		err = h.translateToOrders(row)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", h.LineNum, err)
		}
	}

	// 当有成交数量与成交价格但无成交金额时，表示新股或新债中签，需处理数据，删除交割单中有成交金额而无成交数量与成交价格的记录
	for index, o := range h.Orders {

		if o.Price != 0 && o.Volume != 0 && o.TransactionAmount == 0 {
			for ti, tar := range h.Orders {
				if o.TxTypeOriginal == tar.TxTypeOriginal && tar.TransactionAmount != 0 && tar.Price == 0 && tar.Volume == 0 {
					h.Orders[index].TransactionAmount = tar.TransactionAmount
					h.Orders[index].OccurAmount = tar.OccurAmount

					h.Orders[ti].Useless = true
				}
			}
		}
	}

	// 移除已被合并的新股或新债有成交金额的条目
	for i := 0; i < len(h.Orders); i++ {
		if h.Orders[i].Useless {
			h.Orders = append(h.Orders[:i], h.Orders[i+1:]...)
			i--
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return h.convertToIR(), nil
}

// TranslateFromExcelBytes 从字节数组解析 XLSX 文件（用于 WASM）
func (h *Htsec) TranslateFromExcelBytes(fileData []byte) (*ir.IR, error) {
	log.SetPrefix("[Provider-Htsec] ")
	log.Printf("TranslateFromExcelBytes called with %d bytes", len(fileData))
	
	// 使用 excelize.OpenReader 从字节流读取
	f, err := excelize.OpenReader(strings.NewReader(string(fileData)))
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件。原始错误: %v", err)
	}
	
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("无法获取Excel的第一个工作表。原始错误: %v", err)
	}
	
	for _, row := range rows {
		h.LineNum++
		if h.LineNum == 1 {
			// 第一行是表头，跳过
			continue
		}
		
		if err := h.translateToOrders(row); err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", h.LineNum, err)
		}
	}
	
	log.Printf("Finished to parse the Excel file from bytes")
	return h.convertToIR(), nil
}
