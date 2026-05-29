package spdb_debit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type SpdbDebit struct{}

// GetAllCandidateAccounts returns all accounts defined in config.
func (SpdbDebit) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniq := make(map[string]bool)
	uniq[cfg.DefaultPlusAccount] = true
	uniq[cfg.DefaultMinusAccount] = true
	uniq[cfg.DefaultCashAccount] = true
	if cfg.SpdbDebit != nil {
		for _, rule := range cfg.SpdbDebit.Rules {
			if rule.TargetAccount != nil {
				uniq[*rule.TargetAccount] = true
			}
		}
	}
	return uniq
}

// GetAccountsAndTags returns accounts determined by rules.
func (SpdbDebit) GetAccountsAndTags(order *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false
	assetAccount := cfg.DefaultMinusAccount
	counterpartyAccount := cfg.DefaultPlusAccount
	var tags = make([]string, 0)

	if cfg.SpdbDebit != nil {
		for _, rule := range cfg.SpdbDebit.Rules {
			match := true
			sep := ","
			if rule.Separator != nil {
				sep = *rule.Separator
			}
			matchFunc := util.SplitFindContains
			if rule.FullMatch {
				matchFunc = util.SplitFindEquals
			}
			if rule.Peer != nil {
				match = matchFunc(*rule.Peer, order.Peer, sep, match)
			}
			if rule.Item != nil {
				match = matchFunc(*rule.Item, order.Item, sep, match)
			}
			if rule.TxType != nil {
				match = matchFunc(*rule.TxType, order.TxTypeOriginal, sep, match)
			}
			if match {
				if rule.TargetAccount != nil {
					counterpartyAccount = *rule.TargetAccount
				}
				if rule.Tags != nil {
					tags = strings.Split(*rule.Tags, sep)
				}
				if rule.Ignore {
					ignore = true
					break
				}
			}
		}
	}

	var minus, plus string
	if order.Type == ir.TypeRecv {
		plus = assetAccount
		minus = counterpartyAccount
	} else {
		plus = counterpartyAccount
		minus = assetAccount
	}

	return ignore, minus, plus, nil, tags
}
