package alipay

import (
	"strings"

	"github.com/gaocegege/double-entry-generator/pkg/config"
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

type Alipay struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (a Alipay) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Alipay == nil || len(cfg.Alipay.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Alipay.Rules {
		if r.MinusAccount != nil {
			uniqMap[*r.MinusAccount] = true
		}
		if r.PlusAccount != nil {
			uniqMap[*r.PlusAccount] = true
		}
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccounts returns minus and plus account.
func (a Alipay) GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string) {

	if cfg.Alipay == nil || len(cfg.Alipay.Rules) == 0 {
		return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount
	}

	for _, r := range cfg.Alipay.Rules {
		match := true
		if r.Peer != nil {
			if !strings.Contains(o.Peer, *r.Peer) {
				match = false
			}
		}
		if r.Item != nil {
			if !strings.Contains(o.Item, *r.Item) {
				match = false
			}
		}
		if r.StartTime != nil && r.EndTime != nil {
			// TODO(gaocegege): Support it.
		}
		if match {
			resMinus := cfg.DefaultMinusAccount
			resPlus := cfg.DefaultPlusAccount
			if r.MinusAccount != nil {
				resMinus = *r.MinusAccount
			}
			if r.PlusAccount != nil {
				resPlus = *r.PlusAccount
			}

			if strings.HasPrefix(o.Item, "退款-") {
				return resPlus, resMinus
			}
			return resMinus, resPlus
		}
	}

	if strings.HasPrefix(o.Item, "退款-") {
		return cfg.DefaultPlusAccount, cfg.DefaultMinusAccount
	}
	return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount
}
