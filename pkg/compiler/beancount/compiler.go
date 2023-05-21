package beancount

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"text/template"

	"github.com/deb-sig/double-entry-generator/pkg/analyser"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/io/writer"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

// BeanCount is the implementation.
type BeanCount struct {
	Provider   string
	Target     string
	AppendMode bool
	Output     string
	Config     *config.Config
	IR         *ir.IR

	analyser.Interface
}

// New creates a new BeanCount.
func New(providerName, targetName, output string,
	appendMode bool, c *config.Config, i *ir.IR, a analyser.Interface,
) (*BeanCount, error) {
	b := &BeanCount{
		Provider:   providerName,
		Target:     targetName,
		AppendMode: appendMode,
		Output:     output,
		Config:     c,
		IR:         i,
		Interface:  a,
	}
	err := b.initTemplates()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *BeanCount) initTemplates() error {
	// init the templates
	funcMap := template.FuncMap{
		"EscapeString": util.EscapeString,
	}

	var err error
	normalOrderTemplate, err = template.New("normalOrder").Funcs(funcMap).Parse(normalOrder)
	if err != nil {
		return fmt.Errorf("Failed to init the normalOrder template. %v", err)
	}
	huobiTradeBuyOrderTemplate, err = template.New("tradeBuyOrder").Funcs(funcMap).Parse(huobiTradeBuyOrder)
	if err != nil {
		return fmt.Errorf("Failed to init the tradeBuyOrder template. %v", err)
	}
	huobiTradeBuyOrderDiffCommissionUnitTemplate, err = template.New("tradeBuyOrderDiffCommissionUnit").Funcs(funcMap).Parse(huobiTradeBuyOrderDiffCommissionUnit)
	if err != nil {
		return fmt.Errorf("Failed to init the tradeBuyOrderDiffCommissionUnit template. %v", err)
	}
	huobiTradeSellOrderTemplate, err = template.New("tradeSellOrder").Funcs(funcMap).Parse(huobiTradeSellOrder)
	if err != nil {
		return fmt.Errorf("Failed to init the tradeSellOrder template. %v", err)
	}
	htsecTradeBuyOrderTemplate, err = template.New("httradeBuyOrder").Funcs(funcMap).Parse(htsecTradeBuyOrder)
	if err != nil {
		return fmt.Errorf("Failed to init the httradeBuyOrder template. %v", err)
	}
	htsecTradeSellOrderTemplate, err = template.New("httradeSellOrder").Funcs(funcMap).Parse(htsecTradeSellOrder)
	if err != nil {
		return fmt.Errorf("Failed to init the httradeSellOrder template. %v", err)
	}
	return nil
}

// Compile compiles IR to the given platform.
func (b *BeanCount) Compile() error {
	log.SetPrefix("[Compiler-BeanCount] ")
	log.Printf("Getting the expected account for the bills")
	var orders []ir.Order
	for _, o := range b.IR.Orders {
		// Get the expected accounts according to the configuration.
		ignore, minusAccount, plusAccount, extraAccounts, tags := b.GetAccountsAndTags(&o, b.Config, b.Provider, b.Target)
		if ignore {
			continue
		}
		o.MinusAccount = minusAccount
		o.PlusAccount = plusAccount
		o.ExtraAccounts = extraAccounts
		o.Tags = tags
		orders = append(orders, o)
	}

	b.IR.Orders = orders

	outputWriter, err := writer.GetWriter(b.Output)
	if err != nil {
		return fmt.Errorf("can't get output writer, err: %v", err)
	}
	defer func(outputWriter writer.OutputWriter) {
		err := outputWriter.Close()
		if err != nil {
			log.Printf("output writer close err: %v\n", err)
		}
	}(outputWriter)

	if !b.AppendMode {
		if err := b.writeHeader(outputWriter); err != nil {
			return err
		}
	}

	log.Printf("Finished to write to %s", b.Output)
	return b.writeBills(outputWriter)
}

// writeHeader writes the acounts and title into the file.
func (b *BeanCount) writeHeader(file io.Writer) error {
	_, err := io.WriteString(file, "option \"title\" \""+b.Config.Title+"\"\n")
	if err != nil {
		return fmt.Errorf("write option title error: %v", err)
	}
	_, err = io.WriteString(file,
		"option \"operating_currency\" \""+b.Config.DefaultCurrency+"\"\n\n")
	if err != nil {
		return fmt.Errorf("write option currency error: %v", err)
	}

	accounts := b.GetAllCandidateAccounts(b.Config)
	var sortedAccounts []string
	for k := range accounts {
		if k != "" {
			sortedAccounts = append(sortedAccounts, k)
		}
	}
	sort.Strings(sortedAccounts)

	for _, k := range sortedAccounts {
		_, err = io.WriteString(file, "1970-01-01 open "+k+"\n")
		if err != nil {
			return fmt.Errorf("write open account error: %v", err)
		}
	}
	_, err = io.WriteString(file, "\n")
	if err != nil {
		return fmt.Errorf("write extra enter error: %v", err)
	}
	return nil
}

// writeBills writes bills to the file.
func (b *BeanCount) writeBills(file io.Writer) error {
	// Sort the bills from earliest to lastest.
	// If the bills are the same day, the tx which has lower
	// line number is considered happened earlier than the tx
	// which has a higher line number by beancount default.
	sort.Slice(b.IR.Orders, func(i, j int) bool {
		return b.IR.Orders[i].PayTime.Before(b.IR.Orders[j].PayTime)
	})

	for i := range b.IR.Orders {
		if err := b.writeBill(file, i); err != nil {
			return err
		}
	}
	return nil
}

func (b *BeanCount) writeBill(file io.Writer, index int) error {
	o := b.IR.Orders[index]

	var buf bytes.Buffer
	var err error

	switch o.OrderType {
	default:
		fallthrough
	case ir.OrderTypeNormal:
		err = normalOrderTemplate.Execute(&buf, &NormalOrderVars{
			PayTime:           o.PayTime,
			Peer:              o.Peer,
			Item:              o.Item,
			Note:              o.Note,
			Money:             o.Money,
			Commission:        o.Commission,
			PlusAccount:       o.PlusAccount,
			MinusAccount:      o.MinusAccount,
			PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
			CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
			Metadata:          o.Metadata,
			Currency:          b.Config.DefaultCurrency,
			Tags:              o.Tags,
		})
	case ir.OrderTypeHuobiTrade: // Huobi trades
		switch o.Type {
		case ir.TypeSend: // buy
			isDiffCommissionUnit := false
			commissionUnit, ok := o.Units[ir.CommissionUnit]
			if !ok {
				isDiffCommissionUnit = true
			}
			targetUnit, ok := o.Units[ir.TargetUnit]
			if !ok {
				isDiffCommissionUnit = true
			}
			if commissionUnit != targetUnit {
				// for example, using HT for commission fee.
				isDiffCommissionUnit = true
			}

			if isDiffCommissionUnit {
				err = huobiTradeBuyOrderDiffCommissionUnitTemplate.Execute(&buf, &HuobiTradeBuyOrderVars{
					PayTime:           o.PayTime,
					Peer:              o.Peer,
					TxTypeOriginal:    o.TxTypeOriginal,
					TypeOriginal:      o.TypeOriginal,
					Item:              o.Item,
					Amount:            o.Amount,
					Money:             o.Money,
					Commission:        o.Commission,
					Price:             o.Price,
					CashAccount:       o.ExtraAccounts[ir.CashAccount],
					PositionAccount:   o.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
					BaseUnit:          o.Units[ir.BaseUnit],
					TargetUnit:        o.Units[ir.TargetUnit],
					CommissionUnit:    o.Units[ir.CommissionUnit],
				})
			} else {
				err = huobiTradeBuyOrderTemplate.Execute(&buf, &HuobiTradeBuyOrderVars{
					PayTime:           o.PayTime,
					Peer:              o.Peer,
					TxTypeOriginal:    o.TxTypeOriginal,
					TypeOriginal:      o.TypeOriginal,
					Item:              o.Item,
					Amount:            o.Amount,
					Money:             o.Money,
					Commission:        o.Commission,
					Price:             o.Price,
					CashAccount:       o.ExtraAccounts[ir.CashAccount],
					PositionAccount:   o.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
					BaseUnit:          o.Units[ir.BaseUnit],
					TargetUnit:        o.Units[ir.TargetUnit],
					CommissionUnit:    o.Units[ir.CommissionUnit],
				})
			}
		case ir.TypeRecv: // sell
			err = huobiTradeSellOrderTemplate.Execute(&buf, &HuobiTradeSellOrderVars{
				PayTime:           o.PayTime,
				Peer:              o.Peer,
				TxTypeOriginal:    o.TxTypeOriginal,
				TypeOriginal:      o.TypeOriginal,
				Item:              o.Item,
				Amount:            o.Amount,
				Money:             o.Money,
				Commission:        o.Commission,
				Price:             o.Price,
				CashAccount:       o.ExtraAccounts[ir.CashAccount],
				PositionAccount:   o.ExtraAccounts[ir.PositionAccount],
				CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
				PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
				BaseUnit:          o.Units[ir.BaseUnit],
				TargetUnit:        o.Units[ir.TargetUnit],
				CommissionUnit:    o.Units[ir.CommissionUnit],
			})
		default:
			err = fmt.Errorf("Failed to get the TxType.")
		}
	case ir.OrderTypeSecuritiesTrade:
		switch o.Type {
		case ir.TypeSend: // buy
			err = htsecTradeBuyOrderTemplate.Execute(&buf, &HtsecTradeBuyOrderVars{
				PayTime:           o.PayTime,
				Peer:              o.Peer,
				TxTypeOriginal:    o.TxTypeOriginal,
				TypeOriginal:      o.TypeOriginal,
				Item:              o.Item,
				Amount:            o.Amount,
				Money:             o.Money,
				Commission:        o.Commission,
				Price:             o.Price,
				CashAccount:       o.ExtraAccounts[ir.CashAccount],
				PositionAccount:   o.ExtraAccounts[ir.PositionAccount],
				CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
				PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
				Currency:          b.Config.DefaultCurrency,
			})
		case ir.TypeRecv: // sell
			err = htsecTradeSellOrderTemplate.Execute(&buf, &HtsecTradeSellOrderVars{
				PayTime:           o.PayTime,
				Peer:              o.Peer,
				TxTypeOriginal:    o.TxTypeOriginal,
				TypeOriginal:      o.TypeOriginal,
				Item:              o.Item,
				Amount:            o.Amount,
				Money:             o.Money,
				Commission:        o.Commission,
				Price:             o.Price,
				CashAccount:       o.ExtraAccounts[ir.CashAccount],
				PositionAccount:   o.ExtraAccounts[ir.PositionAccount],
				CommissionAccount: o.ExtraAccounts[ir.CommissionAccount],
				PnlAccount:        o.ExtraAccounts[ir.PnlAccount],
				Currency:          b.Config.DefaultCurrency,
			})
		default:
			err = fmt.Errorf("Failed to get the TxType.")
		}
	}
	if err != nil {
		return err
	}
	if _, err := io.WriteString(file, buf.String()); err != nil {
		return err
	}
	return nil
}
