package cib_debit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type CibDebit struct{}

func (CibDebit) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniq := make(map[string]bool)
	if cfg.DefaultPlusAccount != "" {
		uniq[cfg.DefaultPlusAccount] = true
	}
	if cfg.DefaultMinusAccount != "" {
		uniq[cfg.DefaultMinusAccount] = true
	}
	if cfg.DefaultCashAccount != "" {
		uniq[cfg.DefaultCashAccount] = true
	}
	if cfg.CibDebit != nil {
		for _, rule := range cfg.CibDebit.Rules {
			if rule.MethodAccount != nil {
				uniq[*rule.MethodAccount] = true
			}
			if rule.TargetAccount != nil {
				uniq[*rule.TargetAccount] = true
			}
		}
	}
	return uniq
}

func (CibDebit) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false
	if o.OrderType == ir.OrderTypeCurrencyExchange {
		cash := firstNonEmpty(cfg.DefaultCashAccount, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount)
		return ignore, cash, cash, nil, nil
	}

	minus, plus := defaultAccountsFor(o.Type, cfg)
	tags := make([]string, 0)

	if cfg.CibDebit == nil {
		return ignore, minus, plus, nil, tags
	}

	for _, rule := range cfg.CibDebit.Rules {
		match := true
		sep := ","
		if rule.Separator != nil {
			sep = *rule.Separator
		}
		matchFunc := util.SplitFindContains
		if rule.FullMatch {
			matchFunc = util.SplitFindEquals
		}
		if rule.Peer != nil {
			match = matchFunc(*rule.Peer, o.Peer, sep, match)
		}
		if rule.PeerBank != nil {
			match = matchMetadata(*rule.PeerBank, o.Metadata, "peerBank", sep, match, matchFunc)
		}
		if rule.PeerAccount != nil {
			match = matchMetadata(*rule.PeerAccount, o.Metadata, "peerAccount", sep, match, matchFunc)
		}
		if rule.Item != nil {
			match = matchFunc(*rule.Item, o.Item, sep, match)
		}
		if rule.Type != nil {
			match = matchFunc(*rule.Type, o.TypeOriginal, sep, match)
		}
		if rule.TxType != nil {
			match = matchFunc(*rule.TxType, o.TxTypeOriginal, sep, match)
		}
		if rule.MinPrice != nil {
			match = match && o.Money >= *rule.MinPrice
		}
		if rule.MaxPrice != nil {
			match = match && o.Money <= *rule.MaxPrice
		}
		if !match {
			continue
		}
		if rule.Ignore {
			ignore = true
			break
		}
		if rule.TargetAccount != nil {
			if o.Type == ir.TypeRecv {
				minus = *rule.TargetAccount
			} else {
				plus = *rule.TargetAccount
			}
		}
		if rule.MethodAccount != nil {
			if o.Type == ir.TypeRecv {
				plus = *rule.MethodAccount
			} else {
				minus = *rule.MethodAccount
			}
		}
		if rule.Tag != nil {
			tags = append(tags, strings.Split(*rule.Tag, sep)...)
		}
	}

	return ignore, minus, plus, nil, tags
}

func matchMetadata(pattern string, metadata map[string]string, key, sep string, match bool, matchFunc func(string, string, string, bool) bool) bool {
	if metadata == nil {
		return false
	}
	value, ok := metadata[key]
	if !ok {
		return false
	}
	return matchFunc(pattern, value, sep, match)
}

func defaultAccountsFor(t ir.Type, cfg *config.Config) (string, string) {
	cash := firstNonEmpty(cfg.DefaultCashAccount, cfg.DefaultPlusAccount, cfg.DefaultMinusAccount)
	minus := firstNonEmpty(cfg.DefaultMinusAccount, cash)
	plus := firstNonEmpty(cfg.DefaultPlusAccount, cash)

	switch t {
	case ir.TypeRecv:
		return minus, cash
	case ir.TypeSend:
		return cash, plus
	default:
		return minus, plus
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
