package compiler

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/pkg/analyser"

	"github.com/deb-sig/double-entry-generator/pkg/compiler/beancount"
	"github.com/deb-sig/double-entry-generator/pkg/compiler/ledger"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/consts"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// Interface is the type for the compiler.
type Interface interface {
	Compile() error
}

// New creates a new compiler.
func New(providerName, targetName, output string,
	appendMode bool, c *config.Config, i *ir.IR) (Interface, error) {
	a, err := analyser.New(providerName)
	if err != nil {
		return nil, err
	}
	switch targetName {
	case consts.CompilerBeanCount:
		return beancount.New(providerName, targetName,
			output, appendMode, c, i, a)
	case consts.CompilerLedger:
		return ledger.New(providerName, targetName, output, appendMode, c, i, a)
	default:
		return nil, fmt.Errorf("Fail to create the compiler for the given name %s", targetName)
	}
}
