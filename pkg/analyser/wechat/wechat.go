package wechat

import (
	"log"

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
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccounts returns minus and plus account.
func (w Wechat) GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string, map[ir.Account]string) {
	var resCommission string
	// check this tx whether has commission
	if o.Commission != 0 {
		if cfg.DefaultCommissionAccount == "" {
			log.Fatalf("Found a tx with commission, but not setting the `defaultCommissionAccount` in config file!")
		} else {
			resCommission = cfg.DefaultCommissionAccount
		}
	}

	if cfg.Wechat == nil || len(cfg.Wechat.Rules) == 0 {
		return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, map[ir.Account]string{
			ir.CommissionAccount: resCommission,
		}
	}

	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount

	for _, r := range cfg.Wechat.Rules {
		match := true
		// get seperator
		sep := ","
		if r.Seperator != nil {
			sep = *r.Seperator
		}
		if r.Peer != nil {
			match = util.SplitFindContains(*r.Peer, o.Peer, sep, match)
		}
		if r.Type != nil {
			match = util.SplitFindContains(*r.Type, o.TxTypeOriginal, sep, match)
		}
		if r.TxType != nil {
			match = util.SplitFindContains(*r.TxType, o.TypeOriginal, sep, match)
		}
		if r.Method != nil {
			match = util.SplitFindContains(*r.Method, o.Method, sep, match)
		}
		if r.Item != nil {
			match = util.SplitFindContains(*r.Item, o.Item, sep, match)
		}
		if r.StartTime != nil && r.EndTime != nil {
			// TODO(gaocegege): Support it.
		}
		if match {
			// Support multiple matches, like one rule matches the minus accout, the other rule matches the plus account.
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
			if r.CommissionAccount != nil {
				resCommission = *r.CommissionAccount
			}
		}

	}

	return resMinus, resPlus, map[ir.Account]string{
		ir.CommissionAccount: resCommission,
	}
}
