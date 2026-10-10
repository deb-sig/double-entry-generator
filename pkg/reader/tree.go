package reader

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

// Tree formats (json, xml) share one contract: `records` is an XPath that
// selects one node per row, and each column is an XPath evaluated relative
// to that node. One query language for both keeps templates uniform.

func compileTree(cfg Config) (*xpath.Expr, []*xpath.Expr, error) {
	if strings.TrimSpace(cfg.Records) == "" {
		return nil, nil, fmt.Errorf("%s reader requires `records` (an XPath selecting one node per row)", cfg.Format)
	}
	if cfg.Columns.Len() == 0 {
		return nil, nil, fmt.Errorf("%s reader requires `columns` (name → XPath relative to each record)", cfg.Format)
	}
	records, err := xpath.Compile(cfg.Records)
	if err != nil {
		return nil, nil, fmt.Errorf("records %q: %w", cfg.Records, err)
	}
	cols := make([]*xpath.Expr, 0, cfg.Columns.Len())
	for _, name := range cfg.Columns.Keys {
		expr, err := xpath.Compile(cfg.Columns.Values[name])
		if err != nil {
			return nil, nil, fmt.Errorf("column %q path %q: %w", name, cfg.Columns.Values[name], err)
		}
		cols = append(cols, expr)
	}
	return records, cols, nil
}

func readJSON(data []byte, cfg Config) (Table, error) {
	data, err := decodeText(data, cfg.Encoding)
	if err != nil {
		return Table{}, err
	}
	recordsExpr, colExprs, err := compileTree(cfg)
	if err != nil {
		return Table{}, err
	}
	doc, err := jsonquery.Parse(bytes.NewReader(data))
	if err != nil {
		return Table{}, fmt.Errorf("parse json: %w", err)
	}
	table := Table{Headers: append([]string{}, cfg.Columns.Keys...)}
	for _, node := range jsonquery.QuerySelectorAll(doc, recordsExpr) {
		row := make([]string, len(colExprs))
		for i, expr := range colExprs {
			row[i] = jsonText(jsonquery.QuerySelector(node, expr))
		}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}

func jsonText(n *jsonquery.Node) string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.InnerText())
}

func readXML(data []byte, cfg Config) (Table, error) {
	data, err := decodeText(data, cfg.Encoding)
	if err != nil {
		return Table{}, err
	}
	recordsExpr, colExprs, err := compileTree(cfg)
	if err != nil {
		return Table{}, err
	}
	doc, err := xmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		return Table{}, fmt.Errorf("parse xml: %w", err)
	}
	table := Table{Headers: append([]string{}, cfg.Columns.Keys...)}
	for _, node := range xmlquery.QuerySelectorAll(doc, recordsExpr) {
		row := make([]string, len(colExprs))
		for i, expr := range colExprs {
			row[i] = xmlText(xmlquery.QuerySelector(node, expr))
		}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}

func xmlText(n *xmlquery.Node) string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.InnerText())
}
