package erc20

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// ERC20 is the analyser for ERC20.
type ERC20 struct{}

// GetAllCandidateAccounts returns all accounts defined in config.
func (e ERC20) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniqMap := make(map[string]bool)
	
	// Add default accounts from global config
	if cfg.DefaultMinusAccount != "" {
		uniqMap[cfg.DefaultMinusAccount] = true
	}
	if cfg.DefaultPlusAccount != "" {
		uniqMap[cfg.DefaultPlusAccount] = true
	}
	
	if cfg.ERC20 != nil {
		// Add all accounts from rules
		for _, rule := range cfg.ERC20.Rules {
			if rule.TargetAccount != nil {
				uniqMap[*rule.TargetAccount] = true
			}
			if rule.MethodAccount != nil {
				uniqMap[*rule.MethodAccount] = true
			}
		}
	}
	return uniqMap
}

// GetAccountsForTransaction returns the accounts for a given transaction.
func (e ERC20) GetAccountsForTransaction(o *ir.Order) []string {
	accounts := make([]string, 0)
	if o.MinusAccount != "" {
		accounts = append(accounts, o.MinusAccount)
	}
	if o.PlusAccount != "" {
		accounts = append(accounts, o.PlusAccount)
	}
	return accounts
}

// GetAccountsAndTags returns the accounts for analysis.
func (e ERC20) GetAccountsAndTags(o *ir.Order, c *config.Config, plusAccount string, minusAccount string) (bool, string, string, map[ir.Account]string, []string) {
	// Use the accounts already set in the Order
	return false, o.MinusAccount, o.PlusAccount, nil, nil
}

