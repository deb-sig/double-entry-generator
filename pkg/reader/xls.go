// Package reader reads bill files into rows of cells.
package reader

import (
	"bytes"
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/reader/internal/xls"
)

// XLSRows returns the cells of the first sheet of an .xls workbook, one
// slice per row (nil for rows the file does not define).
//
// internal/xls is a patched copy of extrame/xls (Apache-2.0). Providers used
// shakinm/xlsReader before, which is GPL-3.0 and binds every binary that
// links it. The reader can panic on malformed files, so a panic is
// reported as an ordinary read error.
func XLSRows(data []byte) (rows [][]string, err error) {
	if !hasOLEHeader(data) {
		return nil, fmt.Errorf("xls: not an .xls workbook")
	}
	defer func() {
		if r := recover(); r != nil {
			rows, err = nil, fmt.Errorf("xls: unreadable workbook: %v", r)
		}
	}()
	wb, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil {
		return nil, err
	}
	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("xls has no sheet 0")
	}
	rows = make([][]string, 0, int(sheet.MaxRow)+1)
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			rows = append(rows, nil)
			continue
		}
		record := make([]string, 0, row.DefinedCols())
		for c := 0; c < row.DefinedCols(); c++ {
			record = append(record, row.Col(c))
		}
		rows = append(rows, record)
	}
	return rows, nil
}

func hasOLEHeader(data []byte) bool {
	return len(data) >= 8 && bytes.Equal(data[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
}
