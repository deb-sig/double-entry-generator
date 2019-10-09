package wechat

import (
	"strings"

	"github.com/gaocegege/double-entry-generator/pkg/config"
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

type Wechat struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (w Wechat) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	// uniqMap will be used to create the concepts.
	uniqMap := make(map[string]bool)

	if cfg.Wechat == nil || len(cfg.Wechat.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.Wechat.Rules {
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
func (w Wechat) GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string) {
	if cfg.Wechat == nil || len(cfg.Wechat.Rules) == 0 {
		return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount
	}

	for _, r := range cfg.Wechat.Rules {
		match := true
		if r.Peer != nil {
			if !strings.Contains(o.Peer, *r.Peer) {
				match = false
			}
		}
		if r.Type != nil {
			if !strings.Contains(o.TypeOriginal, *r.Type) {
				match = false
			}
		}
		if r.Method != nil {
			if !strings.Contains(o.Method, *r.Method) {
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
			return resMinus, resPlus
		}
	}
	return cfg.DefaultMinusAccount, cfg.DefaultPlusAccount
}
