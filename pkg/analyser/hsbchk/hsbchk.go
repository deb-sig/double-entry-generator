package hsbchk

import (
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

// HsbcHK is the analyser for HSBC HK provider
type HsbcHK struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (h HsbcHK) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.HsbcHK == nil || len(cfg.HsbcHK.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.HsbcHK.Rules {
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

// GetAccountsAndTags returns accounts and tags for HSBC HK provider.
func (h HsbcHK) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.HsbcHK == nil || len(cfg.HsbcHK.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil, nil
	}

	var tags = make([]string, 0)
	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount
	cashAccount := cfg.DefaultCashAccount

	// method account (bank card account)
	if o.Type == ir.TypeRecv {
		resPlus = cashAccount
	} else {
		resMinus = cashAccount
	}

	for _, r := range cfg.HsbcHK.Rules {
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
		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
		}
		if r.Type != nil {
			match = matchFunc(*r.Type, o.TypeOriginal, sep, match)
		}
		if r.Status != nil {
			// 从metadata中获取status信息
			if status, exists := o.Metadata["status"]; exists {
				match = matchFunc(*r.Status, status, sep, match)
			} else {
				match = false
			}
		}
		if r.Time != nil {
			var err error
			match, err = util.SplitFindTimeInterval(*r.Time, o.PayTime, match)
			if err != nil {
				log.Printf("Failed to parse time interval: %v", err)
				continue
			}
		}
		if r.MinPrice != nil && o.Money < *r.MinPrice {
			match = false
		}
		if r.MaxPrice != nil && o.Money > *r.MaxPrice {
			match = false
		}

		if match {
			if r.Ignore {
				ignore = true
				break
			}
			// Support multiple matches, like one rule matches the minus account, the other rule matches the plus account.
			if r.TargetAccount != nil {
				if o.Type == ir.TypeRecv {
					resMinus = *r.TargetAccount
				} else {
					resPlus = *r.TargetAccount
				}
			}

			if r.Tag != nil {
				tags = strings.Split(*r.Tag, sep)
			}
		}
	}

	return ignore, resMinus, resPlus, nil, tags
}
