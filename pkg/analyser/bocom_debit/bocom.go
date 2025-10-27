package bocom_debit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type Bocom struct{}

// GetAllCandidateAccounts returns all accounts defined in config rules.
func (b Bocom) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniqMap := make(map[string]bool)

	if cfg.BocomDebit == nil || len(cfg.BocomDebit.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.BocomDebit.Rules {
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
	if cfg.DefaultPlusAccount != "" {
		uniqMap[cfg.DefaultPlusAccount] = true
	}
	if cfg.DefaultMinusAccount != "" {
		uniqMap[cfg.DefaultMinusAccount] = true
	}
	if cfg.DefaultCashAccount != "" {
		uniqMap[cfg.DefaultCashAccount] = true
	}
	return uniqMap
}

// GetAccountsAndTags determines the accounts and tags for an order.
func (b Bocom) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.BocomDebit == nil || len(cfg.BocomDebit.Rules) == 0 {
		minus, plus := defaultAccountsFor(o.Type, cfg)
		return ignore, minus, plus, nil, nil
	}

	resMinus, resPlus := defaultAccountsFor(o.Type, cfg)
	var tags []string

	for _, r := range cfg.BocomDebit.Rules {
		match := true
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
		if r.TxType != nil {
			match = matchFunc(*r.TxType, o.TxTypeOriginal, sep, match)
		}
		if r.Status != nil {
			status, exists := o.Metadata["status"]
			if !exists {
				match = false
			} else {
				match = matchFunc(*r.Status, status, sep, match)
			}
		}

		if r.MinPrice != nil {
			match = match && o.Money >= *r.MinPrice
		}
		if r.MaxPrice != nil {
			match = match && o.Money <= *r.MaxPrice
		}

		if !match {
			continue
		}

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

		if r.Tag != nil {
			tags = append(tags, strings.Split(*r.Tag, sep)...)
		}
	}

	return ignore, resMinus, resPlus, nil, tags
}

func defaultAccountsFor(t ir.Type, cfg *config.Config) (string, string) {
	cash := cfg.DefaultCashAccount
	minus := cfg.DefaultMinusAccount
	plus := cfg.DefaultPlusAccount

	if cash == "" {
		cash = plus
	}
	if cash == "" {
		cash = minus
	}
	if minus == "" {
		minus = cash
	}
	if plus == "" {
		plus = cash
	}

	switch t {
	case ir.TypeRecv:
		return minus, cash
	case ir.TypeSend:
		return cash, plus
	default:
		return minus, plus
	}
}
