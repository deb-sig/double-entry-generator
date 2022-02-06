package alipay

import (
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

type Alipay struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (a Alipay) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Alipay == nil || len(cfg.Alipay.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Alipay.Rules {
		if r.MethodAccount != nil {
			uniqMap[*r.MethodAccount] = true
		}
		if r.TargetAccount != nil {
			uniqMap[*r.TargetAccount] = true
		}
		if r.PnlAccount != nil {
			uniqMap[*r.PnlAccount] = true
		}
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccounts returns minus and plus account.
func (a Alipay) GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string, map[ir.Account]string) {

	if cfg.Alipay == nil || len(cfg.Alipay.Rules) == 0 {
		return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil
	}
	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount
	var extraAccounts map[ir.Account]string

	var err error
	for _, r := range cfg.Alipay.Rules {
		match := true
		// get seperator
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
			match = matchFunc(*r.Type, o.TxTypeOriginal, sep, match)
		}
		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
		}
		if r.Method != nil {
			match = matchFunc(*r.Method, o.Method, sep, match)
		}
		if r.Category != nil {
			match = matchFunc(*r.Category, o.Category, sep, match)
		}
		if r.Time != nil {
			match, err = util.SplitFindTimeInterval(*r.Time, o.PayTime, match)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		if match {
			// Support multiple matches, like one rule matches the
			// minus accout, the other rule matches the plus account.
			if r.TargetAccount != nil {
				if o.TxType == ir.TxTypeRecv {
					resMinus = *r.TargetAccount
				} else {
					resPlus = *r.TargetAccount
				}
			}
			if r.MethodAccount != nil {
				if o.TxType == ir.TxTypeRecv {
					resPlus = *r.MethodAccount
				} else {
					resMinus = *r.MethodAccount
				}
			}
			if r.PnlAccount != nil {
				extraAccounts = map[ir.Account]string{
					ir.PnlAccount: *r.PnlAccount,
				}
			}

		}
	}

	if strings.HasPrefix(o.Item, "退款-") {
		return resPlus, resMinus, extraAccounts
	}
	return resMinus, resPlus, extraAccounts
}
