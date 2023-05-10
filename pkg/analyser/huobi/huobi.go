package huobi

import (
	"log"

	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

type Huobi struct {
}

func (h Huobi) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Huobi == nil || len(cfg.Huobi.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Huobi.Rules {
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
	uniqMap[cfg.DefaultCashAccount] = true
	uniqMap[cfg.DefaultPositionAccount] = true
	uniqMap[cfg.DefaultCommissionAccount] = true
	uniqMap[cfg.DefaultPnlAccount] = true
	return uniqMap
}

func (h Huobi) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false
	if cfg.Huobi == nil || len(cfg.Huobi.Rules) == 0 {
		return ignore, "", "", map[ir.Account]string{
			ir.CashAccount:       cfg.DefaultCashAccount,
			ir.PositionAccount:   cfg.DefaultPositionAccount,
			ir.CommissionAccount: cfg.DefaultCommissionAccount,
			ir.PnlAccount:        cfg.DefaultPnlAccount,
		}, nil
	}

	cashAccount := cfg.DefaultCashAccount
	positionAccount := cfg.DefaultPositionAccount
	commissionAccount := cfg.DefaultCommissionAccount
	pnlAccount := cfg.DefaultPnlAccount

	var err error
	for _, r := range cfg.Huobi.Rules {
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
		if r.TxType != nil {
			match = matchFunc(*r.TxType, o.TxTypeOriginal, sep, match)
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
		}
	}

	return ignore, "", "", map[ir.Account]string{
		ir.CashAccount:       cashAccount,
		ir.PositionAccount:   positionAccount,
		ir.CommissionAccount: commissionAccount,
		ir.PnlAccount:        pnlAccount,
	}, nil
}
