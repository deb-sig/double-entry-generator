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
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/deb-sig/double-entry-generator/pkg/compiler"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/consts"
	"github.com/deb-sig/double-entry-generator/pkg/provider"
	"github.com/deb-sig/double-entry-generator/pkg/util"
)

var (
	providerName string
	targetName   string
	appendMode   bool
	output       string
)

var translateCmd = &cobra.Command{
	Use:   "translate [flags] <path to bill file>",
	Short: "Translate the bills to a given format",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Failed to translate: Require the bill file")
		} else if len(args) > 1 {
			// TODO(gaocegege): support it.
			return fmt.Errorf("Failed to translate: Do not support multi-file now")
		}

		if util.IsWasm() {
			return nil
		}

		_, err := os.Stat(args[0])
		if err == nil {
			return nil
		}
		return fmt.Errorf("Failed to translate: %v", err)
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
}

func run(args []string) {
	// Get the config from viper.
	c := &config.Config{}
	err := viper.Unmarshal(c)
	logErrorIfNotNil(err)

	switch providerName {
	case consts.ProviderAlipay:
		fallthrough
	case consts.ProviderWechat:
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

	fmt.Println("p.Translate")
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
