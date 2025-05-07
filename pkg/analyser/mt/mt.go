package mt

import (
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

type MT struct {
}

// GetAllCandidateAccounts returns all accounts defined in config.
func (mt MT) GetAllCandidateAccounts(cfg *config.Config) map[string]bool {
	uniqMap := make(map[string]bool)

	if cfg.MT == nil || len(cfg.MT.Rules) == 0 {
		return uniqMap
	}

	for _, r := range cfg.MT.Rules {
		if r.MethodAccount != nil {
			uniqMap[*r.MethodAccount] = true
		}
		if r.TargetAccount != nil {
			uniqMap[*r.TargetAccount] = true
		}
	}
	uniqMap[cfg.DefaultPlusAccount] = true
	uniqMap[cfg.DefaultMinusAccount] = true
	return uniqMap
}

// GetAccounts matches rules, returns minus and plus account for beancount or ledger.
func (mt MT) GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string) {
	ignore := false

	if cfg.MT == nil || len(cfg.MT.Rules) == 0 {
		return ignore, cfg.DefaultMinusAccount, cfg.DefaultPlusAccount, nil, nil
	}
	resMinus := cfg.DefaultMinusAccount
	resPlus := cfg.DefaultPlusAccount
	var extraAccounts map[ir.Account]string
	var tags = make([]string, 0)

	var err error
	for _, r := range cfg.MT.Rules {
		match := true
		// get separator
		sep := ","
		if r.Separator != nil {
			sep = *r.Separator
		}

		matchFunc := util.SplitFindContains // 匹配规则为包含
		if r.FullMatch {
			matchFunc = util.SplitFindEquals // 匹配规则为完全相等
		}
		if r.Type != nil {
			match = matchFunc(*r.Type, o.TypeOriginal, sep, match)
		}
		if r.Item != nil {
			match = matchFunc(*r.Item, o.Item, sep, match)
		}
		if r.Method != nil {
			match = matchFunc(*r.Method, o.Method, sep, match)
		}
		if r.Time != nil { // 检查支付时间是否在给定的时间范围中
			match, err = util.SplitFindTimeInterval(*r.Time, o.PayTime, match)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
		if r.TimestampRange != nil { // 检查支付时间是否在 unix 时间戳格式的时间范围中
			match, err = util.SplitFindTimeStampInterval(*r.TimestampRange, o.PayTime, match)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
		if r.MinPrice != nil && o.Money < *r.MinPrice {
			match = false
		}
		if r.MaxPrice != nil && o.Money > *r.MaxPrice {
			match = false
		}

		if match {
			if r.Ignore {
				ignore = true
				break
			}
			// 美团目前看来，支付情况比较简单，只有消费与退款收入两种，暂不考虑美团月付的情况
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
			if r.Tags != nil {
				tags = strings.Split(*r.Tags, sep)
			}
		}
	}
	return ignore, resMinus, resPlus, extraAccounts, tags
}
