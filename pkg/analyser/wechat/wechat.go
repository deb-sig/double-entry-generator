package wechat

import (
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

type Wechat struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (w Wechat) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Wechat == nil || len(cfg.Wechat.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Wechat.Rules {
		if r.MethodAccount != nil {
			uniqMap[*r.MethodAccount] = true
		}
		if r.TargetAccount != nil {
			uniqMap[*r.TargetAccount] = true
		}
		if r.CommissionAccount != nil {
			uniqMap[*r.CommissionAccount] = true
		}
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccounts returns minus and plus account.
func (w Wechat) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	var resCommission string
	var tags = make([]string, 0)
	ignore := false

	// check this tx whether it has commission
	if o.Commission != 0 {
		if cfg.DefaultCommissionAccount == "" {
			log.Fatalf("Found a tx with commission, but not setting the `defaultCommissionAccount` in config file!")
		} else {
			resCommission = cfg.DefaultCommissionAccount
		}
	}

	if cfg.Wechat == nil || len(cfg.Wechat.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, map[ir.Account]string{
			ir.CommissionAccount: resCommission,
		}, nil
	}

	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount

	var err error
	for _, r := range cfg.Wechat.Rules {
		match := true
		// get separator
		sep := ","
		if r.Separator != nil {
			sep = *r.Separator
		}

		matchFunc := util.SplitFindContains
		if r.FullMatch {
			matchFunc = util.SplitFindEquals
		}

		if r.Peer != nil {
			match = matchFunc(*r.Peer, o.Peer, sep, match)
		}
		if r.Type != nil {
			match = matchFunc(*r.Type, o.TypeOriginal, sep, match)
		}
		if r.TxType != nil {
			match = matchFunc(*r.TxType, o.TxTypeOriginal, sep, match)
		}
		if r.Method != nil {
			match = matchFunc(*r.Method, o.Method, sep, match)
		}
		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
		}
		if r.Time != nil {
			match, err = util.SplitFindTimeInterval(*r.Time, o.PayTime, match)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		if r.TimestampRange != nil {
			match, err = util.SplitFindTimeStampInterval(*r.TimestampRange, o.PayTime, match)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}

		if match {
			if r.Ignore {
				ignore = true
				break
			}

			// Support multiple matches, like one rule matches the minus accout, the other rule matches the plus account.
			if r.TargetAccount != nil {
				if o.Type == ir.TypeRecv {
					resMinus = *r.TargetAccount
				} else {
					resPlus = *r.TargetAccount
				}
			}
			if r.MethodAccount != nil {
				if o.Type == ir.TypeRecv {
					resPlus = *r.MethodAccount
				} else {
					resMinus = *r.MethodAccount
				}
			}
			if r.CommissionAccount != nil {
				resCommission = *r.CommissionAccount
			}

			if r.Tag != nil {
				tags = strings.Split(*r.Tag, sep)
			}

		}

	}

	return ignore, resMinus, resPlus, map[ir.Account]string{
		ir.CommissionAccount: resCommission,
	}, tags
}
