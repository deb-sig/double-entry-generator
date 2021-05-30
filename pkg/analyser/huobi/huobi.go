package huobi

import (
	"strings"

	"github.com/gaocegege/double-entry-generator/pkg/config"
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

type Huobi struct {
}

func (h Huobi) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// TODO(TripleZ)
	return nil
}

func (h Huobi) GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string, map[ir.Account]string) {
	if cfg.Huobi == nil || len(cfg.Huobi.Rules) == 0 {
		return "", "", map[ir.Account]string{
			ir.CashAccount:       cfg.DefaultCashAccount,
			ir.PositionAccount:   cfg.DefaultPositionAccount,
			ir.CommissionAccount: cfg.DefaultCommissionAccount,
			ir.PnlAccount:        cfg.DefaultPnlAccount,
		}
	}

	cashAccount := cfg.DefaultCashAccount
	positionAccount := cfg.DefaultPositionAccount
	commissionAccount := cfg.DefaultCommissionAccount
	pnlAccount := cfg.DefaultPnlAccount

	for _, r := range cfg.Huobi.Rules {
		match := true
		// get seperator
		sep := ","
		if r.Seperator != nil {
			sep = *r.Seperator
		}
		if r.Type != nil {
			match = SplitFindContains(*r.Type, o.TypeOriginal, sep, match)
		}
		if r.TxType != nil {
			match = SplitFindContains(*r.TxType, o.TxTypeOriginal, sep, match)
		}
		if r.Item != nil {
			match = SplitFindContains(*r.Item, o.Item, sep, match)
		}

		if match {
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

	return "", "", map[ir.Account]string{
		ir.CashAccount:       cashAccount,
		ir.PositionAccount:   positionAccount,
		ir.CommissionAccount: commissionAccount,
		ir.PnlAccount:        pnlAccount,
	}
}

func SplitFindContains(str, target, sep string, match bool) bool {
	ss := strings.Split(str, sep)
	isContain := false
	for _, s := range ss {
		if strings.Contains(target, s) {
			isContain = true
			break
		}
	}
	if !isContain {
		return false
	}
	return match
}
