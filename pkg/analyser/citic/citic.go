package citic

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type Citic struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (c Citic) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Citic == nil || len(cfg.Citic.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Citic.Rules {
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

// GetAccountsAndTags GetAccounts returns minus and plus account.
func (c Citic) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.Citic == nil || len(cfg.Citic.Rules) == 0 {
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

	//var err error
	for _, r := range cfg.Citic.Rules {
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

		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
		}
		if r.Type != nil {
			match = matchFunc(*r.Type, o.TypeOriginal, sep, match)
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
