package importer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/reader"
)

// The shape stage sits between the reader and the rules. It answers one
// question: which rows are one transaction? Everything about rows that is
// not a one-to-one bill line (header position, summary lines, failed
// orders, split rows, one line carrying two postings) is settled here, so
// the mapping and rule layers only ever see one record per transaction.

// ShapeOp is one step of the `shape:` list. Exactly one field is set.
type ShapeOp struct {
	// LocateHeader finds the header row by anchor column names instead of
	// counting leading rows. Rows before it are dropped.
	LocateHeader *LocateHeader `json:"locateHeader,omitempty" yaml:"locateHeader,omitempty"`
	// DropMatching drops any row whose cells, joined by a space, match the
	// regexp. Works before and after the header is located.
	DropMatching string `json:"dropMatching,omitempty" yaml:"dropMatching,omitempty"`
	// DropIf drops records for which the condition holds. Same syntax as a
	// rule's `when`, evaluated against the raw columns.
	DropIf string `json:"dropIf,omitempty" yaml:"dropIf,omitempty"`
	// Merge folds records sharing a key into one.
	Merge *MergeOp `json:"merge,omitempty" yaml:"merge,omitempty"`
	// Split fans one record out into several.
	Split *SplitOp `json:"split,omitempty" yaml:"split,omitempty"`
	// Capture reads statement-level values (account number, card alias,
	// statement month) from rows outside the table, usually the notice
	// lines above the header. Named groups become <file.name>.
	Capture *CaptureOp `json:"capture,omitempty" yaml:"capture,omitempty"`
	// FailIf stops the import when any record matches, naming the row.
	// Use it to catch an export whose format changed instead of writing
	// wrong entries.
	FailIf string `json:"failIf,omitempty" yaml:"failIf,omitempty"`
	// Message explains a FailIf to the person importing.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type CaptureOp struct {
	// Pattern is a regexp with named groups, matched against each row's
	// cells joined by a space. The first non-empty match of each group
	// wins; the import fails if the pattern matches no row at all.
	Pattern string `json:"pattern" yaml:"pattern"`
	// ScanRows bounds the search from the top of the remaining rows.
	// 0 means every row.
	ScanRows int `json:"scanRows,omitempty" yaml:"scanRows,omitempty"`
}

type LocateHeader struct {
	// Anchor is one or more column names the header row must contain.
	Anchor FlexStrings `json:"anchor,omitempty" yaml:"anchor,omitempty"`
	// ScanRows bounds the search from the top. 0 means the whole table.
	ScanRows int `json:"scanRows,omitempty" yaml:"scanRows,omitempty"`
}

type MergeOp struct {
	// Key expressions (column refs or templates) identifying one group.
	// Records with an empty key are left alone.
	Key FlexStrings `json:"key,omitempty" yaml:"key,omitempty"`
	// Take says how each column is combined: first (default), last,
	// first-nonzero, sum, join.
	Take map[string]string `json:"take,omitempty" yaml:"take,omitempty"`
}

type SplitOp struct {
	// When limits the split to matching records. Empty means every record.
	When string `json:"when,omitempty" yaml:"when,omitempty"`
	// Into lists the records to emit in place of the original. Each entry
	// overrides columns; an empty entry is a copy of the original. Values
	// are templates rendered against the original record.
	Into []map[string]string `json:"into,omitempty" yaml:"into,omitempty"`
}

// FlexStrings accepts a scalar or a list in yaml.
type FlexStrings []string

func (f *FlexStrings) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var one string
	if err := unmarshal(&one); err == nil {
		*f = FlexStrings{one}
		return nil
	}
	var many []string
	if err := unmarshal(&many); err != nil {
		return err
	}
	*f = FlexStrings(many)
	return nil
}

// shapeTable runs the ops and returns the header names and the records.
func shapeTable(profile *Profile, table reader.Table) ([]string, [][]string, error) {
	headers, rows, _, err := shapeTableFile(profile, table)
	return headers, rows, err
}

// shapeTableFile is shapeTable plus the statement-level values captured
// along the way, keyed by group name.
func shapeTableFile(profile *Profile, table reader.Table) ([]string, [][]string, map[string]string, error) {
	headers := table.Headers
	rows := table.Rows
	file := map[string]string{}
	for i, op := range profile.Shape {
		var err error
		switch {
		case op.Capture != nil:
			err = shapeCapture(*op.Capture, rows, file)
		case op.FailIf != "":
			headers, rows, err = ensureHeaders(profile, headers, rows)
			if err == nil {
				err = shapeFailIf(profile, op.FailIf, op.Message, headers, rows, file)
			}
		case op.LocateHeader != nil:
			if headers != nil {
				return nil, nil, nil, fmt.Errorf("shape[%d] locateHeader: header already known", i)
			}
			headers, rows, err = shapeLocateHeader(*op.LocateHeader, rows)
		case op.DropMatching != "":
			rows, err = shapeDropMatching(op.DropMatching, rows)
		case op.DropIf != "":
			headers, rows, err = ensureHeaders(profile, headers, rows)
			if err == nil {
				rows, err = shapeDropIf(profile, op.DropIf, headers, rows, file)
			}
		case op.Merge != nil:
			headers, rows, err = ensureHeaders(profile, headers, rows)
			if err == nil {
				rows, err = shapeMerge(profile, *op.Merge, headers, rows, file)
			}
		case op.Split != nil:
			headers, rows, err = ensureHeaders(profile, headers, rows)
			if err == nil {
				rows, err = shapeSplit(profile, *op.Split, headers, rows, file)
			}
		default:
			return nil, nil, nil, fmt.Errorf("shape[%d]: empty step", i)
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("shape[%d]: %w", i, err)
		}
	}
	headers, rows, err := ensureHeaders(profile, headers, rows)
	if err != nil {
		return nil, nil, nil, err
	}
	return headers, rows, file, nil
}

func shapeCapture(op CaptureOp, rows [][]string, file map[string]string) error {
	re, err := regexp.Compile(op.Pattern)
	if err != nil {
		return fmt.Errorf("capture %q: %w", op.Pattern, err)
	}
	names := namedGroupNames(re)
	if len(names) == 0 {
		return fmt.Errorf("capture %q has no named groups; use (?P<name>...)", op.Pattern)
	}
	end := len(rows)
	if op.ScanRows > 0 && op.ScanRows < end {
		end = op.ScanRows
	}
	for i := 0; i < end; i++ {
		if rows[i] == nil {
			continue
		}
		m := re.FindStringSubmatch(strings.Join(normalizeCells(rows[i]), " "))
		if m == nil {
			continue
		}
		for gi, name := range re.SubexpNames() {
			if name == "" {
				continue
			}
			// A matched but empty group (a blank card alias) is a value too;
			// a later non-empty match still replaces it.
			if have, seen := file[name]; !seen || have == "" {
				file[name] = strings.TrimSpace(m[gi])
			}
		}
	}
	for _, name := range names {
		if _, ok := file[name]; !ok {
			return fmt.Errorf("capture %q: group %q not found in the first %d rows", op.Pattern, name, end)
		}
	}
	return nil
}

func namedGroupNames(re *regexp.Regexp) []string {
	var out []string
	for _, n := range re.SubexpNames() {
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func shapeFailIf(profile *Profile, cond, message string, headers []string, rows [][]string, file map[string]string) error {
	for i, row := range rows {
		if emptyRecord(row) {
			continue
		}
		hit, err := evalWhen(cond, shapeRow(profile, headers, row, file), ir.Order{})
		if err != nil {
			return fmt.Errorf("failIf %q: %w", cond, err)
		}
		if !hit {
			continue
		}
		if message == "" {
			message = "the statement has a row this template does not understand"
		}
		return fmt.Errorf("%s (data row %d: %s; failIf %s)", message, i+1, strings.Join(normalizeCells(row), " | "), cond)
	}
	return nil
}

// ensureHeaders falls back to the legacy header rules (skipLeadingRows,
// sourceHeaders, headerLocate) when no shape step located the header.
func ensureHeaders(profile *Profile, headers []string, rows [][]string) ([]string, [][]string, error) {
	if headers != nil {
		return headers, rows, nil
	}
	return legacyHeaders(profile, rows)
}

func shapeLocateHeader(op LocateHeader, rows [][]string) ([]string, [][]string, error) {
	anchors := normalizeCells(op.Anchor)
	if len(anchors) == 0 || emptyRecord(anchors) {
		return nil, nil, fmt.Errorf("locateHeader needs at least one anchor column name")
	}
	wanted := make(map[string]struct{}, len(anchors))
	for _, a := range anchors {
		if a != "" {
			wanted[a] = struct{}{}
		}
	}
	end := len(rows)
	if op.ScanRows > 0 && op.ScanRows < end {
		end = op.ScanRows
	}
	for i := 0; i < end; i++ {
		cells := normalizeCells(rows[i])
		if emptyRecord(cells) || !rowContainsAllHeaders(cells, wanted) {
			continue
		}
		if err := rejectDuplicateHeaderNames(cells); err != nil {
			return nil, nil, fmt.Errorf("header row %d: %w", i, err)
		}
		return cells, rows[i+1:], nil
	}
	return nil, nil, fmt.Errorf("locateHeader: no row within the first %d rows contains %s", end, strings.Join(anchors, ", "))
}

func shapeDropMatching(pattern string, rows [][]string) ([][]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("dropMatching %q: %w", pattern, err)
	}
	out := rows[:0:0]
	for _, row := range rows {
		if row != nil && re.MatchString(strings.Join(normalizeCells(row), " ")) {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func shapeDropIf(profile *Profile, cond string, headers []string, rows [][]string, file map[string]string) ([][]string, error) {
	out := rows[:0:0]
	for _, row := range rows {
		if emptyRecord(row) {
			continue
		}
		drop, err := evalWhen(cond, shapeRow(profile, headers, row, file), ir.Order{})
		if err != nil {
			return nil, fmt.Errorf("dropIf %q: %w", cond, err)
		}
		if !drop {
			out = append(out, row)
		}
	}
	return out, nil
}

func shapeMerge(profile *Profile, op MergeOp, headers []string, rows [][]string, file map[string]string) ([][]string, error) {
	if len(op.Key) == 0 {
		return nil, fmt.Errorf("merge needs a key")
	}
	index := headerIndex(headers)
	for col, mode := range op.Take {
		if _, ok := index[col]; !ok {
			return nil, fmt.Errorf("merge.take column %q is not a header", col)
		}
		switch mode {
		case "", "first", "last", "first-nonzero", "sum", "join":
		default:
			return nil, fmt.Errorf("merge.take %q: unknown mode %q (first, last, first-nonzero, sum, join)", col, mode)
		}
	}
	var out [][]string
	groups := map[string]int{}
	for _, row := range rows {
		if emptyRecord(row) {
			continue
		}
		raw := shapeRow(profile, headers, row, file)
		parts := make([]string, len(op.Key))
		for i, k := range op.Key {
			parts[i] = strings.TrimSpace(renderRuleText(k, raw, ir.Order{}))
		}
		key := strings.Join(parts, "\x00")
		if strings.Trim(key, "\x00") == "" {
			out = append(out, row)
			continue
		}
		at, seen := groups[key]
		if !seen {
			groups[key] = len(out)
			out = append(out, padRow(row, len(headers)))
			continue
		}
		merged := out[at]
		for i, h := range headers {
			merged[i] = mergeCell(op.Take[h], merged[i], cell(row, i), profile.Template.AmountPrefix)
		}
	}
	return out, nil
}

func mergeCell(mode, have, next, amountPrefix string) string {
	switch mode {
	case "last":
		if strings.TrimSpace(next) != "" {
			return next
		}
		return have
	case "first-nonzero":
		if isNonzeroCell(have, amountPrefix) {
			return have
		}
		if isNonzeroCell(next, amountPrefix) {
			return next
		}
		if strings.TrimSpace(have) == "" {
			return next
		}
		return have
	case "sum":
		a, errA := parseAmountExact(have, amountPrefix)
		b, errB := parseAmountExact(next, amountPrefix)
		if errA != nil {
			if errB != nil {
				return have
			}
			return next
		}
		if errB != nil {
			return have
		}
		sum, err := a.Add(b)
		if err != nil {
			return have
		}
		// Keep the widest scale seen so "5.00" + "0.30" prints "5.30".
		return sum.Text(max(decimalScale(have), decimalScale(next)))
	case "join":
		have, next = strings.TrimSpace(have), strings.TrimSpace(next)
		switch {
		case have == "":
			return next
		case next == "" || next == have:
			return have
		default:
			return have + " " + next
		}
	default: // first
		if strings.TrimSpace(have) == "" {
			return next
		}
		return have
	}
}

// decimalScale counts fraction digits in a cell's spelling.
func decimalScale(value string) uint32 {
	value = strings.TrimSpace(value)
	if i := strings.LastIndexAny(value, ".,"); i >= 0 && !strings.ContainsAny(value[i+1:], ".,") {
		n := 0
		for _, r := range value[i+1:] {
			if r < '0' || r > '9' {
				break
			}
			n++
		}
		return uint32(n)
	}
	return 0
}

func isNonzeroCell(value, amountPrefix string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	d, err := parseAmountExact(value, amountPrefix)
	if err != nil {
		return true // not a number: any text counts as a value
	}
	return !d.IsZero()
}

func shapeSplit(profile *Profile, op SplitOp, headers []string, rows [][]string, file map[string]string) ([][]string, error) {
	if len(op.Into) == 0 {
		return nil, fmt.Errorf("split needs `into`")
	}
	index := headerIndex(headers)
	for n, into := range op.Into {
		for col := range into {
			if _, ok := index[col]; !ok {
				return nil, fmt.Errorf("split.into[%d] column %q is not a header", n, col)
			}
		}
	}
	var out [][]string
	for _, row := range rows {
		if emptyRecord(row) {
			continue
		}
		raw := shapeRow(profile, headers, row, file)
		if strings.TrimSpace(op.When) != "" {
			hit, err := evalWhen(op.When, raw, ir.Order{})
			if err != nil {
				return nil, fmt.Errorf("split.when %q: %w", op.When, err)
			}
			if !hit {
				out = append(out, row)
				continue
			}
		}
		for _, into := range op.Into {
			copied := padRow(row, len(headers))
			for col, tmpl := range into {
				copied[index[col]] = strings.TrimSpace(renderRuleText(tmpl, raw, ir.Order{}))
			}
			out = append(out, copied)
		}
	}
	return out, nil
}

func rawRow(headers []string, row []string) Row {
	// Same cell cleanup as the rows rules see (BOM, ="…" wrappers, quotes).
	cells := normalizeCells(row)
	raw := make(map[string]string, len(headers))
	for i, h := range headers {
		raw[h] = strings.TrimSpace(cell(cells, i))
	}
	return Row{Raw: raw, Metadata: map[string]string{}}
}

// shapeRow is a record as conditions see it during shaping: its columns,
// the captured <file.x> values and the template's time zone.
func shapeRow(profile *Profile, headers []string, row []string, file map[string]string) Row {
	r := rawRow(headers, row)
	for k, v := range file {
		r.Raw["file."+k] = v
	}
	r.Loc = profile.Template.Location()
	return r
}

func headerIndex(headers []string) map[string]int {
	index := make(map[string]int, len(headers))
	for i, h := range headers {
		index[h] = i
	}
	return index
}

func padRow(row []string, width int) []string {
	out := make([]string, width)
	copy(out, normalizeCells(row))
	return out
}
