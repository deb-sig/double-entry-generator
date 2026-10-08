package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// The text reader is for layout-preserving text, chiefly what
// `pdftotext -layout` produces from a PDF statement. A record regexp with
// named groups turns matching lines into rows; a continuation regexp folds
// wrapped lines into the previous row. Lines matching neither are ignored,
// so page headers, footers and totals need no special handling.

func readText(name string, data []byte, cfg Config) (Table, error) {
	isPDF := strings.EqualFold(filepath.Ext(name), ".pdf") || bytes.HasPrefix(data, []byte("%PDF"))
	if isPDF {
		if cfg.Convert == "" {
			return Table{}, fmt.Errorf("%s is a pdf but the text reader has no `convert`; set convert: pdftotext-layout", filepath.Base(name))
		}
		converted, err := convert(cfg.Convert, name, data)
		if err != nil {
			return Table{}, err
		}
		data = converted
	} else {
		decoded, err := decodeText(data, cfg.Encoding)
		if err != nil {
			return Table{}, err
		}
		data = decoded
	}

	if strings.TrimSpace(cfg.Record) == "" {
		return Table{}, fmt.Errorf("text reader requires `record` (a regexp with named groups)")
	}
	record, err := regexp.Compile(cfg.Record)
	if err != nil {
		return Table{}, fmt.Errorf("record regexp: %w", err)
	}
	headers := namedGroups(record)
	if len(headers) == 0 {
		return Table{}, fmt.Errorf("record regexp has no named groups; use (?P<name>...) to declare columns")
	}
	index := make(map[string]int, len(headers))
	for i, h := range headers {
		index[h] = i
	}

	var continuation *regexp.Regexp
	if strings.TrimSpace(cfg.Continuation) != "" {
		continuation, err = regexp.Compile(cfg.Continuation)
		if err != nil {
			return Table{}, fmt.Errorf("continuation regexp: %w", err)
		}
		for _, g := range namedGroups(continuation) {
			if _, ok := index[g]; !ok {
				return Table{}, fmt.Errorf("continuation group %q is not a column of `record`", g)
			}
		}
	}

	table := Table{Headers: headers}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if m := record.FindStringSubmatch(line); m != nil {
			row := make([]string, len(headers))
			for i, g := range record.SubexpNames() {
				if g != "" {
					row[index[g]] = strings.TrimSpace(m[i])
				}
			}
			table.Rows = append(table.Rows, row)
			continue
		}
		if continuation == nil || len(table.Rows) == 0 {
			continue
		}
		if m := continuation.FindStringSubmatch(line); m != nil {
			row := table.Rows[len(table.Rows)-1]
			for i, g := range continuation.SubexpNames() {
				if g == "" {
					continue
				}
				if v := strings.TrimSpace(m[i]); v != "" {
					row[index[g]] = strings.TrimSpace(row[index[g]] + " " + v)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Table{}, err
	}
	return table, nil
}

func namedGroups(re *regexp.Regexp) []string {
	var out []string
	for _, name := range re.SubexpNames() {
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}
