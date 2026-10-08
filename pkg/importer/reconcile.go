package importer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// The reconcile stage decides what gets written to the ledger. It runs
// after every transaction is built: running-balance assertions catch a
// misread row (the safety net for layout-text readers), exact keys drop
// transactions already in the ledger or repeated in the bill, and near
// matches or engine-filled accounts are flagged for review instead of
// being written as settled.

// Balance is the template's running-balance column, if the bill has one.
type Balance struct {
	// Column is an expression yielding the balance after this row.
	Column string `json:"column,omitempty" yaml:"column,omitempty"`
}

// Reconcile is the user's `reconcile:` block in the rules file.
type Reconcile struct {
	Dedupe *Dedupe `json:"dedupe,omitempty" yaml:"dedupe,omitempty"`
	// FlagEngineFilled sets this flag (usually "!") on transactions where
	// the engine had to fill a FIXME account.
	FlagEngineFilled string `json:"flagEngineFilled,omitempty" yaml:"flagEngineFilled,omitempty"`
}

type Dedupe struct {
	// Key names the fields that identify a transaction: date, payee,
	// narration, amount, or metadata.<key>. A transaction whose key is
	// incomplete is never deduplicated.
	Key FlexStrings `json:"key,omitempty" yaml:"key,omitempty"`
	// Against is an existing ledger file whose transactions count as
	// already imported.
	Against string `json:"against,omitempty" yaml:"against,omitempty"`
	// Window (e.g. "2d") marks a transaction as a near duplicate when the
	// ledger has one within the window with the same absolute amount.
	Window string `json:"window,omitempty" yaml:"window,omitempty"`
	// Flag set on near duplicates. Default "!".
	Flag string `json:"flag,omitempty" yaml:"flag,omitempty"`
}

// ReconcileReport says what the stage did.
type ReconcileReport struct {
	DroppedInBill   int
	DroppedInLedger int
	Flagged         int
}

func (r ReconcileReport) String() string {
	return fmt.Sprintf("dedupe: %d repeated in bill, %d already in ledger, %d flagged for review", r.DroppedInBill, r.DroppedInLedger, r.Flagged)
}

type rowOrder struct {
	row   Row
	order ir.Order
}

// checkRunningBalance verifies that consecutive balances differ by the
// transaction amount, in either row order. Rows without a balance are
// skipped. The first decidable pair fixes the order for the rest.
func checkRunningBalance(profile *Profile, pairs []rowOrder) error {
	expr := strings.TrimSpace(profile.Template.Balance.Column)
	if expr == "" {
		return nil
	}
	type point struct {
		index   int
		balance ir.Decimal
		amount  ir.Decimal
	}
	var points []point
	for i, p := range pairs {
		text := strings.TrimSpace(renderRuleText(expr, p.row, p.order))
		if text == "" || p.order.ExactMoney == nil {
			continue
		}
		balance, err := parseAmountExact(text, profile.Template.AmountPrefix)
		if err != nil {
			return fmt.Errorf("balance column %q row %d: %q: %w", expr, i+1, text, err)
		}
		// ExactMoney is absolute; Type carries the direction.
		amount := p.order.ExactMoney.Abs()
		if p.order.Type == ir.TypeSend {
			amount = amount.Neg()
		}
		points = append(points, point{index: i + 1, balance: balance, amount: amount})
	}
	ascending := 0 // +1 oldest first, -1 newest first, 0 undecided
	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		diff, err := cur.balance.Sub(prev.balance)
		if err != nil {
			return err
		}
		asc := diff.Cmp(cur.amount) == 0         // balance grew by this row's amount
		desc := diff.Neg().Cmp(prev.amount) == 0 // previous row is the later one
		broken := !asc && !desc
		switch {
		case broken, ascending == 1 && !asc, ascending == -1 && !desc:
			return fmt.Errorf("running balance breaks between rows %d and %d: balance %s -> %s but amount is %s (row %d) / %s (row %d); a row was misread or is missing",
				prev.index, cur.index, prev.balance.Text(2), cur.balance.Text(2), prev.amount.Text(2), prev.index, cur.amount.Text(2), cur.index)
		case ascending == 0 && asc && !desc:
			ascending = 1
		case ascending == 0 && desc && !asc:
			ascending = -1
		}
	}
	return nil
}

// applyReconcile drops duplicates and flags uncertain transactions.
func applyReconcile(profile *Profile, orders []ir.Order) ([]ir.Order, ReconcileReport, error) {
	var report ReconcileReport
	rc := profile.Reconcile
	if rc == nil {
		return orders, report, nil
	}
	out := orders[:0:0]
	var ledger *ledgerIndex
	if rc.Dedupe != nil && strings.TrimSpace(rc.Dedupe.Against) != "" {
		var err error
		ledger, err = loadLedgerIndex(expandHome(rc.Dedupe.Against))
		if err != nil {
			return nil, report, fmt.Errorf("reconcile.dedupe.against: %w", err)
		}
		ledger.indexKeys(rc.Dedupe.Key)
	}
	window, err := parseWindow(rc.Dedupe)
	if err != nil {
		return nil, report, err
	}
	seen := map[string]struct{}{}
	for _, order := range orders {
		if rc.Dedupe != nil {
			key := dedupeKey(rc.Dedupe.Key, order)
			if key != "" {
				if _, dup := seen[key]; dup {
					report.DroppedInBill++
					continue
				}
				seen[key] = struct{}{}
				if ledger != nil && ledger.hasKey(key) {
					report.DroppedInLedger++
					continue
				}
			}
			if ledger != nil && window > 0 && ledger.hasNear(order, window) {
				order.Flag = firstNonEmptyString(rc.Dedupe.Flag, "!")
				report.Flagged++
			}
		}
		if rc.FlagEngineFilled != "" && hasEngineFilledSide(order) {
			order.Flag = rc.FlagEngineFilled
			report.Flagged++
		}
		out = append(out, order)
	}
	return out, report, nil
}

func parseWindow(d *Dedupe) (time.Duration, error) {
	if d == nil || strings.TrimSpace(d.Window) == "" {
		return 0, nil
	}
	w := strings.TrimSpace(d.Window)
	if strings.HasSuffix(w, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(w, "d"))
		if err != nil {
			return 0, fmt.Errorf("reconcile.dedupe.window %q: %w", w, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	dur, err := time.ParseDuration(w)
	if err != nil {
		return 0, fmt.Errorf("reconcile.dedupe.window %q: use e.g. 2d or 36h", w)
	}
	return dur, nil
}

func hasEngineFilledSide(order ir.Order) bool {
	for _, src := range order.Sources {
		if src.Origin == ir.FieldOriginEngine {
			return true
		}
	}
	return false
}

// dedupeKey joins the configured fields. Any empty field makes the key
// empty, so partial information never causes a drop.
func dedupeKey(fields []string, order ir.Order) string {
	if len(fields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		v := orderField(strings.TrimSpace(f), order)
		if v == "" {
			return ""
		}
		parts = append(parts, v)
	}
	return strings.Join(parts, "\x1f")
}

func orderField(field string, order ir.Order) string {
	switch {
	case field == "date":
		if order.PayTime.IsZero() {
			return ""
		}
		return order.PayTime.Format("2006-01-02")
	case field == "payee":
		return strings.TrimSpace(order.Peer)
	case field == "narration":
		return strings.TrimSpace(order.Item)
	case field == "amount":
		if order.ExactMoney == nil {
			return ""
		}
		return order.ExactMoney.Abs().Normalize().Text(0)
	case strings.HasPrefix(field, "metadata."):
		return strings.TrimSpace(order.Metadata[strings.TrimPrefix(field, "metadata.")])
	default:
		return ""
	}
}

// ledgerIndex is what the reconcile stage knows about an existing ledger:
// enough of each transaction to compute the same keys the bill produces.
type ledgerIndex struct {
	txns []ledgerTxn
	keys map[string]struct{}
}

type ledgerTxn struct {
	date      time.Time
	payee     string
	narration string
	metadata  map[string]string
	amounts   []ir.Decimal // absolute posting amounts
}

var (
	ledgerTxnPattern     = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s+(?:\*|!|txn|[A-Z])\s*(.*)$`)
	ledgerMetaPattern    = regexp.MustCompile(`^\s+([a-z][A-Za-z0-9_-]*):\s*"(.*)"\s*$`)
	ledgerPostingPattern = regexp.MustCompile(`^\s+([A-Z][A-Za-z0-9:\-\p{L}]*)\s+(-?[\d,]+(?:\.\d+)?)\s+\S+`)
	quotedPattern        = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

func loadLedgerIndex(path string) (*ledgerIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	idx := &ledgerIndex{keys: map[string]struct{}{}}
	var cur *ledgerTxn
	flush := func() {
		if cur != nil {
			idx.txns = append(idx.txns, *cur)
			cur = nil
		}
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if m := ledgerTxnPattern.FindStringSubmatch(line); m != nil {
			flush()
			date, err := time.Parse("2006-01-02", m[1])
			if err != nil {
				continue
			}
			cur = &ledgerTxn{date: date, metadata: map[string]string{}}
			quoted := quotedPattern.FindAllStringSubmatch(m[2], -1)
			switch len(quoted) {
			case 1:
				cur.narration = quoted[0][1]
			default:
				if len(quoted) >= 2 {
					cur.payee, cur.narration = quoted[0][1], quoted[1][1]
				}
			}
			continue
		}
		if cur == nil {
			continue
		}
		if strings.TrimSpace(line) == "" || !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			flush()
			continue
		}
		if m := ledgerMetaPattern.FindStringSubmatch(line); m != nil {
			cur.metadata[m[1]] = m[2]
			continue
		}
		if m := ledgerPostingPattern.FindStringSubmatch(line); m != nil {
			if d, err := ir.ParseDecimal(strings.ReplaceAll(m[2], ",", "")); err == nil {
				cur.amounts = append(cur.amounts, d.Abs().Normalize())
			}
		}
	}
	flush()
	return idx, scanner.Err()
}

// indexKeys computes the dedupe keys of every ledger transaction once the
// key fields are known.
func (l *ledgerIndex) indexKeys(fields []string) {
	for _, t := range l.txns {
		order := ir.Order{PayTime: t.date, Peer: t.payee, Item: t.narration, Metadata: t.metadata}
		if len(t.amounts) > 0 {
			amount := t.amounts[0]
			order.ExactMoney = &amount
		}
		if key := dedupeKey(fields, order); key != "" {
			l.keys[key] = struct{}{}
		}
	}
}

func (l *ledgerIndex) hasKey(key string) bool {
	_, ok := l.keys[key]
	return ok
}

func (l *ledgerIndex) hasNear(order ir.Order, window time.Duration) bool {
	if order.ExactMoney == nil || order.PayTime.IsZero() {
		return false
	}
	want := order.ExactMoney.Abs().Normalize()
	for _, t := range l.txns {
		gap := order.PayTime.Sub(t.date)
		if gap < 0 {
			gap = -gap
		}
		if gap > window {
			continue
		}
		for _, a := range t.amounts {
			if a.Cmp(want) == 0 {
				return true
			}
		}
	}
	return false
}
