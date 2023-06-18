package bmo

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

type Bmo struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (bmo Bmo) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Bmo == nil || len(cfg.Bmo.Rules) == 0 {
		return uniqMap
	}

	for _, rule := range cfg.Bmo.Rules {
		if rule.MethodAccount != nil {
			uniqMap[*rule.MethodAccount] = true
		}
		if rule.TargetAccount != nil {
			uniqMap[*rule.TargetAccount] = true
		}
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccountsAndTags GetAccounts returns minus and plus account.
func (bmo Bmo) GetAccountsAndTags(order *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.Bmo == nil || len(cfg.Bmo.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil, nil
	}

	var tags = make([]string, 0)
	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount
	cashAccount := cfg.DefaultCashAccount

	// method account (bank card account)
	if order.Type == ir.TypeRecv {
		resPlus = cashAccount
	} else {
		resMinus = cashAccount
	}

	for _, rule := range cfg.Bmo.Rules {
		match := true
		// get separator
		sep := ","
		if rule.Separator != nil {
			sep = *rule.Separator
		}

		matchFunc := util.SplitFindContains
		if rule.FullMatch {
			matchFunc = util.SplitFindEquals
		}

		if rule.Item != nil {
			match = matchFunc(*rule.Item, order.Item, sep, match)
		}
		if rule.Type != nil {
			match = matchFunc(*rule.Type, order.TypeOriginal, sep, match)
		}

		if match {
			if rule.Ignore {
				ignore = true
				break
			}
			// Support multiple matches, like one rule matches the minus account, the other rule matches the plus account.
			if rule.TargetAccount != nil {
				if order.Type == ir.TypeRecv {
					resMinus = *rule.TargetAccount
				} else {
					resPlus = *rule.TargetAccount
				}
			}

			if rule.Tag != nil {
				tags = strings.Split(*rule.Tag, sep)
			}

		}

	}
	return ignore, resMinus, resPlus, nil, tags
}
