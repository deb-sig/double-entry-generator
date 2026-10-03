package cgb_credit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type CgbCredit struct{}

// GetAllCandidateAccounts 返回配置中可能用到的全部账户。
func (CgbCredit) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniq := make(map[string]bool)

	if cfg.CgbCredit == nil || len(cfg.CgbCredit.Rules) == 0 {
		uniq[cfg.DefaultPlusAccount] = true
		uniq[cfg.DefaultMinusAccount] = true
		uniq[cfg.DefaultCashAccount] = true
		return uniq
	}

	for _, rule := range cfg.CgbCredit.Rules {
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

// GetAccountsAndTags 根据广发信用卡规则决定交易两端账户和标签。
func (CgbCredit) GetAccountsAndTags(order *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	tags := make([]string, 0)
	minus := cfg.DefaultMinusAccount
	plus := cfg.DefaultPlusAccount
	cashAccount := cfg.DefaultCashAccount

	if order.Type == ir.TypeRecv {
		plus = cashAccount
	} else {
		minus = cashAccount
	}

	if cfg.CgbCredit == nil || len(cfg.CgbCredit.Rules) == 0 {
		return ignore, minus, plus, nil, nil
	}

	for _, rule := range cfg.CgbCredit.Rules {
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
		if rule.Peer != nil {
			match = matchFunc(*rule.Peer, order.Peer, sep, match)
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
			if rule.MethodAccount != nil {
				if order.Type == ir.TypeRecv {
					plus = *rule.MethodAccount
				} else {
					minus = *rule.MethodAccount
				}
			}
			if rule.Tag != nil {
				tags = strings.Split(*rule.Tag, sep)
			}
		}
	}

	return ignore, minus, plus, nil, tags
}
