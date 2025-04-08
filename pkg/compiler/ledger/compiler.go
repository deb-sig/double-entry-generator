package ledger

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
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

	huobiTradeBuyOrderTemplate, err = template.New("tradeBuyOrder").Funcs(funcMap).Parse((huobiTradeBuyOrder))
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
	etfMergeOrderLedgerTemplate, err = template.New("etfMergeOrderLedger").Funcs(funcMap).Parse(etfMergeOrderLedger) // Initialize new template
	if err != nil {
		return fmt.Errorf("Failed to init the etfMergeOrderLedger template. %v", err)
	}

	return nil
}

// Compile compiles IR to the given platform.
func (ledger *Ledger) Compile() error {
	log.SetPrefix("[Compiler-Ledger] ")
	log.Printf("Getting the expected account for the bills")
	var orders []ir.Order
	for _, order := range ledger.IR.Orders {
		// Get the expected accounts according to the configuration.
		ignore, minusAccount, plusAccount, extraAccounts, tags := ledger.GetAccountsAndTags(&order, ledger.Config, ledger.Provider, ledger.Target)
		if ignore {
			continue
		}
		order.MinusAccount = minusAccount
		order.PlusAccount = plusAccount
		order.ExtraAccounts = extraAccounts
		order.Tags = tags
		orders = append(orders, order)
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

	_, err = io.WriteString(file, "1970/01/01 * Open Balance\n")
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
			Money:             order.Money,
			Commission:        order.Commission,
			PlusAccount:       order.PlusAccount,
			MinusAccount:      order.MinusAccount,
			PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
			CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
			Metadata:          order.Metadata,
			Currency:          ledger.Config.DefaultCurrency,
			Tags:              order.Tags,
		})
	case ir.OrderTypeHuobiTrade: // Huobi trades
		switch order.Type {
		case ir.TypeSend: // buy
			isDiffCommissionUnit := false
			commissionUnit, ok := order.Units[ir.CommissionUnit]
			if !ok {
				isDiffCommissionUnit = true
			}
			targetUnit, ok := order.Units[ir.TargetUnit]
			if !ok {
				isDiffCommissionUnit = true
			}
			if commissionUnit != targetUnit {
				// for example, using HT for commission fee.
				isDiffCommissionUnit = true
			}

			if isDiffCommissionUnit {
				err = huobiTradeBuyOrderDiffCommissionUnitTemplate.Execute(&buf, &HuobiTradeBuyOrderVars{
					PayTime:           order.PayTime,
					Peer:              order.Peer,
					TxTypeOriginal:    order.TxTypeOriginal,
					TypeOriginal:      order.TypeOriginal,
					Item:              order.Item,
					Amount:            order.Amount,
					Money:             order.Money,
					Commission:        order.Commission,
					Price:             order.Price,
					CashAccount:       order.ExtraAccounts[ir.CashAccount],
					PositionAccount:   order.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
					BaseUnit:          order.Units[ir.BaseUnit],
					TargetUnit:        order.Units[ir.TargetUnit],
					CommissionUnit:    order.Units[ir.CommissionUnit],
				})
			} else {
				err = huobiTradeBuyOrderTemplate.Execute(&buf, &HuobiTradeBuyOrderVars{
					PayTime:           order.PayTime,
					Peer:              order.Peer,
					TxTypeOriginal:    order.TxTypeOriginal,
					TypeOriginal:      order.TypeOriginal,
					Item:              order.Item,
					Amount:            order.Amount,
					Money:             order.Money,
					Commission:        order.Commission,
					Price:             order.Price,
					CashAccount:       order.ExtraAccounts[ir.CashAccount],
					PositionAccount:   order.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
					BaseUnit:          order.Units[ir.BaseUnit],
					TargetUnit:        order.Units[ir.TargetUnit],
					CommissionUnit:    order.Units[ir.CommissionUnit],
				})
			}
		case ir.TypeRecv: // sell
			err = huobiTradeSellOrderTemplate.Execute(&buf, &HuobiTradeSellOrderVars{
				PayTime:           order.PayTime,
				Peer:              order.Peer,
				TxTypeOriginal:    order.TxTypeOriginal,
				TypeOriginal:      order.TypeOriginal,
				Item:              order.Item,
				Amount:            order.Amount,
				Money:             order.Money,
				Commission:        order.Commission,
				Price:             order.Price,
				CashAccount:       order.ExtraAccounts[ir.CashAccount],
				PositionAccount:   order.ExtraAccounts[ir.PositionAccount],
				CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
				PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
				BaseUnit:          order.Units[ir.BaseUnit],
				TargetUnit:        order.Units[ir.TargetUnit],
				CommissionUnit:    order.Units[ir.CommissionUnit],
			})
		default:
			err = fmt.Errorf("Failed to get the TxType.")
		}

	case ir.OrderTypeSecuritiesTrade:
		// Special handling for Hxsec based on original type
		switch order.TypeOriginal {
		case "银行转证券":
			// Bank to broker cash account
			err = normalOrderTemplate.Execute(&buf, &NormalOrderVars{
				PayTime:      order.PayTime,
				Peer:         order.Peer,
				Item:         "银行转证券", // Simplified item
				Note:         order.Note,
				Money:        order.Money,
				Commission:   0, // No commission for transfers
				PlusAccount:  order.ExtraAccounts[ir.CashAccount],
				MinusAccount: order.ExtraAccounts[ir.ThirdPartyCustodyAccount], // Use the correct constant
				// PnlAccount and CommissionAccount are not typically used here
				Metadata: order.Metadata,
				Currency: ledger.Config.DefaultCurrency,
				Tags:     order.Tags,
			})
		case "证券转银行":
			// Broker cash account to Bank
			err = normalOrderTemplate.Execute(&buf, &NormalOrderVars{
				PayTime:      order.PayTime,
				Peer:         order.Peer,
				Item:         "证券转银行", // Simplified item
				Note:         order.Note,
				Money:        math.Abs(order.Money), // Use absolute value
				Commission:   0,                     // No commission for transfers
				PlusAccount:  order.ExtraAccounts[ir.ThirdPartyCustodyAccount], // Money goes TO the bank
				MinusAccount: order.ExtraAccounts[ir.CashAccount],              // Money comes FROM the broker cash
				// PnlAccount and CommissionAccount are not typically used here
				Metadata: order.Metadata,
				Currency: ledger.Config.DefaultCurrency,
				Tags:     order.Tags,
			})
		case "利息归本":
			err = normalOrderTemplate.Execute(&buf, &NormalOrderVars{
				PayTime:      order.PayTime,
				Peer:         order.Peer,
				Item:         "利息归本", // Simplified item
				Note:         order.Note,
				Money:        order.Money,
				Commission:   0, // No commission for interest
				PlusAccount:  order.ExtraAccounts[ir.CashAccount],
				MinusAccount: order.ExtraAccounts[ir.PnlAccount], // Interest comes from PnL account
				// CommissionAccount not used here
				Metadata: order.Metadata,
				Currency: ledger.Config.DefaultCurrency,
				Tags:     order.Tags,
			})
		case "ETF份额合并":
			// Handle ETF Share Merge (e.g., reverse split)
			// Generates two postings: one removing old shares, one adding new shares.
			newAmountStr, ok := order.Metadata["new_amount"]
			if !ok {
				err = fmt.Errorf("missing 'new_amount' metadata for ETF份额合并 transaction")
				break // Exit the inner switch
			}
			// Note: order.Amount already holds the *removed* amount from the provider logic

			// Remove new_amount from metadata before passing to template
			delete(order.Metadata, "new_amount")

			// Use the new template
			err = etfMergeOrderLedgerTemplate.Execute(&buf, &EtfMergeOrderVars{
				PayTime:         order.PayTime,
				Peer:            order.Peer,
				TypeOriginal:    order.TypeOriginal,
				Item:            order.Item,
				PositionAccount: order.ExtraAccounts[ir.PositionAccount],
				RemovedAmount:   order.Amount,
				AddedAmount:     newAmountStr,
				TxTypeOriginal:  order.TxTypeOriginal,
				Metadata:        order.Metadata, // Pass the whole map, template will filter
			})

		default: // Handle actual trades (Buy/Sell/etc.)
			switch order.Type {
			case ir.TypeSend: // buy
				err = htsecTradeBuyOrderTemplate.Execute(&buf, &HtsecTradeBuyOrderVars{
					PayTime:           order.PayTime,
					Peer:              order.Peer,
					TxTypeOriginal:    order.TxTypeOriginal,
					TypeOriginal:      order.TypeOriginal,
					Item:              order.Item,
					Amount:            order.Amount,
					Money:             order.Money,
					Commission:        order.Commission,
					Price:             order.Price,
					CashAccount:       order.ExtraAccounts[ir.CashAccount],
					PositionAccount:   order.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
					Currency:          ledger.Config.DefaultCurrency,
					Metadata:          order.Metadata,
				})
			case ir.TypeRecv: // sell
				err = htsecTradeSellOrderTemplate.Execute(&buf, &HtsecTradeSellOrderVars{
					PayTime:           order.PayTime,
					Peer:              order.Peer,
					TxTypeOriginal:    order.TxTypeOriginal,
					TypeOriginal:      order.TypeOriginal,
					Item:              order.Item,
					Amount:            order.Amount,
					Money:             order.Money,
					Commission:        order.Commission,
					Price:             order.Price,
					CashAccount:       order.ExtraAccounts[ir.CashAccount],
					PositionAccount:   order.ExtraAccounts[ir.PositionAccount],
					CommissionAccount: order.ExtraAccounts[ir.CommissionAccount],
					PnlAccount:        order.ExtraAccounts[ir.PnlAccount],
					Currency:          ledger.Config.DefaultCurrency,
					Metadata:          order.Metadata,
				})
			default:
				err = fmt.Errorf("Failed to get the TxType.")
			}
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
