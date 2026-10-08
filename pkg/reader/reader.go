// Package reader turns bill sources of any format into one shape: a Table of
// string cells. Readers know structure only. They never interpret a cell as
// an amount, a date, or a transaction; that is the job of later layers.
package reader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Table is the only thing a Reader produces.
//
// Grid formats (csv, xlsx, xls) leave Headers nil: the header row, if any, is
// somewhere in Rows and the caller locates it. Named formats (json, xml,
// text) fill Headers with the declared column names and every row has one
// cell per header.
type Table struct {
	Headers []string
	Rows    [][]string
	Source  string
}

// Config is the `reader:` block of a template. Only Format is required.
type Config struct {
	// Format selects the reader: csv, xlsx, xls, json, xml, text.
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// Encoding of text input: utf-8 (default), gbk, gb18030, utf-16le, utf-16be.
	Encoding string `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	// Delimiter for csv: "," (default), "\t", ";" or any single rune.
	Delimiter string `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
	// StripTabs removes every tab before csv parsing. Some exports pad cells
	// with tabs to stop spreadsheets from reformatting them.
	StripTabs bool `json:"stripTabs,omitempty" yaml:"stripTabs,omitempty"`
	// KeepBlankLines makes csv emit a nil row where the file has a blank
	// line, so row indexes match spreadsheet formats.
	KeepBlankLines bool `json:"keepBlankLines,omitempty" yaml:"keepBlankLines,omitempty"`

	// Sheet picks the worksheet for xlsx/xls by name or 0-based index.
	// Empty means the first sheet.
	Sheet string `json:"sheet,omitempty" yaml:"sheet,omitempty"`

	// Records is an XPath selecting one node per row (json and xml).
	Records string `json:"records,omitempty" yaml:"records,omitempty"`
	// Columns maps column name to an XPath relative to each record node
	// (json and xml). Declaration order is the column order.
	Columns OrderedMap `json:"columns,omitempty" yaml:"columns,omitempty"`

	// Record is a regexp with named groups matched against each line (text).
	// Group names become columns in the order they appear.
	Record string `json:"record,omitempty" yaml:"record,omitempty"`
	// Continuation is a regexp for lines that belong to the previous record
	// (text). Its named groups are appended to the same-named columns.
	Continuation string `json:"continuation,omitempty" yaml:"continuation,omitempty"`
	// Convert names a pre-processor that produces text from a binary file.
	// Currently "pdftotext-layout".
	Convert string `json:"convert,omitempty" yaml:"convert,omitempty"`
}

// Normalize fills defaults and canonicalizes Format.
func (c Config) Normalize() Config {
	c.Format = NormalizeFormat(c.Format)
	return c
}

// NormalizeFormat maps aliases to the canonical format name. Unknown values
// are returned lowercased so the dispatcher can report them.
func NormalizeFormat(format string) string {
	switch f := strings.ToLower(strings.TrimSpace(format)); f {
	case "", "txt", "csv", "tsv":
		return "csv"
	case "xlsx", "xls", "json", "xml":
		return f
	case "text", "pdf":
		return "text"
	default:
		return f
	}
}

// FormatForFile guesses the format of a bill file from its extension.
// Returns "" when the extension says nothing useful.
func FormatForFile(filename string) string {
	switch ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), ".")); ext {
	case "csv", "tsv", "txt":
		return "csv"
	case "xlsx", "xls", "json", "xml":
		return ext
	case "pdf":
		return "text"
	default:
		return ""
	}
}

// Compatible reports whether a bill file can feed a reader of this format.
// A text reader with a converter accepts the converter's input (pdf).
func (c Config) Compatible(filename string) bool {
	c = c.Normalize()
	billFmt := FormatForFile(filename)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch c.Format {
	case "text":
		if ext == "pdf" {
			return c.Convert != ""
		}
		return ext == "txt" || ext == "text"
	case "csv":
		return billFmt == "csv"
	default:
		return billFmt == c.Format
	}
}

// ReadFile reads a bill file with the configured reader.
func ReadFile(filename string, cfg Config) (Table, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Table{}, err
	}
	return ReadBytes(filename, data, cfg)
}

// ReadBytes reads bill content already in memory. name is used for error
// messages, format sniffing and converters.
func ReadBytes(name string, data []byte, cfg Config) (Table, error) {
	cfg = cfg.Normalize()
	var (
		table Table
		err   error
	)
	switch cfg.Format {
	case "csv":
		table, err = readCSV(data, cfg)
	case "xlsx":
		table, err = readXLSX(data, cfg)
	case "xls":
		table, err = readXLS(data, cfg)
	case "json":
		table, err = readJSON(data, cfg)
	case "xml":
		table, err = readXML(data, cfg)
	case "text":
		table, err = readText(name, data, cfg)
	default:
		return Table{}, fmt.Errorf("unsupported reader format %q", cfg.Format)
	}
	if err != nil {
		return Table{}, err
	}
	table.Source = name
	return table, nil
}

func hasOLEHeader(data []byte) bool {
	return len(data) >= 8 && bytes.Equal(data[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
}
