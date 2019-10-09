package beancount

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gaocegege/double-entry-generator/pkg/config"
	"github.com/gaocegege/double-entry-generator/pkg/ir"
	"github.com/gaocegege/double-entry-generator/pkg/util"
)

// BeanCount is the implementation.
type BeanCount struct {
	Provider   string
	Target     string
	AppendMode bool
	Output     string
	Config     *config.Config
	IR         *ir.IR
}

// New creates a new BeanCount.
func New(providerName, targetName, output string,
	appendMode bool, c *config.Config, i *ir.IR) *BeanCount {
	return &BeanCount{
		Provider:   providerName,
		Target:     targetName,
		AppendMode: appendMode,
		Output:     output,
		Config:     c,
		IR:         i,
	}
}

// Compile compiles IR to the given platform.
func (b *BeanCount) Compile() error {
	log.SetPrefix("[Compiler-BeanCount] ")
	log.Printf("Getting the expected account for the bills")
	for index, o := range b.IR.Orders {
		// Get the expected accounts according to the configuration.
		minusAccount, plusAccount := util.GetAccounts(&o, b.Config, b.Provider, b.Target)
		b.IR.Orders[index].MinusAccount = minusAccount
		b.IR.Orders[index].PlusAccount = plusAccount
	}

	log.Printf("Writing to %s", b.Output)
	file, err := os.Create(b.Output)
	if err != nil {
		return fmt.Errorf("create output file  %s error: %v", b.Output, err)
	}
	defer file.Close()

	if !b.AppendMode {
		if err := b.writeHeader(file); err != nil {
			return err
		}
	}

	log.Printf("Finished to write to %s", b.Output)
	return b.writeBills(file)
}

// writeHeader writes the acounts and title into the file.
func (b *BeanCount) writeHeader(file *os.File) error {
	_, err := io.WriteString(file, "option \"title\" \""+b.Config.Title+"\"\n")
	if err != nil {
		return fmt.Errorf("write option title error: %v", err)
	}
	_, err = io.WriteString(file,
		"option \"operating_currency\" \""+b.Config.DefaultCurrency+"\"\n\n")
	if err != nil {
		return fmt.Errorf("write option currency error: %v", err)
	}

	accounts := util.GetAllCandidateAccounts(b.Config)
	for k := range accounts {
		if k == "" {
			continue
		}
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
func (b *BeanCount) writeBills(file *os.File) error {
	for i := range b.IR.Orders {
		if err := b.writeBill(file, i); err != nil {
			return err
		}
	}
	return nil
}

func (b *BeanCount) writeBill(file *os.File, index int) error {
	o := b.IR.Orders[index]

	str := o.PayTime.Format("2006-01-02")
	str = str + " * \"" + o.Peer + "\" \"" + o.Item + "\"\n"
	if _, err := io.WriteString(file, str); err != nil {
		return err
	}

	str = "\t"
	str = str + o.PlusAccount + " " + fmt.Sprintf("%.2f", o.Money) + " "
	str = str + b.Config.DefaultCurrency + "\n"
	str = str + "\t" + o.MinusAccount + " -" + fmt.Sprintf("%.2f", o.Money) + " "
	str = str + b.Config.DefaultCurrency + "\n\n"
	if _, err := io.WriteString(file, str); err != nil {
		return err
	}
	return nil
}
