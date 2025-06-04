package hxsec

import (
	"log"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type Hxsec struct {
}

func (h Hxsec) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	// Use Hxsec config section
	if cfg.Hxsec == nil || len(cfg.Hxsec.Rules) == 0 {
		// Still add defaults even if no rules
		uniqMap[cfg.DefaultCashAccount] = true
		uniqMap[cfg.DefaultPositionAccount] = true
		uniqMap[cfg.DefaultCommissionAccount] = true
		uniqMap[cfg.DefaultPnlAccount] = true
		uniqMap[cfg.DefaultThirdPartyCustodyAccount] = true
		return uniqMap
	}

	// Add accounts from rules
	for _, r := range cfg.Hxsec.Rules {
		if r.CashAccount != nil {
			uniqMap[*r.CashAccount] = true
		}
		if r.PositionAccount != nil {
			uniqMap[*r.PositionAccount] = true
		}
		if r.PnlAccount != nil {
			uniqMap[*r.PnlAccount] = true
		}
		if r.CommissionAccount != nil {
			uniqMap[*r.CommissionAccount] = true
		}
	}
	// Add defaults
	uniqMap[cfg.DefaultCashAccount] = true
	uniqMap[cfg.DefaultPositionAccount] = true
	uniqMap[cfg.DefaultCommissionAccount] = true
	uniqMap[cfg.DefaultPnlAccount] = true
	uniqMap[cfg.DefaultThirdPartyCustodyAccount] = true
	return uniqMap
}

func (h Hxsec) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false
	// Initialize accounts with defaults
	cashAccount := cfg.DefaultCashAccount
	positionAccount := cfg.DefaultPositionAccount
	commissionAccount := cfg.DefaultCommissionAccount
	pnlAccount := cfg.DefaultPnlAccount
	thirdPartyCustodyAccount := cfg.DefaultThirdPartyCustodyAccount // Use global default

	// If no rules, return defaults
	if cfg.Hxsec == nil || len(cfg.Hxsec.Rules) == 0 {
		return ignore, "", "", map[ir.Account]string{
			ir.CashAccount:              cashAccount,
			ir.PositionAccount:          positionAccount,
			ir.CommissionAccount:        commissionAccount,
			ir.PnlAccount:               pnlAccount,
			ir.ThirdPartyCustodyAccount: thirdPartyCustodyAccount, // Return global default
		}, nil
	}

	var err error
	// Iterate through rules to find a match
	for _, r := range cfg.Hxsec.Rules {
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

		if r.Type != nil {
			match = matchFunc(*r.Type, o.TypeOriginal, sep, match)
		}
		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
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
			if r.CashAccount != nil {
				cashAccount = *r.CashAccount
			}
			if r.PositionAccount != nil {
				positionAccount = *r.PositionAccount
			}
			if r.CommissionAccount != nil {
				commissionAccount = *r.CommissionAccount
			}
			if r.PnlAccount != nil {
				pnlAccount = *r.PnlAccount
			}
			// If a match is found, break the loop (assuming first match wins)
			break
		}
	}

	// Return the determined accounts
	return ignore, "", "", map[ir.Account]string{
		ir.CashAccount:              cashAccount,
		ir.PositionAccount:          positionAccount,
		ir.CommissionAccount:        commissionAccount,
		ir.PnlAccount:               pnlAccount,
		ir.ThirdPartyCustodyAccount: thirdPartyCustodyAccount, // Return global default
	}, nil
}
