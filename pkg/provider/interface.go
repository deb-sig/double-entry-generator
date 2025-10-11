/*
Copyright © 2019 Ce Gao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/bmo"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/ccb"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/citic"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/hsbchk"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/icbc"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/jd"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/mt"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/td"

	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/alipay"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/htsec"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/huobi"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/hxsec"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/wechat"
)

// Interface is the interface for the provider.
type Interface interface {
	Translate(filename string) (*ir.IR, error)
}

// supportedProviders 是所有支持的 provider 列表（按字母排序）
var supportedProviders = []string{
	consts.ProviderAlipay,
	consts.ProviderBmo,
	consts.ProviderCCB,
	consts.ProviderCitic,
	consts.ProviderHsbcHK,
	consts.ProviderHtsec,
	consts.ProviderHuobi,
	consts.ProviderHxsec,
	consts.ProviderIcbc,
	consts.ProviderJD,
	consts.ProviderMT,
	consts.ProviderTd,
	consts.ProviderWechat,
}

// GetSupportedProviders 返回所有支持的 provider 列表
func GetSupportedProviders() []string {
	// 返回副本，避免外部修改
	result := make([]string, len(supportedProviders))
	copy(result, supportedProviders)
	return result
}

// New creates a new interface.
func New(name string) (Interface, error) {
	switch name {
	case consts.ProviderAlipay:
		return alipay.New(), nil
	case consts.ProviderWechat:
		return wechat.New(), nil
	case consts.ProviderHuobi:
		return huobi.New(), nil
	case consts.ProviderHtsec:
		return htsec.New(), nil
	case consts.ProviderIcbc:
		return icbc.New(), nil
	case consts.ProviderTd:
		return td.New(), nil
	case consts.ProviderBmo:
		return bmo.New(), nil
	case consts.ProviderJD:
		return jd.New(), nil
	case consts.ProviderCitic:
		return citic.New(), nil
	case consts.ProviderHsbcHK:
		return hsbchk.New(), nil
	case consts.ProviderMT:
		return mt.New(), nil
	case consts.ProviderHxsec:
		return hxsec.New(), nil
	case consts.ProviderCCB:
		return ccb.New(), nil
	default:
		return nil, fmt.Errorf("Fail to create the provider for the given name %s", name)
	}
}
