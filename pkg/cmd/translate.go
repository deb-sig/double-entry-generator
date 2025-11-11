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

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/deb-sig/double-entry-generator/v2/pkg/cmd/validator"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/oklink"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/wechat"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/bmo"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/ccb"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/citic"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/htsec"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/hxsec"
	_ "github.com/deb-sig/double-entry-generator/v2/pkg/provider/icbc"
)

var (
	providerName               string
	targetName                 string
	appendMode                 bool
	output                     string
	ignoreInvalidWechatTxTypes bool
)

var translateCmd = &cobra.Command{
	Use:   "translate [flags] <path to bill file>",
	Short: "Translate the bills to a given format",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		return validator.TranslateArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		run(args)
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	translateCmd.Flags().StringVarP(&providerName, "provider", "p", "alipay", "Bills provider (alipay)")
	translateCmd.Flags().StringVarP(&targetName, "target", "t", "beancount", "Target (beancount)")
	translateCmd.Flags().BoolVarP(&appendMode, "append", "a", false, "Append mode")
	translateCmd.Flags().StringVarP(&output, "output", "o", "default_output.beancount", "Output file")
	translateCmd.Flags().BoolVar(&ignoreInvalidWechatTxTypes, "ignore-invalid-tx-types", false, "Ignore invalid transaction types (ONLY support WeChat provider)")
}

func run(args []string) {
	// Get the config from viper.
	log.Printf("Use config file: %s", viper.ConfigFileUsed())
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read config file error: %v", err)
	}

	c := &config.Config{}
	err := viper.Unmarshal(c)
	logErrorIfNotNil(err)

	switch providerName {
	case consts.ProviderAlipay:
		fallthrough
	case consts.ProviderWechat:
		fallthrough
	case consts.ProviderCCB:
		if c.DefaultCurrency == "" ||
			c.DefaultMinusAccount == "" ||
			c.DefaultPlusAccount == "" {
			log.Fatalf("Failed to get default options in config")
		}
	case consts.ProviderHuobi:
		if c.DefaultCurrency == "" ||
			c.DefaultCashAccount == "" ||
			c.DefaultPositionAccount == "" ||
			c.DefaultCommissionAccount == "" ||
			c.DefaultPnlAccount == "" {
			log.Fatalf("Failed to get default options in config")
		}
	}

	p, err := provider.New(providerName)
	logErrorIfNotNil(err)

	if providerName == consts.ProviderWechat {
		if w, ok := p.(*wechat.Wechat); ok {
			w.IgnoreInvalidTxTypes = ignoreInvalidWechatTxTypes
		}
	}

	// Pass config to OKLink provider
	if providerName == consts.ProviderOKLink {
		if e, ok := p.(*oklink.OKLink); ok {
			e.Config = c.OKLink
			e.DefaultMinusAccount = c.DefaultMinusAccount
			e.DefaultPlusAccount = c.DefaultPlusAccount
			
			// 处理多地址配置：从 viper 中提取地址作为 key 的配置
			if c.OKLink != nil && c.OKLink.Addresses == nil {
				// 从 viper 中获取 oklink 的所有设置
				oklinkSettings := viper.GetStringMap("oklink")
				if len(oklinkSettings) > 0 {
					addresses := make(map[string]*oklink.AddressConfig)
					
					for key := range oklinkSettings {
						// 检查是否是地址格式（0x 开头或 T 开头）
						isAddress := (len(key) >= 2 && key[0:2] == "0x") || (len(key) >= 1 && key[0] == 'T')
						if isAddress {
							// 解析地址配置（使用 viper 的子配置）
							subViper := viper.Sub("oklink." + key)
							if subViper != nil {
								var addrConfig oklink.AddressConfig
								if err := subViper.Unmarshal(&addrConfig); err == nil {
									addresses[key] = &addrConfig
								}
							}
						}
					}
					
					if len(addresses) > 0 {
						c.OKLink.Addresses = addresses
					}
				}
			}
		}
	}

	i, err := p.Translate(args[0])
	logErrorIfNotNil(err)

	cpl, err := compiler.New(providerName, targetName, output, appendMode, c, i)
	logErrorIfNotNil(err)
	err = cpl.Compile()
	logErrorIfNotNil(err)
}

func logErrorIfNotNil(err error) {
	if err != nil {
		log.Fatalf("Failed to translate: %v", err)
	}
}
