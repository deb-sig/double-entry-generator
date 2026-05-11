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

import "time"

// Order 单个代币交易记录
type Order struct {
	TxHash           string    // 交易哈希（原始值）
	BlockNo          string    // 区块号
	UnixTimestamp    int64     // Unix 时间戳
	DateTime         time.Time // 时间
	From             string    // 发送地址（小写，用于匹配）
	FromOriginal     string    // 发送地址（原始值，用于输出）
	To               string    // 接收地址（小写，用于匹配）
	ToOriginal       string    // 接收地址（原始值，用于输出）
	TokenValue       float64   // 代币数量
	ContractAddress  string    // 合约地址（小写，用于匹配）
	ContractAddressOriginal string // 合约地址（原始值，用于输出）
	TokenName        string    // 代币名称
	TokenSymbol      string    // 代币符号
	
	// 解析后填充
	Direction        string    // "recv" 或 "send"
	Peer             string    // 对方地址（原始值，用于输出）
	Tags             []string  // 标签
}

// Type 交易类型
type Type string

const (
	TypeRecv Type = "recv" // 接收
	TypeSend Type = "send" // 发送
)

