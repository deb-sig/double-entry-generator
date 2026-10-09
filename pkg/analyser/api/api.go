// Package api holds the analyser interface and the runtime analyser on
// their own, so the compilers and the browser build of the template
// importer do not drag in every provider-specific analyser.
package api

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// Interface is what a compiler asks of an analyser.
type Interface interface {
	GetAllCandidateAccounts(cfg *config.Config) map[string]bool
	GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string)
}

// Runtime is the analyser used by template-driven imports. Accounts and tags
// are already produced by the rule engine, so this analyser only fills defaults.
type Runtime struct{}

func (Runtime) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	accounts := map[string]bool{}
	for _, account := range []string{
		cfg.DefaultMinusAccount,
		cfg.DefaultPlusAccount,
		cfg.DefaultCashAccount,
		cfg.DefaultPositionAccount,
		cfg.DefaultCommissionAccount,
		cfg.DefaultPnlAccount,
		cfg.DefaultThirdPartyCustodyAccount,
	} {
		if account != "" {
			accounts[account] = true
		}
	}
	return accounts
}

func (Runtime) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	minus := o.MinusAccount
	if minus == "" {
		minus = cfg.DefaultMinusAccount
	}
	plus := o.PlusAccount
	if plus == "" {
		plus = cfg.DefaultPlusAccount
	}
	return false, minus, plus, o.ExtraAccounts, o.Tags
}
