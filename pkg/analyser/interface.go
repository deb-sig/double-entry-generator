package analyser

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/alipay"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/bmo"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/bocomcredit"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/ccb"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/citic"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/cmb"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/hsbchk"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/htsec"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/huobi"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/hxsec"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/icbc"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/jd"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/mt"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/td"
	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/wechat"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
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
	case consts.ProviderTd:
		return td.Td{}, nil
	case consts.ProviderBmo:
		return bmo.Bmo{}, nil
	case consts.ProviderBocomCredit:
		return bocomcredit.BocomCredit{}, nil
	case consts.ProviderJD:
		return jd.JD{}, nil
	case consts.ProviderCitic:
		return citic.Citic{}, nil
	case consts.ProviderHsbcHK:
		return hsbchk.HsbcHK{}, nil
	case consts.ProviderMT:
		return mt.MT{}, nil
	case consts.ProviderHxsec:
		return hxsec.Hxsec{}, nil
	case consts.ProviderCCB:
		return ccb.CCB{}, nil
	case consts.ProviderCmb:
		return cmb.Cmb{}, nil

	default:
		return nil, fmt.Errorf("Fail to create the analyser for the given name %s", providerName)
	}
}
