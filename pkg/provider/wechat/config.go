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

package wechat

// Config is the configuration for Alipay.
type Config struct {
	Rules []Rule `yaml:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer         *string `yaml:"peer,omitempty"`
	Item         *string `yaml:"item,omitempty"`
	Type         *string `yaml:"type,omitempty"`
	Method       *string `yaml:"method,omitempty"`
	StartTime    *string `yaml:"startTime,omitempty"`
	EndTime      *string `yaml:"endTime,omitempty"`
	MinusAccount *string `yaml:"minusAccount,omitempty"`
	PlusAccount  *string `yaml:"plusAccount,omitempty"`
}
