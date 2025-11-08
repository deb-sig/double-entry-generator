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

	// 地址保存原始值和小写值（支持中英文表头）
	fromOriginal := fieldMap["发送方"]
	if fromOriginal == "" {
		fromOriginal = fieldMap["from"]
	}
	order.FromOriginal = fromOriginal
	order.From = strings.ToLower(fromOriginal)
	
	toOriginal := fieldMap["接收方"]
	if toOriginal == "" {
		toOriginal = fieldMap["to"]
	}
	order.ToOriginal = toOriginal
	order.To = strings.ToLower(toOriginal)

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

	// 合约地址保存原始值和小写值（支持中英文表头）
	contractAddrOriginal := fieldMap["代币地址"]
	if contractAddrOriginal == "" {
		contractAddrOriginal = fieldMap["tokenAddress"]
	}
	order.ContractAddressOriginal = contractAddrOriginal
	order.ContractAddress = strings.ToLower(contractAddrOriginal)
	
	order.TokenSymbol = fieldMap["代币符号"]
	if order.TokenSymbol == "" {
		order.TokenSymbol = fieldMap["symbol"]
	}
	
	// OKLink 没有单独的 TokenName 字段，使用 TokenSymbol
	order.TokenName = order.TokenSymbol

	// 判断方向：优先匹配 from（send），再匹配 to（recv）
	if e.Config != nil && e.Config.Addresses != nil {
		// order.From 和 order.To 已经是小写的了
		
		// 优先检查 from 地址（send 视角）
		for addr := range e.Config.Addresses {
			addrLower := strings.ToLower(addr)
			if order.From == addrLower {
				order.Direction = "send"
				order.Peer = order.ToOriginal // 使用原始值
				return order, nil
			}
		}
		
		// 如果 from 不在配置中，检查 to 地址（recv 视角）
		for addr := range e.Config.Addresses {
			addrLower := strings.ToLower(addr)
			if order.To == addrLower {
				order.Direction = "recv"
				order.Peer = order.FromOriginal // 使用原始值
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
		// 检查两个地址是否都在配置中（自己的账户互转）
		fromConfig, fromAddr := e.getAddressConfigByAddr(&order, order.From)
		toConfig, toAddr := e.getAddressConfigByAddr(&order, order.To)
		
		// 如果两个地址都在配置中，按资产转移处理（资产账户 A → 资产账户 B）
		if fromConfig != nil && toConfig != nil {
			// 匹配两个地址的规则
			fromRules := e.matchAllRules(&order, fromConfig)
			toRules := e.matchAllRules(&order, toConfig)
			
			// 检查是否有规则标记为 ignore
			shouldIgnore := false
			for _, rule := range append(fromRules, toRules...) {
				if rule.Ignore {
					log.Printf("[OKLink] Ignoring order %d (tx: %s) due to rule", i, order.TxHash[:10])
					shouldIgnore = true
					break
				}
			}
			
			if shouldIgnore {
				continue
			}
			
			// 构建资产转移交易（资产账户 A → 资产账户 B）
			irOrder := e.buildTransferOrder(&order, fromRules, toRules, fromConfig, toConfig, fromAddr, toAddr)
			result.Orders = append(result.Orders, irOrder)
			continue
		}
		
		// 如果只有一个地址在配置中，按正常 send/recv 处理
		addrConfig, _ := e.getAddressConfig(&order)
		
		// 如果地址配置不存在，跳过（说明这个交易不属于配置中的任何地址）
		if addrConfig == nil {
			log.Printf("[OKLink] Skipping order %d (tx: %s): address not in config", i, order.TxHash[:10])
			continue
		}
		
		// 匹配所有规则（支持多个规则匹配，累加设置）
		matchedRules := e.matchAllRules(&order, addrConfig)
		
		// 检查是否有规则标记为 ignore
		shouldIgnore := false
		for _, rule := range matchedRules {
			if rule.Ignore {
				log.Printf("[OKLink] Ignoring order %d (tx: %s) due to rule", i, order.TxHash[:10])
				shouldIgnore = true
				break
			}
		}
		
		if shouldIgnore {
			continue
		}

		// 构建 IR Order（传入所有匹配的规则和地址配置）
		irOrder := e.buildIROrder(&order, matchedRules, addrConfig)
		result.Orders = append(result.Orders, irOrder)
	}

	log.Printf("[OKLink] Converted %d orders to IR format", len(result.Orders))
	return result, nil
}

// getAddressConfig 根据交易地址获取对应的配置
// 逻辑：
// 1. 优先匹配 from 地址（如果是 send，使用 from 的规则）
// 2. 如果 from 不在配置中，匹配 to 地址（如果是 recv，使用 to 的规则）
func (e *OKLink) getAddressConfig(order *Order) (*AddressConfig, string) {
	// 优先匹配 from 地址
	if config, addr := e.getAddressConfigByAddr(order, order.From); config != nil {
		return config, addr
	}
	
	// 如果 from 不在配置中，匹配 to 地址
	return e.getAddressConfigByAddr(order, order.To)
}

// getAddressConfigByAddr 根据指定地址获取对应的配置
func (e *OKLink) getAddressConfigByAddr(order *Order, address string) (*AddressConfig, string) {
	if e.Config == nil || e.Config.Addresses == nil {
		return nil, ""
	}
	
	addrLower := strings.ToLower(address)
	for addr, addrConfig := range e.Config.Addresses {
		if strings.ToLower(addr) == addrLower {
			return addrConfig, addr
		}
	}
	
	return nil, ""
}

// matchAllRules 匹配所有规则（支持多个规则匹配，累加设置）
// 参考 alipay provider 的实现方式
func (e *OKLink) matchAllRules(order *Order, addrConfig *AddressConfig) []*Rule {
	if addrConfig == nil || len(addrConfig.Rules) == 0 {
		return nil
	}

	var matchedRules []*Rule
	for i := range addrConfig.Rules {
		rule := &addrConfig.Rules[i]
		if e.ruleMatches(order, rule) {
			matchedRules = append(matchedRules, rule)
		}
	}

	return matchedRules
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

	// ContractAddress 匹配（大小写不敏感）
	if rule.ContractAddress != nil {
		ruleAddr := strings.ToLower(*rule.ContractAddress)
		if order.ContractAddress != ruleAddr {
			return false
		}
	}

	// From 地址匹配（大小写不敏感）
	if rule.From != nil {
		ruleFrom := strings.ToLower(*rule.From)
		if order.From != ruleFrom {
			return false
		}
	}

	// To 地址匹配（大小写不敏感）
	if rule.To != nil {
		ruleTo := strings.ToLower(*rule.To)
		if order.To != ruleTo {
			return false
		}
	}

	// Peer 地址匹配（大小写不敏感，但需要比较原始值的小写版本）
	if rule.Peer != nil {
		rulePeer := strings.ToLower(*rule.Peer)
		// Peer 可能是 FromOriginal 或 ToOriginal，需要比较小写版本
		peerLower := strings.ToLower(order.Peer)
		if peerLower != rulePeer {
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
// 支持多个规则匹配，累加设置（参考 alipay provider 的实现方式）
func (e *OKLink) buildIROrder(order *Order, matchedRules []*Rule, addrConfig *AddressConfig) ir.Order {
	irOrder := ir.Order{
		OrderType: ir.OrderTypeCrypto, // 使用加密货币模板（高精度）
		PayTime:   order.DateTime,
		Peer:      order.Peer,
		Money:     order.TokenValue,
		Currency:  order.TokenSymbol, // 使用代币符号作为货币单位
		Tags:      make([]string, 0),  // 初始化 tags 切片
	}

	// 构建描述
	direction := "Transfer"
	if order.Direction == "recv" {
		direction = "Receive"
	} else if order.Direction == "send" {
		direction = "Send"
	}
	irOrder.Item = fmt.Sprintf("%s %s", order.TokenSymbol, direction)

	// 设置账户（支持多个规则匹配，累加设置）
	e.setAccountsFromRules(&irOrder, order, matchedRules, addrConfig)

	// 应用规则中的其他配置（tags、note、currency）
	for _, rule := range matchedRules {
		e.applyRuleSettings(&irOrder, rule)
	}

	// 添加元数据（使用原始值）
	irOrder.Metadata = map[string]string{
		"txHash":          order.TxHash,                    // 原始值
		"blockNo":         order.BlockNo,
		"from":            order.FromOriginal,               // 原始值
		"to":              order.ToOriginal,                 // 原始值
		"contractAddress": order.ContractAddressOriginal,   // 原始值
		"tokenName":       order.TokenName,
		"tokenSymbol":     order.TokenSymbol,
		"direction":       order.Direction,
	}

	return irOrder
}

// buildTransferOrder 构建资产转移交易（两个地址都在配置中）
// 资产账户 A（from）→ 资产账户 B（to）
func (e *OKLink) buildTransferOrder(order *Order, fromRules []*Rule, toRules []*Rule, fromConfig *AddressConfig, toConfig *AddressConfig, fromAddr string, toAddr string) ir.Order {
	irOrder := ir.Order{
		OrderType: ir.OrderTypeCrypto, // 使用加密货币模板（高精度）
		PayTime:   order.DateTime,
		Peer:      order.ToOriginal, // 使用 to 地址作为 peer
		Money:     order.TokenValue,
		Currency:  order.TokenSymbol, // 使用代币符号作为货币单位
		Tags:      make([]string, 0),  // 初始化 tags 切片
	}

	// 构建描述
	irOrder.Item = fmt.Sprintf("%s Transfer", order.TokenSymbol)

	// 设置账户：资产账户 A（from）→ 资产账户 B（to）
	fromAccount := e.getMethodAccountFromRules(order, fromRules, fromConfig)
	toAccount := e.getMethodAccountFromRules(order, toRules, toConfig)
	
	// 如果找不到，使用默认账户
	if fromAccount == "" {
		fromAccount = e.DefaultPlusAccount
	}
	if toAccount == "" {
		toAccount = e.DefaultPlusAccount
	}
	
	// 资产转移：from 账户减少，to 账户增加
	irOrder.MinusAccount = fromAccount
	irOrder.PlusAccount = toAccount

	// 应用规则中的其他配置（tags、note、currency）
	// 合并两个地址的规则
	allRules := append(fromRules, toRules...)
	for _, rule := range allRules {
		e.applyRuleSettings(&irOrder, rule)
	}

	// 添加元数据（使用原始值）
	irOrder.Metadata = map[string]string{
		"txHash":          order.TxHash,                    // 原始值
		"blockNo":         order.BlockNo,
		"from":            order.FromOriginal,               // 原始值
		"to":              order.ToOriginal,                 // 原始值
		"contractAddress": order.ContractAddressOriginal,   // 原始值
		"tokenName":       order.TokenName,
		"tokenSymbol":     order.TokenSymbol,
		"direction":       "transfer", // 标记为资产转移
	}

	return irOrder
}

// getMethodAccountFromRules 从规则中获取 methodAccount
func (e *OKLink) getMethodAccountFromRules(order *Order, matchedRules []*Rule, addrConfig *AddressConfig) string {
	// 优先从匹配的规则中查找
	for _, rule := range matchedRules {
		if rule.MethodAccount != nil {
			return *rule.MethodAccount
		}
	}
	
	// 如果规则中没有，从地址配置的其他规则中查找
	return e.findMethodAccountFromRules(order, addrConfig)
}

// setAccountsFromRules 设置交易的借贷账户（支持多个规则匹配，累加设置）
// 参考 alipay provider 的实现方式：Support multiple matches, like one rule matches the
// minus account, the other rule matches the plus account.
func (e *OKLink) setAccountsFromRules(irOrder *ir.Order, order *Order, matchedRules []*Rule, addrConfig *AddressConfig) {
	var assetAccount, targetAccount string
	
	// 初始化默认账户
	if order.Direction == "recv" {
		assetAccount = e.DefaultPlusAccount
		targetAccount = e.DefaultMinusAccount
	} else {
		assetAccount = e.DefaultPlusAccount
		targetAccount = e.DefaultPlusAccount
	}
	
	// 遍历所有匹配的规则，累加设置账户
	// 参考 alipay: Support multiple matches, like one rule matches the minus account, the other rule matches the plus account.
	for _, rule := range matchedRules {
		// 如果规则有 MethodAccount，设置资产账户
		if rule.MethodAccount != nil {
			assetAccount = *rule.MethodAccount
		}
		
		// 如果规则有 TargetAccount，设置目标账户
		if rule.TargetAccount != nil {
			targetAccount = *rule.TargetAccount
		}
	}
	
	// 如果所有规则都没有 MethodAccount，尝试从地址配置的其他规则中查找
	if assetAccount == e.DefaultPlusAccount {
		assetAccount = e.findMethodAccountFromRules(order, addrConfig)
		if assetAccount == "" {
			assetAccount = e.DefaultPlusAccount
		}
	}
	
	// 如果所有规则都没有 TargetAccount，根据方向自动推断
	if targetAccount == e.DefaultPlusAccount || targetAccount == e.DefaultMinusAccount {
		if order.Direction == "recv" {
			// 收款：使用收入账户（Income）
			if strings.HasPrefix(e.DefaultMinusAccount, "Income:") {
				targetAccount = e.DefaultMinusAccount
			} else {
				targetAccount = "Income:Crypto:Transfer"
			}
		} else {
			// 发送：使用支出账户（Expenses）
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

// findMethodAccountFromRules 从地址配置的其他规则中查找 methodAccount
// 优先查找匹配相同 tokenSymbol 的规则
func (e *OKLink) findMethodAccountFromRules(order *Order, addrConfig *AddressConfig) string {
	if addrConfig == nil || len(addrConfig.Rules) == 0 {
		return ""
	}
	
	// 优先查找匹配相同 tokenSymbol 的规则
	for _, r := range addrConfig.Rules {
		if r.MethodAccount != nil {
			// 如果规则匹配 tokenSymbol，使用它的 methodAccount
			if r.TokenSymbol != nil && *r.TokenSymbol == order.TokenSymbol {
				return *r.MethodAccount
			}
		}
	}
	
	// 如果没找到匹配 tokenSymbol 的，查找第一个有 methodAccount 的规则
	for _, r := range addrConfig.Rules {
		if r.MethodAccount != nil {
			return *r.MethodAccount
		}
	}
	
	return ""
}

// applyRuleSettings 应用规则中的标签、备注、货币单位等配置
func (e *OKLink) applyRuleSettings(irOrder *ir.Order, rule *Rule) {
	// 应用标签（去重）
	if rule.Tags != nil {
		sep := ","
		if rule.Separator != nil {
			sep = *rule.Separator
		}
		tags := strings.Split(*rule.Tags, sep)
		
		// 使用 map 去重
		tagMap := make(map[string]bool)
		for _, tag := range irOrder.Tags {
			tagMap[tag] = true
		}
		
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" && !tagMap[tag] {
				irOrder.Tags = append(irOrder.Tags, tag)
				tagMap[tag] = true
			}
		}
	}
	
	// 应用自定义备注（后面的规则覆盖前面的）
	if rule.Note != nil {
		irOrder.Item = *rule.Note
	}
	
	// 应用自定义货币单位（后面的规则覆盖前面的）
	if rule.Currency != nil {
		irOrder.Currency = *rule.Currency
	}
}


