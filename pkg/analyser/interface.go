package analyser

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/pkg/analyser/icbc"

	"github.com/deb-sig/double-entry-generator/pkg/analyser/alipay"
	"github.com/deb-sig/double-entry-generator/pkg/analyser/htsec"
	"github.com/deb-sig/double-entry-generator/pkg/analyser/huobi"
	"github.com/deb-sig/double-entry-generator/pkg/analyser/wechat"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/consts"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// Interface is the interface of analyser.
type Interface interface {
	GetAllCandidateAccounts(cfg *config.Config) map[string]bool
	GetAccountsAndTags(o *ir.Order, cfg *config.Config, target, provider string) (bool, string, string, map[ir.Account]string, []string)
}

// New creates a new analyser.
func New(providerName string) (Interface, error) {
	switch providerName {
	case consts.ProviderAlipay:
		return alipay.Alipay{}, nil
	case consts.ProviderWechat:
		return wechat.Wechat{}, nil
	case consts.ProviderHuobi:
		return huobi.Huobi{}, nil
	case consts.ProviderHtsec:
		return htsec.Htsec{}, nil
	case consts.ProviderIcbc:
		return icbc.Icbc{}, nil
	default:
		return nil, fmt.Errorf("Fail to create the analyser for the given name %s", providerName)
	}
}
