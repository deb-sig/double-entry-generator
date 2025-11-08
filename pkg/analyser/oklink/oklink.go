package oklink

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// OKLink is the analyser for OKLink.
type OKLink struct{}

// GetAllCandidateAccounts returns all accounts defined in config.
func (e OKLink) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniqMap := make(map[string]bool)
	
	// Add default accounts from global config
	if cfg.DefaultMinusAccount != "" {
		uniqMap[cfg.DefaultMinusAccount] = true
	}
	if cfg.DefaultPlusAccount != "" {
		uniqMap[cfg.DefaultPlusAccount] = true
	}
	
	if cfg.OKLink != nil && cfg.OKLink.Addresses != nil {
		// 从所有地址的配置中获取账户
		for _, addrConfig := range cfg.OKLink.Addresses {
			for _, rule := range addrConfig.Rules {
				if rule.TargetAccount != nil {
					uniqMap[*rule.TargetAccount] = true
				}
				if rule.MethodAccount != nil {
					uniqMap[*rule.MethodAccount] = true
				}
			}
		}
	}
	return uniqMap
}

// GetAccountsForTransaction returns the accounts for a given transaction.
func (e OKLink) GetAccountsForTransaction(o *ir.Order) []string {
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
func (e OKLink) GetAccountsAndTags(o *ir.Order, c *config.Config, plusAccount string, minusAccount string) (bool, string, string, map[ir.Account]string, []string) {
	// Use the accounts already set in the Order
	return false, o.MinusAccount, o.PlusAccount, nil, nil
}

