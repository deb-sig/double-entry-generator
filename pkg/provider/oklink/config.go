/*
Copyright © 2024

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

package oklink

// Config OKLink 多链配置
// 使用地址作为 key，每个地址可以有独立的配置
// 示例:
//    oklink:
//      "0x...":  # Ethereum 地址
//        defaultCashAccount: "Assets:Crypto:Ethereum"
//        rules: [...]
//      "T...":   # TRON 地址
//        defaultCashAccount: "Assets:Crypto:TRON"
//        rules: [...]
// 
// 单地址时也使用相同格式（只有一个地址作为 key）
type Config struct {
	// 地址 -> 地址配置的映射（地址作为 key）
	Addresses map[string]*AddressConfig `yaml:",inline" mapstructure:",remain"`
}

// AddressConfig 单个地址的配置
type AddressConfig struct {
	// 规则列表
	Rules []Rule `yaml:"rules,omitempty"`
}

// Rule 匹配规则（参考支付宝配置风格）
type Rule struct {
	// ===== 匹配条件 =====
	
	// 代币相关
	TokenSymbol      *string  `yaml:"tokenSymbol,omitempty"`      // 代币符号，如 "USDT", "USDC"
	TokenName        *string  `yaml:"tokenName,omitempty"`        // 代币名称，如 "Tether USD"
	ContractAddress  *string  `yaml:"contractAddress,omitempty"`  // 合约地址
	
	// 地址相关（直接对应 CSV 中的 from/to）
	From             *string  `yaml:"from,omitempty"`             // 发送地址（CSV From 列）
	To               *string  `yaml:"to,omitempty"`               // 接收地址（CSV To 列）
	Peer             *string  `yaml:"peer,omitempty"`             // 对方地址（收款时匹配 from，付款时匹配 to）
	
	// 方向相关
	Direction        *string  `yaml:"direction,omitempty"`        // 交易方向：recv（收款）或 send（发送）
	
	// 金额相关
	MinAmount        *float64 `yaml:"minAmount,omitempty"`        // 最小金额
	MaxAmount        *float64 `yaml:"maxAmount,omitempty"`        // 最大金额
	
	// 时间相关
	Time             *string  `yaml:"time,omitempty"`             // 时间范围，如 "2024-01-01~2024-12-31"
	TimestampRange   *string  `yaml:"timestamp_range,omitempty"`  // Unix 时间戳范围
	
	// 交易哈希
	TxHash           *string  `yaml:"txHash,omitempty"`           // 交易哈希（精确匹配）
	
	// 区块相关
	MinBlockNo       *int64   `yaml:"minBlockNo,omitempty"`       // 最小区块号
	MaxBlockNo       *int64   `yaml:"maxBlockNo,omitempty"`       // 最大区块号
	
	// 匹配选项
	FullMatch        bool     `yaml:"fullMatch,omitempty"`        // 是否完全匹配（字符串字段）
	Separator        *string  `yaml:"sep,omitempty"`              // 标签分隔符，默认 ","
	
	// ===== 应用的账户和标签 =====
	// 参考支付宝配置：targetAccount(收入/支出账户) + methodAccount(资产账户)
	
	// 主要账户配置
	TargetAccount    *string  `yaml:"targetAccount,omitempty"`    // 目标账户（收入 Income:xxx 或支出 Expenses:xxx）
	MethodAccount    *string  `yaml:"methodAccount,omitempty"`    // 资产账户（代币账户 Assets:Crypto:xxx）
	
	// 货币单位
	Currency         *string  `yaml:"currency,omitempty"`         // 自定义货币单位（默认使用 tokenSymbol）
	
	// 标签
	Tags             *string  `yaml:"tags,omitempty"`             // 标签，多个用分隔符分隔
	
	// 忽略标记
	Ignore           bool     `yaml:"ignore,omitempty"`           // 是否忽略此交易，默认 false
	
	// 备注
	Note             *string  `yaml:"note,omitempty"`             // 自定义备注
}

