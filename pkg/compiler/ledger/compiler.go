package ledger

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"sort"

	"github.com/deb-sig/double-entry-generator/pkg/analyser"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/io/writer"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

// Ledger is the implementation
type Ledger struct {
	Provider   string
	Target     string
	AppendMode bool
	Output     string
	Config     *config.Config
	IR         *ir.IR

	analyser.Interface
}

func New(providerName, targetName, ouput string, appendMode bool, config *config.Config,
	ir *ir.IR, analyser analyser.Interface) (*Ledger, error) {
	ledger := &Ledger{
		Provider:   providerName,
		Target:     targetName,
		AppendMode: appendMode,
		Output:     ouput,
		Config:     config,
		IR:         ir,
		Interface:  analyser,
	}

	err := ledger.initTemplates()
	if err != nil {
		return nil, err
	}

	return ledger, nil
}

func (ledger *Ledger) initTemplates() error {
	funcMap := template.FuncMap{
		"EscapeString": util.EscapeString,
	}
	var err error
	normalOrderTemplate, err = template.New("normalOrder").Funcs(funcMap).Parse(normalOrder)

	if err != nil {
		return fmt.Errorf("Failed to init the normalOrder Template. %v", err)
	}
	return nil
}

// Compile compiles IR to the given platform.
func (ledger *Ledger) Compile() error {
	log.SetPrefix("[Compiler-Ledger] ")
	log.Printf("Getting the expected account for the bills")
	var orders []ir.Order
	for _, o := range ledger.IR.Orders {
		// Get the expected accounts according to the configuration.
		ignore, minusAccount, plusAccount, extraAccounts, tags := ledger.GetAccountsAndTags(&o, ledger.Config, ledger.Provider, ledger.Target)
		if ignore {
			continue
		}
		o.MinusAccount = minusAccount
		o.PlusAccount = plusAccount
		o.ExtraAccounts = extraAccounts
		o.Tags = tags
		orders = append(orders, o)
	}

	ledger.IR.Orders = orders

	outputWriter, err := writer.GetWriter(ledger.Output)
	if err != nil {
		return fmt.Errorf("can't get output writer, err: %v", err)
	}
	defer func(outputWriter writer.OutputWriter) {
		err := outputWriter.Close()
		if err != nil {
			log.Printf("output writer close err: %v\n", err)
		}
	}(outputWriter)

	if !ledger.AppendMode {
		if err := ledger.writeHeader(outputWriter); err != nil {
			return err
		}
	}

	log.Printf("Finished to write to %s", ledger.Output)
	return ledger.writeBills(outputWriter)
}

// writeHeader writes the acounts and title into the file.
func (ledger *Ledger) writeHeader(file io.Writer) error {
	var err error

	accounts := ledger.GetAllCandidateAccounts(ledger.Config)
	var sortedAccounts []string
	for k := range accounts {
		if k != "" {
			sortedAccounts = append(sortedAccounts, k)
		}
	}
	sort.Strings(sortedAccounts)

	_, err = io.WriteString(file, "1970-01-01 * Open Balance\n")
	if err != nil {
		return fmt.Errorf("write open account error: %v", err)
	}

	for _, k := range sortedAccounts {
		_, err = io.WriteString(file, "    "+k+"     0 "+ledger.Config.DefaultCurrency+"\n")
		if err != nil {
			return fmt.Errorf("write open account error: %v", err)
		}
	}
	_, err = io.WriteString(file, "    Equity:Opening Balances \n")
	if err != nil {
		return fmt.Errorf("write extra enter error: %v", err)
	}
	return nil
}

// writeBills writes bills to the file.
func (ledger *Ledger) writeBills(file io.Writer) error {
	// Sort the bills from earliest to lastest.
	// If the bills are the same day, the transaction which has lower
	// line number is considered happened earlier than the transaction
	// which has a higher line number.
	sort.Slice(ledger.IR.Orders, func(i, j int) bool {
		return ledger.IR.Orders[i].PayTime.Before(ledger.IR.Orders[j].PayTime)
	})

	for i := range ledger.IR.Orders {
		if err := ledger.writeBill(file, i); err != nil {
			return err
		}
	}
	return nil
}

func (ledger *Ledger) writeBill(file io.Writer, index int) error {
	order := ledger.IR.Orders[index]

	var buf bytes.Buffer
	var err error

	switch order.OrderType {
	default:
		fallthrough
	case ir.OrderTypeNormal:
		err = normalOrderTemplate.Execute(&buf, &NormalOrderVars{
			PayTime:           order.PayTime,
			Peer:              order.Peer,
			Item:              order.Item,
			Note:              order.Note,
			Amount:            order.Money,
			Commission:        order.Commission,
			PlusAccount:       order.PlusAccount,
			MinusAccount:      order.MinusAccount,
			PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
			CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
			Metadata:          order.Metadata,
			Currency:          ledger.Config.DefaultCurrency,
		})
	}
	if err != nil {
		return err
	}
	if _, err := io.WriteString(file, buf.String()); err != nil {
		return err
	}
	return nil
}
