package spdb_debit

import (
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/util"
)

type SpdbDebit struct{}

// GetAllCandidateAccounts returns all accounts defined in config.
func (SpdbDebit) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniq := make(map[string]bool)

	// Add default accounts
	uniq[cfg.DefaultPlusAccount] = true
	uniq[cfg.DefaultMinusAccount] = true
	uniq[cfg.DefaultCashAccount] = true

	// Add accounts from rules
	if cfg.SpdbDebit != nil {
		for _, rule := range cfg.SpdbDebit.Rules {
			if rule.TargetAccount != nil {
				uniq[*rule.TargetAccount] = true
			}
		}
	}

	return uniq
}

// GetAccountsAndTags returns accounts determined by rules.
func (SpdbDebit) GetAccountsAndTags(order *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	// 默认账户设置
	assetAccount := cfg.DefaultMinusAccount
	counterpartyAccount := cfg.DefaultPlusAccount

	// 初始化 tags
	var tags = make([]string, 0)

	// 应用规则匹配
	if cfg.SpdbDebit != nil {
		for _, rule := range cfg.SpdbDebit.Rules {
			match := true
			sep := ","
			if rule.Separator != nil {
				sep = *rule.Separator
			}

			// 使用默认的匹配函数
			matchFunc := util.SplitFindContains
			if rule.FullMatch {
				matchFunc = util.SplitFindEquals
			}

			// 匹配对方账户
			if rule.Peer != nil {
				match = matchFunc(*rule.Peer, order.Peer, sep, match)
			}

			// 匹配交易摘要
			if rule.Item != nil {
				match = matchFunc(*rule.Item, order.Item, sep, match)
			}

			// 匹配交易类型（TxType）
			if rule.TxType != nil {
				match = matchFunc(*rule.TxType, order.TxTypeOriginal, sep, match)
			}

			// 如果匹配，应用规则
			if match {
				// 设置目标账户
				if rule.TargetAccount != nil {
					counterpartyAccount = *rule.TargetAccount
				}

				// 设置 tags
				if rule.Tags != nil {
					tags = strings.Split(*rule.Tags, sep)
				}

				// 设置忽略标志
				if rule.Ignore {
					ignore = true
					break
				}
			}
		}
	}

	var minus, plus string

	// 根据交易类型确定正确的账户方向
	if order.Type == ir.TypeRecv {
		plus = assetAccount
		minus = counterpartyAccount
	} else {
		plus = counterpartyAccount
		minus = assetAccount
	}

	return ignore, minus, plus, nil, tags
}