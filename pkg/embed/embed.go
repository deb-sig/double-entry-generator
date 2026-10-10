// Package embed runs one import entirely in memory and returns a result
// that hosts can serialise: the browser runtime (cmd/wasm-runtime) and the
// C library for apps (cmd/libdeg) both go through Run, so the CLI, the
// template site and an embedding app produce the same transactions.
package embed

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/api"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler/beancount"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/deb-sig/double-entry-generator/v2/pkg/io/writer"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// ContractVersion changes whenever a field of Result or Transaction is
// renamed, removed or changes meaning. Adding a field does not bump it.
const ContractVersion = 1

var (
	outputSeq atomic.Int64
	// compileMu serialises compilation: the beancount compiler keeps its
	// text templates in package variables that New re-initialises.
	compileMu sync.Mutex
)

// Result is everything one import produces.
type Result struct {
	Contract int    `json:"contract"`
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	// Beancount is the compiled ledger text, identical to the CLI's output.
	Beancount    string        `json:"beancount,omitempty"`
	Transactions []Transaction `json:"transactions"`
	// NeedsAccount counts transactions where at least one side was filled
	// by the engine (Expenses:FIXME and friends) rather than a rule.
	NeedsAccount int      `json:"needsAccount"`
	Warnings     []string `json:"warnings"`
	Report       string   `json:"report,omitempty"`
}

// Transaction is one imported transaction in a host-friendly shape.
type Transaction struct {
	Date      string      `json:"date"`
	Time      string      `json:"time"`
	Flag      string      `json:"flag,omitempty"`
	Payee     string      `json:"payee"`
	Narration string      `json:"narration"`
	Tags      []string    `json:"tags,omitempty"`
	Links     []string    `json:"links,omitempty"`
	Metadata  [][2]string `json:"metadata"`
	// Postings are the transaction's legs in ledger order.
	Postings []Posting `json:"postings"`
	// From, To, Amount and Currency summarise a plain two-leg transfer
	// (money leaves From, arrives at To). They are empty when the
	// transaction has any other shape; read Postings then.
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Amount   string `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
	// Sources says who wrote each value: the template, a rule (with its
	// id) or the engine.
	Sources      []Source `json:"sources,omitempty"`
	NeedsAccount bool     `json:"needsAccount"`
}

// Posting is one leg. Amount and Currency are empty for a leg Beancount
// balances automatically; Cost is the text inside {}, Price the text after
// @ or @@ (with the marker).
type Posting struct {
	Account  string `json:"account"`
	Amount   string `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
	Cost     string `json:"cost,omitempty"`
	Price    string `json:"price,omitempty"`
	Line     string `json:"line"`
}

// Source mirrors ir.FieldSource.
type Source struct {
	Slot    string   `json:"slot"`
	Rule    string   `json:"rule,omitempty"`
	Columns []string `json:"columns,omitempty"`
	Origin  string   `json:"origin"`
}

// Run imports bill with the given template and rules (rules may be empty).
// It never panics; failures come back as Result.Error.
func Run(templateYAML, rulesYAML, billName string, bill []byte) (res Result) {
	res = Result{Contract: ContractVersion, Transactions: []Transaction{}, Warnings: []string{}}
	defer func() {
		if r := recover(); r != nil {
			res = Result{Contract: ContractVersion, Error: fmt.Sprint(r), Transactions: []Transaction{}, Warnings: []string{}}
		}
	}()
	profile, err := importer.LoadProfileBytes([]byte(templateYAML), "template")
	if err != nil {
		res.Error = "template: " + err.Error()
		return res
	}
	if strings.TrimSpace(rulesYAML) != "" {
		rf, err := importer.ParseRulesFile([]byte(rulesYAML))
		if err != nil {
			res.Error = err.Error()
			return res
		}
		profile.ApplyRulesFile(rf)
		res.Warnings = append(res.Warnings, importer.PersonalRuleWarnings(profile, rf.AllPersonalRules(), "", "")...)
	}
	out, report, err := importer.ImportBytes(profile, billName, bill)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	text, err := compile(profile, out)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.OK = true
	res.Beancount = text
	res.Report = report.String()
	currency := firstNonEmpty(profile.Template.DefaultCurrency, "CNY")
	for _, o := range out.Orders {
		t := transaction(o, currency)
		if t.NeedsAccount {
			res.NeedsAccount++
		}
		res.Transactions = append(res.Transactions, t)
	}
	return res
}

// JSON is Run serialised; marshalling errors become an error result.
func JSON(templateYAML, rulesYAML, billName string, bill []byte) []byte {
	b, err := json.Marshal(Run(templateYAML, rulesYAML, billName, bill))
	if err != nil {
		b, _ = json.Marshal(Result{Contract: ContractVersion, Error: err.Error(), Transactions: []Transaction{}, Warnings: []string{}})
	}
	return b
}

func compile(profile *importer.Profile, out *ir.IR) (string, error) {
	compileMu.Lock()
	defer compileMu.Unlock()
	cfg := &config.Config{
		Title:               firstNonEmpty(profile.Name, profile.ID, "DEG Import"),
		DefaultMinusAccount: profile.Template.DefaultMinus,
		DefaultPlusAccount:  profile.Template.DefaultPlus,
		DefaultCurrency:     firstNonEmpty(profile.Template.DefaultCurrency, "CNY"),
	}
	name := fmt.Sprintf("%sembed-%d", writer.MemoryPrefix, outputSeq.Add(1))
	c, err := beancount.New(importer.DefaultProviderName, consts.CompilerBeanCount, name, false, cfg, out, api.Runtime{})
	if err != nil {
		return "", err
	}
	if err := c.Compile(); err != nil {
		return "", err
	}
	text, ok := writer.TakeMemory(name)
	if !ok {
		return "", fmt.Errorf("no in-memory output")
	}
	return text, nil
}

func transaction(o ir.Order, defaultCurrency string) Transaction {
	t := Transaction{
		Date:      o.PayTime.Format("2006-01-02"),
		Time:      o.PayTime.Format(time.RFC3339),
		Flag:      o.Flag,
		Payee:     o.Peer,
		Narration: o.Item,
		Tags:      o.Tags,
		Links:     o.Links,
		Metadata:  metadata(o),
	}
	if len(o.Postings) > 0 {
		for _, p := range o.Postings {
			t.Postings = append(t.Postings, parsePosting(p.Line))
		}
		summariseTransfer(&t)
	} else {
		t.From, t.To = o.MinusAccount, o.PlusAccount
		t.Currency = firstNonEmpty(o.Currency, defaultCurrency)
		t.Amount = amountText(o)
		t.Postings = []Posting{
			{Account: t.To, Amount: t.Amount, Currency: t.Currency, Line: t.To + " " + t.Amount + " " + t.Currency},
			{Account: t.From, Amount: "-" + t.Amount, Currency: t.Currency, Line: t.From + " -" + t.Amount + " " + t.Currency},
		}
	}
	for _, s := range o.Sources {
		t.Sources = append(t.Sources, Source{Slot: s.Slot, Rule: s.RuleID, Columns: s.Columns, Origin: string(s.Origin)})
		if s.Origin == ir.FieldOriginEngine {
			t.NeedsAccount = true
		}
	}
	return t
}

// parsePosting splits a rendered posting line ("Account 1.00 CNY {..} @ ..").
func parsePosting(line string) Posting {
	p := Posting{Line: line}
	rest := strings.TrimSpace(line)
	if i := strings.Index(rest, "@"); i >= 0 {
		p.Price = strings.TrimSpace(rest[i:])
		rest = strings.TrimSpace(rest[:i])
	}
	if i := strings.Index(rest, "{"); i >= 0 {
		if j := strings.LastIndex(rest, "}"); j > i {
			p.Cost = strings.TrimSpace(rest[i+1 : j])
			rest = strings.TrimSpace(rest[:i] + rest[j+1:])
		}
	}
	fields := strings.Fields(rest)
	if len(fields) > 0 {
		p.Account = fields[0]
	}
	if len(fields) > 1 {
		p.Amount = fields[1]
	}
	if len(fields) > 2 {
		p.Currency = fields[2]
	}
	return p
}

// summariseTransfer fills From/To/Amount/Currency when the legs are one
// outflow and one inflow of the same amount, with no cost or price.
func summariseTransfer(t *Transaction) {
	if len(t.Postings) != 2 {
		return
	}
	a, b := t.Postings[0], t.Postings[1]
	if a.Cost != "" || a.Price != "" || b.Cost != "" || b.Price != "" || a.Currency != b.Currency || a.Amount == "" {
		return
	}
	in, out := a, b
	if strings.HasPrefix(a.Amount, "-") {
		in, out = b, a
	}
	if strings.TrimPrefix(out.Amount, "-") != in.Amount || !strings.HasPrefix(out.Amount, "-") {
		return
	}
	t.From, t.To, t.Amount, t.Currency = out.Account, in.Account, in.Amount, in.Currency
}

// metadata keeps the template's declared key order; keys the template did
// not declare follow sorted.
func metadata(o ir.Order) [][2]string {
	out := [][2]string{}
	seen := map[string]bool{}
	for _, k := range o.MetadataKeys {
		if v, ok := o.Metadata[k]; ok && !seen[k] {
			out = append(out, [2]string{k, v})
			seen[k] = true
		}
	}
	rest := make([]string, 0, len(o.Metadata))
	for k := range o.Metadata {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	for _, k := range rest {
		out = append(out, [2]string{k, o.Metadata[k]})
	}
	return out
}

func amountText(o ir.Order) string {
	scale := uint32(2)
	if o.OrderType == ir.OrderTypeCrypto {
		scale = 8
	}
	if o.ExactMoney != nil {
		return o.ExactMoney.Text(scale)
	}
	return strconv.FormatFloat(o.Money, 'f', int(scale), 64)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
