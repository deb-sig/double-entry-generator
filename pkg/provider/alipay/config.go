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

package alipay

// Config is the configuration for Alipay.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer           *string `mapstructure:"peer,omitempty"`
	Item           *string `mapstructure:"item,omitempty"`
	Category       *string `mapstructure:"category,omitempty"`
	Type           *string `mapstructure:"type,omitempty"`
	Method         *string `mapstructure:"method,omitempty"`
	Separator      *string `mapstructure:"sep,omitempty"` // default: ,
	Time           *string `mapstructure:"time,omitempty"`
	TimestampRange *string `mapstructure:"timestamp_range,omitempty"`
	MethodAccount  *string `mapstructure:"methodAccount,omitempty"`
	TargetAccount  *string `mapstructure:"targetAccount,omitempty"`
	PnlAccount     *string `mapstructure:"pnlAccount,omitempty"`
	FullMatch      bool    `mapstructure:"fullMatch,omitempty"`
	Tags           *string `mapstructure:"tags,omitempty"`
	Ignore         bool    `mapstructure:"ignore,omitempty"` // default: false
}
