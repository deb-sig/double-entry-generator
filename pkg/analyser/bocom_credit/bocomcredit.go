package bocomcredit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type BocomCredit struct{}

// GetAllCandidateAccounts returns all accounts defined in config.
func (BocomCredit) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniq := make(map[string]bool)

	if cfg.BocomCredit == nil || len(cfg.BocomCredit.Rules) == 0 {
		uniq[cfg.DefaultPlusAccount] = true
		uniq[cfg.DefaultMinusAccount] = true
		uniq[cfg.DefaultCashAccount] = true
		return uniq
	}

	for _, rule := range cfg.BocomCredit.Rules {
		if rule.MethodAccount != nil {
			uniq[*rule.MethodAccount] = true
		}
		if rule.TargetAccount != nil {
			uniq[*rule.TargetAccount] = true
		}
	}

	uniq[cfg.DefaultPlusAccount] = true
	uniq[cfg.DefaultMinusAccount] = true
	uniq[cfg.DefaultCashAccount] = true
	return uniq
}

// GetAccountsAndTags returns accounts determined by rules.
func (BocomCredit) GetAccountsAndTags(order *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.BocomCredit == nil || len(cfg.BocomCredit.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil, nil
	}

	tags := make([]string, 0)
	minus := cfg.DefaultMinusAccount
	plus := cfg.DefaultPlusAccount
	cashAccount := cfg.DefaultCashAccount

	if order.Type == ir.TypeRecv {
		plus = cashAccount
	} else {
		minus = cashAccount
	}

	for _, rule := range cfg.BocomCredit.Rules {
		match := true
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
			if rule.TargetAccount != nil {
				if order.Type == ir.TypeRecv {
					minus = *rule.TargetAccount
				} else {
					plus = *rule.TargetAccount
				}
			}
			if rule.Tag != nil {
				tags = strings.Split(*rule.Tag, sep)
			}
		}
	}

	return ignore, minus, plus, nil, tags
}
