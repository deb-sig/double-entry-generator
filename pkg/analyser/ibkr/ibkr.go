package ibkr

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type Ibkr struct{}

func (i Ibkr) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniqMap := map[string]bool{
		cfg.DefaultCashAccount:       true,
		cfg.DefaultPositionAccount:   true,
		cfg.DefaultCommissionAccount: true,
		cfg.DefaultPnlAccount:        true,
		cfg.DefaultMinusAccount:      true,
		cfg.DefaultPlusAccount:       true,
	}
	if cfg.Ibkr == nil {
		return uniqMap
	}
	for _, rule := range cfg.Ibkr.Rules {
		if rule.CashAccount != nil {
			uniqMap[*rule.CashAccount] = true
		}
		if rule.PositionAccount != nil {
			uniqMap[*rule.PositionAccount] = true
		}
		if rule.CommissionAccount != nil {
			uniqMap[*rule.CommissionAccount] = true
		}
		if rule.PnlAccount != nil {
			uniqMap[*rule.PnlAccount] = true
		}
	}
	return uniqMap
}

func (i Ibkr) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false
	cashAccount := cfg.DefaultCashAccount
	positionAccount := cfg.DefaultPositionAccount
	commissionAccount := cfg.DefaultCommissionAccount
	pnlAccount := cfg.DefaultPnlAccount

	if cfg.Ibkr != nil {
		for _, rule := range cfg.Ibkr.Rules {
			match := true
			sep := ","
			if rule.Separator != nil {
				sep = *rule.Separator
			}
			matchFunc := util.SplitFindContains
			if rule.FullMatch {
				matchFunc = util.SplitFindEquals
			}
			if rule.Type != nil {
				match = matchFunc(*rule.Type, o.TypeOriginal, sep, match)
			}
			if rule.Item != nil {
				match = matchFunc(*rule.Item, o.Item, sep, match)
			}
			if !match {
				continue
			}
			if rule.Ignore {
				ignore = true
				break
			}
			if rule.CashAccount != nil {
				cashAccount = *rule.CashAccount
			}
			if rule.PositionAccount != nil {
				positionAccount = *rule.PositionAccount
			}
			if rule.CommissionAccount != nil {
				commissionAccount = *rule.CommissionAccount
			}
			if rule.PnlAccount != nil {
				pnlAccount = *rule.PnlAccount
			}
		}
	}

	extraAccounts := map[ir.Account]string{
		ir.CashAccount:       cashAccount,
		ir.PositionAccount:   positionAccount,
		ir.CommissionAccount: commissionAccount,
		ir.PnlAccount:        pnlAccount,
	}

	switch o.OrderType {
	case ir.OrderTypeSecuritiesTrade:
		return ignore, "", "", extraAccounts, nil
	case ir.OrderTypeCurrencyExchange:
		return ignore, cashAccount, cashAccount, extraAccounts, nil
	default:
		if o.TypeOriginal == "Deposits/Withdrawals" {
			if o.Type == ir.TypeRecv {
				return ignore, cfg.DefaultMinusAccount, cashAccount, nil, nil
			}
			return ignore, cashAccount, cfg.DefaultPlusAccount, nil, nil
		}
		if o.Type == ir.TypeRecv {
			return ignore, pnlAccount, cashAccount, nil, nil
		}
		return ignore, cashAccount, commissionAccount, nil, nil
	}
}
