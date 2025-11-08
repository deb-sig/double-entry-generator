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

package erc20

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

// ERC20 ERC-20 代币 provider
type ERC20 struct {
	Config              *Config
	DefaultMinusAccount string
	DefaultPlusAccount  string
	Orders              []Order
}

// New 创建新的 ERC20 provider
func New() *ERC20 {
	return &ERC20{
		Config: &Config{},
	}
}

// Translate 实现 Provider 接口
func (e *ERC20) Translate(filename string) (*ir.IR, error) {
	log.Printf("[ERC20] Reading CSV file: %s", filename)

	// 读取 CSV 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
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
	log.Printf("[ERC20] CSV headers: %v", headers)

	// 解析数据行
	for i, record := range records[1:] {
		if len(record) == 0 || (len(record) == 1 && record[0] == "") {
			continue // 跳过空行
		}

		order, err := e.parseRecord(headers, record)
		if err != nil {
			log.Printf("[ERC20] Warning: failed to parse row %d: %v", i+2, err)
			continue
		}

		e.Orders = append(e.Orders, order)
	}

	log.Printf("[ERC20] Parsed %d orders", len(e.Orders))

	// 转换为 IR
	return e.convertToIR()
}

// parseRecord 解析单行记录
func (e *ERC20) parseRecord(headers, record []string) (Order, error) {
	order := Order{}

	// 创建字段映射
	fieldMap := make(map[string]string)
	for i, header := range headers {
		if i < len(record) {
			fieldMap[header] = record[i]
		}
	}

	// 解析各个字段
	order.TxHash = fieldMap["Transaction Hash"]
	order.BlockNo = fieldMap["Blockno"]
	
	// 解析时间戳
	timestampStr := fieldMap["UnixTimestamp"]
	if timestampStr != "" {
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return order, fmt.Errorf("invalid timestamp: %w", err)
		}
		order.UnixTimestamp = timestamp
		order.DateTime = time.Unix(timestamp, 0)
	}

	// 地址统一转小写
	order.From = strings.ToLower(fieldMap["From"])
	order.To = strings.ToLower(fieldMap["To"])

	// 解析代币数量
	tokenValueStr := fieldMap["TokenValue"]
	if tokenValueStr != "" {
		// 移除千位分隔符逗号（如 "5,586" -> "5586"）
		tokenValueStr = strings.ReplaceAll(tokenValueStr, ",", "")
		value, err := strconv.ParseFloat(tokenValueStr, 64)
		if err != nil {
			return order, fmt.Errorf("invalid token value: %w", err)
		}
		order.TokenValue = value
	}

	order.USDValueDayOfTx = fieldMap["USDValueDayOfTx"]
	order.ContractAddress = strings.ToLower(fieldMap["ContractAddress"])
	order.TokenName = fieldMap["TokenName"]
	order.TokenSymbol = fieldMap["TokenSymbol"]

	// 判断方向
	if e.Config.MyAddress != "" {
		myAddr := strings.ToLower(e.Config.MyAddress)
		if order.To == myAddr {
			order.Direction = "recv"
			order.Peer = order.From
		} else if order.From == myAddr {
			order.Direction = "send"
			order.Peer = order.To
		}
	}

	return order, nil
}

// convertToIR 转换为 IR 格式
func (e *ERC20) convertToIR() (*ir.IR, error) {
	result := &ir.IR{
		Orders: make([]ir.Order, 0, len(e.Orders)),
	}

	for i, order := range e.Orders {
		// 应用规则匹配
		matchedRule := e.matchRule(&order)
		
		// 如果规则标记为 ignore，则跳过
		if matchedRule != nil && matchedRule.Ignore {
			log.Printf("[ERC20] Ignoring order %d (tx: %s) due to rule", i, order.TxHash[:10])
			continue
		}

		// 构建 IR Order
		irOrder := e.buildIROrder(&order, matchedRule)
		result.Orders = append(result.Orders, irOrder)
	}

	log.Printf("[ERC20] Converted %d orders to IR format", len(result.Orders))
	return result, nil
}

// matchRule 匹配规则
func (e *ERC20) matchRule(order *Order) *Rule {
	if e.Config == nil || len(e.Config.Rules) == 0 {
		return nil
	}

	for _, rule := range e.Config.Rules {
		if e.ruleMatches(order, &rule) {
			return &rule
		}
	}

	return nil
}

// ruleMatches 检查规则是否匹配
func (e *ERC20) ruleMatches(order *Order, rule *Rule) bool {
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
func (e *ERC20) matchTimeRange(t time.Time, timeRange string) bool {
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
func (e *ERC20) buildIROrder(order *Order, rule *Rule) ir.Order {
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
	e.setAccounts(&irOrder, order, rule)

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
func (e *ERC20) setAccounts(irOrder *ir.Order, order *Order, rule *Rule) {
	var assetAccount, targetAccount string
	
	// 获取资产账户
	if rule != nil && rule.MethodAccount != nil {
		assetAccount = *rule.MethodAccount
	} else {
		assetAccount = e.DefaultPlusAccount
	}
	
	// 获取目标账户
	if rule != nil && rule.TargetAccount != nil {
		targetAccount = *rule.TargetAccount
	} else {
		if order.Direction == "recv" {
			targetAccount = e.DefaultMinusAccount
		} else {
			targetAccount = e.DefaultPlusAccount
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
func (e *ERC20) applyRuleSettings(irOrder *ir.Order, rule *Rule) {
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


