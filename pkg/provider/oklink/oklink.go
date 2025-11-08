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

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// OKLink OKLink 多链代币 provider (支持 ERC20, TRC20, BSC 等)
type OKLink struct {
	Config              *Config
	DefaultMinusAccount string
	DefaultPlusAccount  string
	Orders              []Order
}

// New 创建新的 OKLink provider
func New() *OKLink {
	return &OKLink{
		Config: &Config{},
	}
}

// Translate 实现 Provider 接口
func (e *OKLink) Translate(filename string) (*ir.IR, error) {
	log.Printf("[OKLink] Reading CSV file: %s", filename)

	// 检查配置
	if e.Config == nil {
		return nil, fmt.Errorf("配置缺失：请在 config.yaml 中添加 oklink 配置段\n" +
			"示例:\n" +
			"oklink:\n" +
			"  \"0x你的地址\":  # 地址作为 key\n" +
			"    defaultCashAccount: \"Assets:Crypto:Ethereum\"\n" +
			"    rules:\n" +
			"      - tokenSymbol: \"USDT\"\n" +
			"        methodAccount: \"Assets:Crypto:Ethereum:USDT\"\n" +
			"        targetAccount: \"Income:Crypto:Transfer\"\n" +
			"\n" +
			"多地址示例:\n" +
			"oklink:\n" +
			"  \"0x...\":  # Ethereum 地址\n" +
			"    defaultCashAccount: \"Assets:Crypto:Ethereum\"\n" +
			"    rules: [...]\n" +
			"  \"T...\":  # TRON 地址\n" +
			"    defaultCashAccount: \"Assets:Crypto:TRON\"\n" +
			"    rules: [...]")
	}

	// 检查是否有地址配置
	if e.Config.Addresses == nil || len(e.Config.Addresses) == 0 {
		return nil, fmt.Errorf("配置错误：必须配置至少一个地址\n" +
			"使用地址作为 key，例如:\n" +
			"oklink:\n" +
			"  \"0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5\":\n" +
			"    rules: [...]\n" +
			"  \"TFGqVkQCdHxMEZd7Ys6MbvTh8MwPuB7Lkh\":\n" +
			"    rules: [...]")
	}

	// 读取 CSV 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true        // 允许不规范的引号
	reader.TrimLeadingSpace = true  // 去除前导空格
	reader.FieldsPerRecord = -1     // 允许字段数不一致
	
	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// 第一行是表头
	headers := records[0]
	log.Printf("[OKLink] CSV headers: %v", headers)

	// 解析数据行
	for i, record := range records[1:] {
		if len(record) == 0 || (len(record) == 1 && record[0] == "") {
			continue // 跳过空行
		}

		order, err := e.parseRecord(headers, record)
		if err != nil {
			log.Printf("[OKLink] Warning: failed to parse row %d: %v", i+2, err)
			continue
		}

		e.Orders = append(e.Orders, order)
	}

	log.Printf("[OKLink] Parsed %d orders", len(e.Orders))

	// 转换为 IR
	return e.convertToIR()
}

// parseRecord 解析单行记录
func (e *OKLink) parseRecord(headers, record []string) (Order, error) {
	order := Order{}

	// 创建字段映射
	fieldMap := make(map[string]string)
	for i, header := range headers {
		if i < len(record) {
			fieldMap[header] = record[i]
		}
	}

	// 解析各个字段（支持中文和英文表头）
	// 交易哈希
	order.TxHash = fieldMap["交易哈希"]
	if order.TxHash == "" {
		order.TxHash = fieldMap["Tx Hash"]
	}
	
	// 区块高度
	order.BlockNo = fieldMap["区块高度"]
	if order.BlockNo == "" {
		order.BlockNo = fieldMap["blockHeight"]
	}
	
	// 解析 UTC 时间（支持两种格式）
	utcTimeStr := fieldMap["UTC时间"]
	if utcTimeStr == "" {
		utcTimeStr = fieldMap["blockTime(UTC)"]
	}
	if utcTimeStr != "" {
		// OKLink 格式: "2025/11/07 15:14:23" 或 "2025-11-07 15:14:23"
		t, err := time.Parse("2006/01/02 15:04:05", utcTimeStr)
		if err != nil {
			// 尝试另一种格式
			t, err = time.Parse("2006-01-02 15:04:05", utcTimeStr)
			if err != nil {
				return order, fmt.Errorf("invalid UTC time: %w", err)
			}
		}
		order.DateTime = t
		order.UnixTimestamp = t.Unix()
	}

	// 地址统一转小写（支持中英文表头）
	order.From = strings.ToLower(fieldMap["发送方"])
	if order.From == "" {
		order.From = strings.ToLower(fieldMap["from"])
	}
	
	order.To = strings.ToLower(fieldMap["接收方"])
	if order.To == "" {
		order.To = strings.ToLower(fieldMap["to"])
	}

	// 解析代币数量（支持"数量"和"value"）
	tokenValueStr := fieldMap["数量"]
	if tokenValueStr == "" {
		tokenValueStr = fieldMap["value"]
	}
	if tokenValueStr != "" {
		// 移除千位分隔符逗号（如 "5,586" -> "5586"）
		tokenValueStr = strings.ReplaceAll(tokenValueStr, ",", "")
		value, err := strconv.ParseFloat(tokenValueStr, 64)
		if err != nil {
			return order, fmt.Errorf("invalid token value: %w", err)
		}
		order.TokenValue = value
	}

	// 合约地址和代币符号（支持中英文表头）
	order.ContractAddress = strings.ToLower(fieldMap["代币地址"])
	if order.ContractAddress == "" {
		order.ContractAddress = strings.ToLower(fieldMap["tokenAddress"])
	}
	
	order.TokenSymbol = fieldMap["代币符号"]
	if order.TokenSymbol == "" {
		order.TokenSymbol = fieldMap["symbol"]
	}
	
	// OKLink 没有单独的 TokenName 字段，使用 TokenSymbol
	order.TokenName = order.TokenSymbol

	// 判断方向：优先匹配 from（send），再匹配 to（recv）
	if e.Config != nil && e.Config.Addresses != nil {
		fromLower := strings.ToLower(order.From)
		toLower := strings.ToLower(order.To)
		
		// 优先检查 from 地址（send 视角）
		for addr := range e.Config.Addresses {
			addrLower := strings.ToLower(addr)
			if fromLower == addrLower {
				order.Direction = "send"
				order.Peer = order.To
				return order, nil
			}
		}
		
		// 如果 from 不在配置中，检查 to 地址（recv 视角）
		for addr := range e.Config.Addresses {
			addrLower := strings.ToLower(addr)
			if toLower == addrLower {
				order.Direction = "recv"
				order.Peer = order.From
				return order, nil
			}
		}
	}

	return order, nil
}

// convertToIR 转换为 IR 格式
func (e *OKLink) convertToIR() (*ir.IR, error) {
	result := &ir.IR{
		Orders: make([]ir.Order, 0, len(e.Orders)),
	}

	for i, order := range e.Orders {
		// 应用规则匹配（返回规则和地址配置）
		matchedRule, addrConfig := e.matchRule(&order)
		
		// 如果地址配置不存在，跳过（说明这个交易不属于配置中的任何地址）
		if addrConfig == nil {
			log.Printf("[OKLink] Skipping order %d (tx: %s): address not in config", i, order.TxHash[:10])
			continue
		}
		
		// 如果规则标记为 ignore，则跳过
		if matchedRule != nil && matchedRule.Ignore {
			log.Printf("[OKLink] Ignoring order %d (tx: %s) due to rule", i, order.TxHash[:10])
			continue
		}

		// 构建 IR Order（传入规则和地址配置）
		irOrder := e.buildIROrder(&order, matchedRule, addrConfig)
		result.Orders = append(result.Orders, irOrder)
	}

	log.Printf("[OKLink] Converted %d orders to IR format", len(result.Orders))
	return result, nil
}

// getAddressConfig 根据交易地址获取对应的配置
// 逻辑：
// 1. 优先匹配 from 地址（如果是 send，使用 from 的规则）
// 2. 如果 from 不在配置中，匹配 to 地址（如果是 recv，使用 to 的规则）
// 3. 如果两个地址都在配置中，优先使用 from 地址的规则（send 视角）
func (e *OKLink) getAddressConfig(order *Order) (*AddressConfig, string) {
	if e.Config == nil || e.Config.Addresses == nil {
		return nil, ""
	}
	
	fromLower := strings.ToLower(order.From)
	toLower := strings.ToLower(order.To)
	
	// 优先匹配 from 地址（send 视角）
	for addr, addrConfig := range e.Config.Addresses {
		addrLower := strings.ToLower(addr)
		if fromLower == addrLower {
			return addrConfig, addr
		}
	}
	
	// 如果 from 不在配置中，匹配 to 地址（recv 视角）
	for addr, addrConfig := range e.Config.Addresses {
		addrLower := strings.ToLower(addr)
		if toLower == addrLower {
			return addrConfig, addr
		}
	}
	
	return nil, ""
}

// matchRule 匹配规则
func (e *OKLink) matchRule(order *Order) (*Rule, *AddressConfig) {
	// 先获取地址配置
	addrConfig, _ := e.getAddressConfig(order)
	if addrConfig == nil {
		return nil, nil
	}
	
	// 在地址配置的规则中匹配
	if len(addrConfig.Rules) == 0 {
		return nil, addrConfig
	}

	for _, rule := range addrConfig.Rules {
		if e.ruleMatches(order, &rule) {
			return &rule, addrConfig
		}
	}

	return nil, addrConfig
}

// ruleMatches 检查规则是否匹配
func (e *OKLink) ruleMatches(order *Order, rule *Rule) bool {
	// TokenSymbol 匹配
	if rule.TokenSymbol != nil {
		if rule.FullMatch {
			if order.TokenSymbol != *rule.TokenSymbol {
				return false
			}
		} else {
			if !strings.Contains(order.TokenSymbol, *rule.TokenSymbol) {
				return false
			}
		}
	}

	// TokenName 匹配
	if rule.TokenName != nil {
		if rule.FullMatch {
			if order.TokenName != *rule.TokenName {
				return false
			}
		} else {
			if !strings.Contains(order.TokenName, *rule.TokenName) {
				return false
			}
		}
	}

	// ContractAddress 匹配（地址统一小写）
	if rule.ContractAddress != nil {
		ruleAddr := strings.ToLower(*rule.ContractAddress)
		if order.ContractAddress != ruleAddr {
			return false
		}
	}

	// From 地址匹配
	if rule.From != nil {
		ruleFrom := strings.ToLower(*rule.From)
		if order.From != ruleFrom {
			return false
		}
	}

	// To 地址匹配
	if rule.To != nil {
		ruleTo := strings.ToLower(*rule.To)
		if order.To != ruleTo {
			return false
		}
	}

	// Peer 地址匹配
	if rule.Peer != nil {
		rulePeer := strings.ToLower(*rule.Peer)
		if order.Peer != rulePeer {
			return false
		}
	}

	// Direction 方向匹配
	if rule.Direction != nil {
		if order.Direction != *rule.Direction {
			return false
		}
	}

	// 金额范围匹配
	if rule.MinAmount != nil && order.TokenValue < *rule.MinAmount {
		return false
	}
	if rule.MaxAmount != nil && order.TokenValue > *rule.MaxAmount {
		return false
	}

	// TxHash 精确匹配
	if rule.TxHash != nil {
		if order.TxHash != *rule.TxHash {
			return false
		}
	}

	// 区块号范围匹配
	if rule.MinBlockNo != nil || rule.MaxBlockNo != nil {
		blockNo, err := strconv.ParseInt(order.BlockNo, 10, 64)
		if err == nil {
			if rule.MinBlockNo != nil && blockNo < *rule.MinBlockNo {
				return false
			}
			if rule.MaxBlockNo != nil && blockNo > *rule.MaxBlockNo {
				return false
			}
		}
	}

	// 时间范围匹配
	if rule.Time != nil {
		if !e.matchTimeRange(order.DateTime, *rule.Time) {
			return false
		}
	}

	// 所有条件都匹配
	return true
}

// matchTimeRange 匹配时间范围
func (e *OKLink) matchTimeRange(t time.Time, timeRange string) bool {
	// 格式：2024-01-01~2024-12-31
	parts := strings.Split(timeRange, "~")
	if len(parts) != 2 {
		return true // 格式错误，默认匹配
	}

	start, err1 := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	end, err2 := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
	
	if err1 != nil || err2 != nil {
		return true // 解析错误，默认匹配
	}

	// 结束时间设置为当天的最后一秒
	end = end.Add(24*time.Hour - time.Second)

	return t.After(start) && t.Before(end) || t.Equal(start) || t.Equal(end)
}

// buildIROrder 构建 IR Order
func (e *OKLink) buildIROrder(order *Order, rule *Rule, addrConfig *AddressConfig) ir.Order {
	irOrder := ir.Order{
		OrderType: ir.OrderTypeCrypto, // 使用加密货币模板（高精度）
		PayTime:   order.DateTime,
		Peer:      order.Peer,
		Money:     order.TokenValue,
		Currency:  order.TokenSymbol, // 使用代币符号作为货币单位
	}

	// 构建描述
	direction := "Transfer"
	if order.Direction == "recv" {
		direction = "Receive"
	} else if order.Direction == "send" {
		direction = "Send"
	}
	irOrder.Item = fmt.Sprintf("%s %s", order.TokenSymbol, direction)

	// 设置账户（根据规则或使用默认）
	e.setAccounts(&irOrder, order, rule, addrConfig)

	// 应用规则中的其他配置
	if rule != nil {
		e.applyRuleSettings(&irOrder, rule)
	}

	// 添加元数据
	irOrder.Metadata = map[string]string{
		"txHash":          order.TxHash,
		"blockNo":         order.BlockNo,
		"from":            order.From,
		"to":              order.To,
		"contractAddress": order.ContractAddress,
		"tokenName":       order.TokenName,
		"tokenSymbol":     order.TokenSymbol,
		"direction":       order.Direction,
	}

	return irOrder
}

// setAccounts 设置交易的借贷账户
func (e *OKLink) setAccounts(irOrder *ir.Order, order *Order, rule *Rule, addrConfig *AddressConfig) {
	var assetAccount, targetAccount string
	
	// 获取资产账户
	if rule != nil && rule.MethodAccount != nil {
		assetAccount = *rule.MethodAccount
	} else {
		// 如果没有规则指定，使用全局默认账户
		assetAccount = e.DefaultPlusAccount
	}
	
	// 获取目标账户
	if rule != nil && rule.TargetAccount != nil {
		targetAccount = *rule.TargetAccount
	} else {
		// 如果没有规则指定，根据方向自动推断合适的账户类型
		if order.Direction == "recv" {
			// 收款：使用收入账户（Income）
			// 如果全局默认账户是资产账户，则使用默认的收入账户
			if strings.HasPrefix(e.DefaultMinusAccount, "Income:") {
				targetAccount = e.DefaultMinusAccount
			} else {
				targetAccount = "Income:Crypto:Transfer"
			}
		} else {
			// 发送：使用支出账户（Expenses）
			// 如果全局默认账户是支出账户，则使用默认的支出账户
			if strings.HasPrefix(e.DefaultPlusAccount, "Expenses:") {
				targetAccount = e.DefaultPlusAccount
			} else {
				targetAccount = "Expenses:Crypto:Transfer"
			}
		}
	}
	
	// 根据方向分配账户
	if order.Direction == "recv" {
		irOrder.PlusAccount = assetAccount
		irOrder.MinusAccount = targetAccount
	} else {
		irOrder.MinusAccount = assetAccount
		irOrder.PlusAccount = targetAccount
	}
}

// applyRuleSettings 应用规则中的标签、备注、货币单位等配置
func (e *OKLink) applyRuleSettings(irOrder *ir.Order, rule *Rule) {
	// 应用标签
	if rule.Tags != nil {
		sep := ","
		if rule.Separator != nil {
			sep = *rule.Separator
		}
		tags := strings.Split(*rule.Tags, sep)
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				irOrder.Tags = append(irOrder.Tags, tag)
			}
		}
	}
	
	// 应用自定义备注
	if rule.Note != nil {
		irOrder.Item = *rule.Note
	}
	
	// 应用自定义货币单位
	if rule.Currency != nil {
		irOrder.Currency = *rule.Currency
	}
}


