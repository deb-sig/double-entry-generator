package jd

import (
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type JD struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (a JD) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.JD == nil || len(cfg.JD.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.JD.Rules {
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
func (a JD) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.JD == nil || len(cfg.JD.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil, nil
	}
	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount
	var extraAccounts map[ir.Account]string
	var tags = make([]string, 0)

	var err error
	for _, r := range cfg.JD.Rules {
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
				log.Fatalf("%v", err)
			}
		}
		if r.TimestampRange != nil {
			match, err = util.SplitFindTimeStampInterval(*r.TimestampRange, o.PayTime, match)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}

		if match {
			if r.Ignore {
				ignore = true
				break
			}

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
			if r.PnlAccount != nil {
				extraAccounts = map[ir.Account]string{
					ir.PnlAccount: *r.PnlAccount,
				}
			}

			if r.Tags != nil {
				tags = strings.Split(*r.Tags, sep)
			}

		}
	}

	if o.TypeOriginal == "不计收支" && (strings.HasPrefix(o.Item, "冻结-") ||
		strings.HasPrefix(o.Item, "解冻-")) {
		ignore = true
	}
	return ignore, resMinus, resPlus, extraAccounts, tags
}
