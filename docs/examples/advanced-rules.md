---
title: 高级规则配置
layout: default
parent: 示例
nav_order: 2
description: "学习如何编写复杂的匹配规则和账户映射"
permalink: /examples/advanced-rules/
---

# 高级规则配置

本指南将展示如何使用 Double Entry Generator 的高级功能来创建复杂的匹配规则和账户映射。

## 复杂条件匹配

### 多字段组合条件

```yaml
rules:
  - name: "信用卡还款"
    conditions:
      - field: "description"
        contains: ["信用卡", "还款"]
      - field: "amount"
        greater_than: 0
      - field: "type"
        equals: "转账"
    target_account: "Assets:Bank:Checking"
    target_account_minus: "Liabilities:CreditCard"
    tags: ["credit-card", "payment"]
```

### 正则表达式匹配

```yaml
rules:
  - name: "股票交易"
    conditions:
      - field: "description"
        regex: ".*股票.*|.*证券.*|.*交易.*"
      - field: "amount"
        regex: "^-?[0-9]+(\.[0-9]+)?$"
    target_account: "Assets:Securities"
    tags: ["stock", "investment"]
```

### 金额范围匹配

```yaml
rules:
  - name: "大额消费"
    conditions:
      - field: "amount"
        greater_than: 1000
      - field: "description"
        not_contains: ["工资", "奖金", "转账"]
    target_account: "Expenses:Large"
    tags: ["large-expense"]
    note: "大额消费，需要特别关注"
```

## 动态账户映射

### 基于金额的账户选择

```yaml
rules:
  - name: "餐饮消费 - 小额"
    conditions:
      - field: "description"
        contains: ["美团", "饿了么", "餐厅", "外卖"]
      - field: "amount"
        less_than: 50
    target_account: "Expenses:Food:Small"
    tags: ["food", "small"]

  - name: "餐饮消费 - 大额"
    conditions:
      - field: "description"
        contains: ["美团", "饿了么", "餐厅", "外卖"]
      - field: "amount"
        greater_than_or_equal: 50
    target_account: "Expenses:Food:Large"
    tags: ["food", "large"]
```

### 基于时间的账户选择

```yaml
rules:
  - name: "工作日交通"
    conditions:
      - field: "description"
        contains: ["地铁", "公交", "出租车"]
      - field: "time"
        weekday: [1, 2, 3, 4, 5]  # 周一到周五
    target_account: "Expenses:Transport:Work"
    tags: ["transport", "work"]

  - name: "周末交通"
    conditions:
      - field: "description"
        contains: ["地铁", "公交", "出租车"]
      - field: "time"
        weekday: [6, 7]  # 周六和周日
    target_account: "Expenses:Transport:Personal"
    tags: ["transport", "personal"]
```

## 复杂账户映射

### 多级账户结构

```yaml
accounts:
  # 基础账户映射
  "支付宝": "Assets:Alipay"
  "微信": "Assets:WeChat"
  "建设银行": "Assets:Bank:CCB"
  "工商银行": "Assets:Bank:ICBC"
  
  # 商户账户映射
  "美团": "Expenses:Food:Delivery"
  "饿了么": "Expenses:Food:Delivery"
  "星巴克": "Expenses:Food:Coffee"
  "麦当劳": "Expenses:Food:FastFood"
  
  # 服务账户映射
  "中国移动": "Expenses:Utilities:Phone"
  "中国联通": "Expenses:Utilities:Phone"
  "国家电网": "Expenses:Utilities:Electricity"
  "自来水公司": "Expenses:Utilities:Water"
```

### 正则表达式账户映射

```yaml
accounts:
  # 使用正则表达式匹配
  ".*银行.*": "Assets:Bank:Other"
  ".*证券.*": "Assets:Securities"
  ".*基金.*": "Assets:Funds"
  ".*保险.*": "Expenses:Insurance"
```

## 标签和元数据

### 智能标签系统

```yaml
rules:
  - name: "网购消费"
    conditions:
      - field: "description"
        contains: ["淘宝", "京东", "天猫", "拼多多"]
    target_account: "Expenses:Shopping:Online"
    tags: ["shopping", "online"]
    metadata:
      category: "shopping"
      platform: "online"
      review_required: true

  - name: "实体店消费"
    conditions:
      - field: "description"
        contains: ["超市", "商场", "便利店"]
    target_account: "Expenses:Shopping:Offline"
    tags: ["shopping", "offline"]
    metadata:
      category: "shopping"
      platform: "offline"
      review_required: false
```

### 动态标签生成

```yaml
rules:
  - name: "月度订阅"
    conditions:
      - field: "description"
        contains: ["会员", "订阅", "VIP"]
    target_account: "Expenses:Subscriptions"
    tags: ["subscription", "monthly"]
    dynamic_tags:
      - field: "description"
        extract: "([A-Za-z]+).*会员"
        prefix: "service:"
```

## 错误处理和验证

### 数据验证规则

```yaml
validation:
  required_fields: ["date", "amount", "description"]
  amount_format: "decimal"
  date_format: "YYYY-MM-DD"
  
rules:
  - name: "数据验证 - 金额"
    conditions:
      - field: "amount"
        regex: "^-?[0-9]+(\.[0-9]{1,2})?$"
    action: "validate"
    error_message: "金额格式不正确"
    
  - name: "数据验证 - 日期"
    conditions:
      - field: "date"
        regex: "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
    action: "validate"
    error_message: "日期格式不正确"
```

### 异常处理

```yaml
rules:
  - name: "异常交易处理"
    conditions:
      - field: "amount"
        greater_than: 10000
    target_account: "Expenses:Unknown:Large"
    tags: ["exception", "review-required"]
    note: "大额异常交易，需要人工审核"
    requires_review: true
```

## 配置文件组织

### 模块化配置

```yaml
# config.yaml - 主配置文件
default:
  default_minus_account: "Assets:Bank:Checking"
  default_plus_account: "Expenses:Unknown"
  default_currency: "CNY"
  default_tags: ["imported"]

# 引入其他配置文件
includes:
  - rules/bank-rules.yaml
  - rules/payment-rules.yaml
  - rules/investment-rules.yaml
  - accounts/account-mapping.yaml
```

```yaml
# rules/bank-rules.yaml - 银行规则
rules:
  - name: "银行转账"
    conditions:
      - field: "description"
        contains: ["转账", "汇款"]
    target_account: "Assets:Bank:Checking"
    tags: ["transfer"]
```

```yaml
# accounts/account-mapping.yaml - 账户映射
accounts:
  "建设银行": "Assets:Bank:CCB"
  "工商银行": "Assets:Bank:ICBC"
  "支付宝": "Assets:Alipay"
  "微信": "Assets:WeChat"
```

## 最佳实践

### 1. 规则优先级

```yaml
# 规则按顺序匹配，越具体的规则应该放在前面
rules:
  - name: "星巴克消费"  # 最具体
    conditions:
      - field: "description"
        contains: ["星巴克"]
    target_account: "Expenses:Food:Coffee:Starbucks"
    
  - name: "咖啡消费"    # 较具体
    conditions:
      - field: "description"
        contains: ["咖啡", "coffee"]
    target_account: "Expenses:Food:Coffee"
    
  - name: "餐饮消费"    # 最通用
    conditions:
      - field: "description"
        contains: ["餐厅", "外卖"]
    target_account: "Expenses:Food"
```

### 2. 性能优化

```yaml
# 使用索引字段提高匹配性能
indexed_fields:
  - "description"
  - "amount"
  - "type"

# 避免过于复杂的正则表达式
rules:
  - name: "简单匹配"
    conditions:
      - field: "description"
        contains: ["关键词1", "关键词2"]  # 推荐
    # 而不是使用复杂的正则表达式
```

### 3. 测试和调试

```yaml
# 添加调试信息
debug: true
log_level: "info"

rules:
  - name: "调试规则"
    conditions:
      - field: "description"
        contains: ["测试"]
    target_account: "Expenses:Test"
    tags: ["debug", "test"]
    debug_info: "这是一个测试规则"
```

## 常见问题

### Q: 如何处理重复的规则？

A: 规则按顺序匹配，第一个匹配的规则会被使用。建议将更具体的规则放在前面。

### Q: 正则表达式性能如何优化？

A: 避免使用过于复杂的正则表达式，优先使用简单的字符串匹配。

### Q: 如何调试规则匹配问题？

A: 启用调试模式，查看详细的匹配日志。

### Q: 如何处理特殊字符？

A: 在正则表达式中正确转义特殊字符，或使用字符串匹配。
