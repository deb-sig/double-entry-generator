package analyser

import (
	"fmt"

	"github.com/gaocegege/double-entry-generator/pkg/analyser/alipay"
	"github.com/gaocegege/double-entry-generator/pkg/analyser/huobi"
	"github.com/gaocegege/double-entry-generator/pkg/analyser/wechat"
	"github.com/gaocegege/double-entry-generator/pkg/config"
	"github.com/gaocegege/double-entry-generator/pkg/consts"
	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

// Interface is the interface of analyser.
type Interface interface {
	GetAllCandidateAccounts(cfg *config.Config) map[string]bool
	GetAccounts(o *ir.Order, cfg *config.Config, target, provider string) (string, string, map[string]string)
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
	default:
		return nil, fmt.Errorf("Fail to create the analyser for the given name %s", providerName)
	}
}
