package spdb_debit

import (
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
	
	// 应用规则匹配
	if cfg.SpdbDebit != nil {
		for _, rule := range cfg.SpdbDebit.Rules {
			match := true
			sep := ","
			
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
			
			// 如果匹配，应用规则
			if match {
				// 设置目标账户
				if rule.TargetAccount != nil {
					counterpartyAccount = *rule.TargetAccount
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
	// 在Beancount模板中：
	// - PlusAccount 总是正数金额
	// - MinusAccount 总是负数金额
	
	// 收入交易（TypeRecv）：资金流入，资产账户增加
	// 应该是：资产账户（PlusAccount）增加（正数），对方账户（MinusAccount）减少（负数）
	if order.Type == ir.TypeRecv {
		plus = assetAccount      // 资产账户增加（正数）
		minus = counterpartyAccount // 对方账户减少（负数）
	} else {
		// 支出交易（TypeSend）：资金流出，资产账户减少
		// 应该是：对方账户（PlusAccount）增加（正数），资产账户（MinusAccount）减少（负数）
		plus = counterpartyAccount // 对方账户增加（正数）
		minus = assetAccount      // 资产账户减少（负数）
	}

	return ignore, minus, plus, nil, nil
}
