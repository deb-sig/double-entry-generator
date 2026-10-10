package reader

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
)

// HTML bills (bank e-mails, exported statement pages) use the same
// records/columns contract as json and xml, with XPath over the DOM:
//
//	records: "//table[@id='txns']//tr[position()>1]"
//	columns: { date: "td[1]", narration: "td[2]", amount: "td[3]" }
func readHTML(data []byte, cfg Config) (Table, error) {
	data, err := decodeText(data, cfg.Encoding)
	if err != nil {
		return Table{}, err
	}
	recordsExpr, colExprs, err := compileTree(cfg)
	if err != nil {
		return Table{}, err
	}
	doc, err := htmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		return Table{}, fmt.Errorf("parse html: %w", err)
	}
	table := Table{Headers: append([]string{}, cfg.Columns.Keys...)}
	for _, node := range htmlquery.QuerySelectorAll(doc, recordsExpr) {
		row := make([]string, len(colExprs))
		for i, expr := range colExprs {
			if n := htmlquery.QuerySelector(node, expr); n != nil {
				row[i] = strings.Join(strings.Fields(htmlquery.InnerText(n)), " ")
			}
		}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}
