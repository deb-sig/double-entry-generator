package citic

import (
	"fmt"
	"log"
	"time"

	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/extrame/xls"
)

type Citic struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

func New() *Citic {
	return &Citic{
		Statistics: Statistics{},
		LineNum:    2, // 表格前2行是标题
		Orders:     make([]Order, 0),
	}
}

// Translate the citic bill records to IR.
func (citic *Citic) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Citic] ")

	xlsFile, err := xls.Open(filename, "utf-8")

	if err != nil {
		return nil, err
	}

	sheet := xlsFile.GetSheet(0)

	for citic.LineNum = 2; citic.LineNum <= int(sheet.MaxRow); citic.LineNum++ {
		var row []string
		// 一行有8列
		for i := 0; i < 8; i++ {
			row = append(row, sheet.Row(citic.LineNum).Col(i))
		}

		// 跳过可能的空行
		if row[0] == "" {
			continue
		}

		err = citic.TranslateToOrders(row)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", citic.LineNum, err)
		}
	}

	// hack:
	// 中信账单只有日期没有时间，且顺序是倒序。
	// 补上ns时差，以便排序后为准确的正序
	for index := range citic.Orders {
		hackDuration := time.Duration(len(citic.Orders)-index) * time.Nanosecond
		citic.Orders[index].TradeTime = citic.Orders[index].TradeTime.Add(hackDuration)
		citic.Orders[index].PostTime = citic.Orders[index].PostTime.Add(hackDuration)
	}

	log.Printf("Finished to parse the file %s", filename)

	return citic.convertToIR(), nil
}
