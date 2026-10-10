package reader

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/reader/internal/xls"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func decodeText(data []byte, enc string) ([]byte, error) {
	var dec *encoding.Decoder
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "", "utf-8", "utf8":
		return data, nil
	case "gbk", "gb2312", "gb18030":
		dec = simplifiedchinese.GB18030.NewDecoder()
	case "utf-16le", "utf16le":
		dec = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	case "utf-16be", "utf16be":
		dec = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case "utf-16", "utf16":
		dec = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
	default:
		return nil, fmt.Errorf("unsupported encoding %q", enc)
	}
	out, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), dec))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", enc, err)
	}
	return out, nil
}

// Delimiter turns the template spelling of a csv delimiter into a rune.
func Delimiter(value string) rune {
	if value == "\t" {
		return '\t'
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "comma", ",":
		return ','
	case "\\t", "tab", "tsv":
		return '\t'
	case "semicolon", ";":
		return ';'
	default:
		return []rune(value)[0]
	}
}

func readCSV(data []byte, cfg Config) (Table, error) {
	data, err := decodeText(data, cfg.Encoding)
	if err != nil {
		return Table{}, err
	}
	if cfg.StripTabs {
		data = bytes.ReplaceAll(data, []byte{'\t'}, nil)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.Comma = Delimiter(cfg.Delimiter)

	if !cfg.KeepBlankLines {
		rows, err := reader.ReadAll()
		if err != nil {
			return Table{}, err
		}
		return Table{Rows: rows}, nil
	}

	// encoding/csv skips blank lines. Restore their positions so header
	// windows behave the same for csv and spreadsheets. FieldPos identifies
	// logical records, including quoted multiline cells.
	var rows [][]string
	nextLine := 1
	var previousOffset int64
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Table{}, err
		}
		startLine, _ := reader.FieldPos(0)
		for line := nextLine; line < startLine; line++ {
			rows = append(rows, nil)
		}
		rows = append(rows, record)
		offset := reader.InputOffset()
		nextLine += bytes.Count(data[previousOffset:offset], []byte{'\n'})
		previousOffset = offset
	}
	return Table{Rows: rows}, nil
}

func readXLSX(data []byte, cfg Config) (Table, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return Table{}, err
	}
	defer func() { _ = f.Close() }()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return Table{}, fmt.Errorf("xlsx has no sheets")
	}
	name, err := pickSheet(sheets, cfg.Sheet)
	if err != nil {
		return Table{}, err
	}
	rows, err := f.GetRows(name)
	if err != nil {
		return Table{}, err
	}
	raw, err := f.GetRows(name, excelize.Options{RawCellValue: true})
	if err != nil {
		return Table{}, err
	}
	return Table{Rows: plainNumbers(rows, raw)}, nil
}

// plainNumbers undoes the zero padding a cell's number format adds.
// excelize applies custom formats such as "0.000", so a stored 0.85 reads
// as "0.850", and an amount would gain a decimal place. When the shown
// text is a plain decimal of the same value as the stored one, its trailing
// fractional zeros are dropped. Dates, percentages, thousands separators
// and text cells keep their shown text.
func plainNumbers(shown, raw [][]string) [][]string {
	for r, row := range shown {
		if r >= len(raw) {
			break
		}
		for c, cell := range row {
			if c >= len(raw[r]) || cell == raw[r][c] || !isPlainDecimal(cell) {
				continue
			}
			a, errA := strconv.ParseFloat(cell, 64)
			b, errB := strconv.ParseFloat(raw[r][c], 64)
			if errA != nil || errB != nil || a != b {
				continue
			}
			row[c] = trimFractionZeros(cell)
		}
	}
	return shown
}

func isPlainDecimal(s string) bool {
	s = strings.TrimPrefix(s, "-")
	whole, frac, dotted := strings.Cut(s, ".")
	digits := func(t string) bool {
		for _, ch := range t {
			if ch < '0' || ch > '9' {
				return false
			}
		}
		return true
	}
	return whole != "" && digits(whole) && (!dotted || (frac != "" && digits(frac)))
}

func trimFractionZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// XLSRows returns the cells of the first sheet of an .xls workbook, one
// slice per row (nil for rows the file does not define). Legacy providers
// use it so the whole tool reads .xls through one permissively licensed
// reader.
func XLSRows(data []byte) ([][]string, error) {
	t, err := readXLS(data, Config{})
	return t.Rows, err
}

func readXLS(data []byte, cfg Config) (table Table, err error) {
	// Many "xls" exports are really csv or html with a spreadsheet extension.
	if !hasOLEHeader(data) {
		return readCSV(data, cfg)
	}
	// internal/xls is a patched copy of extrame/xls (Apache-2.0); the reader
	// DEG used before was GPL-3.0, which rules out embedding the engine in
	// closed apps. It can panic on malformed files, so a panic is reported
	// as an ordinary read error.
	defer func() {
		if r := recover(); r != nil {
			table, err = Table{}, fmt.Errorf("xls: unreadable workbook: %v", r)
		}
	}()
	wb, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil {
		return readCSV(data, cfg)
	}
	index := 0
	if cfg.Sheet != "" {
		names := make([]string, wb.NumSheets())
		for i := range names {
			if s := wb.GetSheet(i); s != nil {
				names[i] = s.Name
			}
		}
		name, err := pickSheet(names, cfg.Sheet)
		if err != nil {
			return Table{}, err
		}
		for i, n := range names {
			if n == name {
				index = i
			}
		}
	}
	sheet := wb.GetSheet(index)
	if sheet == nil {
		return Table{}, fmt.Errorf("xls has no sheet %d", index)
	}
	rows := make([][]string, 0, int(sheet.MaxRow)+1)
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
	return Table{Rows: rows}, nil
}

// pickSheet resolves a sheet selector (name or 0-based index) against the
// workbook's sheet names.
func pickSheet(names []string, selector string) (string, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return names[0], nil
	}
	for _, n := range names {
		if n == selector {
			return n, nil
		}
	}
	if i, err := strconv.Atoi(selector); err == nil {
		if i < 0 || i >= len(names) {
			return "", fmt.Errorf("sheet index %d out of range (workbook has %d sheets)", i, len(names))
		}
		return names[i], nil
	}
	return "", fmt.Errorf("sheet %q not found; available: %s", selector, strings.Join(names, ", "))
}
