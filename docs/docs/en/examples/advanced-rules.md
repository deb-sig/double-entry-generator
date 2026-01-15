---
title: Advanced Rules Configuration
description: Learn how to write complex matching rules and account mappings
---


# Advanced Rules Configuration

This guide demonstrates how to use advanced features of Double Entry Generator to create complex matching rules and account mappings.

## Complex Condition Matching

### Multi-Field Combined Conditions

```yaml
rules:
  - name: "Credit Card Payment"
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

### Regular Expression Matching

```yaml
rules:
  - name: "Stock Trading"
    conditions:
      - field: "description"
        regex: ".*股票.*|.*证券.*|.*交易.*"
      - field: "amount"
        regex: "^-?[0-9]+(\.[0-9]+)?$"
    target_account: "Assets:Securities"
    tags: ["stock", "investment"]
```

### Amount Range Matching

```yaml
rules:
  - name: "Large Expense"
    conditions:
      - field: "amount"
        greater_than: 1000
      - field: "description"
        not_contains: ["工资", "奖金", "转账"]
    target_account: "Expenses:Large"
    tags: ["large-expense"]
    note: "Large expense, requires special attention"
```

## Dynamic Account Mapping

### Amount-Based Account Selection

```yaml
rules:
  - name: "Food Expense - Small"
    conditions:
      - field: "description"
        contains: ["美团", "饿了么", "餐厅", "外卖"]
      - field: "amount"
        less_than: 50
    target_account: "Expenses:Food:Small"
    tags: ["food", "small"]

  - name: "Food Expense - Large"
    conditions:
      - field: "description"
        contains: ["美团", "饿了么", "餐厅", "外卖"]
      - field: "amount"
        greater_than_or_equal: 50
    target_account: "Expenses:Food:Large"
    tags: ["food", "large"]
```

### Time-Based Account Selection

```yaml
rules:
  - name: "Weekday Transportation"
    conditions:
      - field: "description"
        contains: ["地铁", "公交", "出租车"]
      - field: "time"
        weekday: [1, 2, 3, 4, 5]  # Monday to Friday
    target_account: "Expenses:Transport:Work"
    tags: ["transport", "work"]

  - name: "Weekend Transportation"
    conditions:
      - field: "description"
        contains: ["地铁", "公交", "出租车"]
      - field: "time"
        weekday: [6, 7]  # Saturday and Sunday
    target_account: "Expenses:Transport:Personal"
    tags: ["transport", "personal"]
```

## Complex Account Mapping

### Multi-Level Account Structure

```yaml
accounts:
  # Basic account mapping
  "支付宝": "Assets:Alipay"
  "微信": "Assets:WeChat"
  "建设银行": "Assets:Bank:CCB"
  "工商银行": "Assets:Bank:ICBC"
  
  # Merchant account mapping
  "美团": "Expenses:Food:Delivery"
  "饿了么": "Expenses:Food:Delivery"
  "星巴克": "Expenses:Food:Coffee"
  "麦当劳": "Expenses:Food:FastFood"
  
  # Service account mapping
  "中国移动": "Expenses:Utilities:Phone"
  "中国联通": "Expenses:Utilities:Phone"
  "国家电网": "Expenses:Utilities:Electricity"
  "自来水公司": "Expenses:Utilities:Water"
```

### Regular Expression Account Mapping

```yaml
accounts:
  # Use regular expression matching
  ".*银行.*": "Assets:Bank:Other"
  ".*证券.*": "Assets:Securities"
  ".*基金.*": "Assets:Funds"
  ".*保险.*": "Expenses:Insurance"
```

## Tags and Metadata

### Smart Tag System

```yaml
rules:
  - name: "Online Shopping"
    conditions:
      - field: "description"
        contains: ["淘宝", "京东", "天猫", "拼多多"]
    target_account: "Expenses:Shopping:Online"
    tags: ["shopping", "online"]
    metadata:
      category: "shopping"
      platform: "online"
      review_required: true

  - name: "Offline Shopping"
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

### Dynamic Tag Generation

```yaml
rules:
  - name: "Monthly Subscription"
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

## Error Handling and Validation

### Data Validation Rules

```yaml
validation:
  required_fields: ["date", "amount", "description"]
  amount_format: "decimal"
  date_format: "YYYY-MM-DD"
  
rules:
  - name: "Data Validation - Amount"
    conditions:
      - field: "amount"
        regex: "^-?[0-9]+(\.[0-9]{1,2})?$"
    action: "validate"
    error_message: "Amount format is incorrect"
    
  - name: "Data Validation - Date"
    conditions:
      - field: "date"
        regex: "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
    action: "validate"
    error_message: "Date format is incorrect"
```

### Exception Handling

```yaml
rules:
  - name: "Exception Transaction Handling"
    conditions:
      - field: "amount"
        greater_than: 10000
    target_account: "Expenses:Unknown:Large"
    tags: ["exception", "review-required"]
    note: "Large exception transaction, requires manual review"
    requires_review: true
```

## Configuration File Organization

### Modular Configuration

```yaml
# config.yaml - Main configuration file
default:
  default_minus_account: "Assets:Bank:Checking"
  default_plus_account: "Expenses:Unknown"
  default_currency: "CNY"
  default_tags: ["imported"]

# Include other configuration files
includes:
  - rules/bank-rules.yaml
  - rules/payment-rules.yaml
  - rules/investment-rules.yaml
  - accounts/account-mapping.yaml
```

```yaml
# rules/bank-rules.yaml - Bank rules
rules:
  - name: "Bank Transfer"
    conditions:
      - field: "description"
        contains: ["转账", "汇款"]
    target_account: "Assets:Bank:Checking"
    tags: ["transfer"]
```

```yaml
# accounts/account-mapping.yaml - Account mapping
accounts:
  "建设银行": "Assets:Bank:CCB"
  "工商银行": "Assets:Bank:ICBC"
  "支付宝": "Assets:Alipay"
  "微信": "Assets:WeChat"
```

## Best Practices

### 1. Rule Priority

```yaml
# Rules are matched in order, more specific rules should be placed first
rules:
  - name: "Starbucks Expense"  # Most specific
    conditions:
      - field: "description"
        contains: ["星巴克"]
    target_account: "Expenses:Food:Coffee:Starbucks"
    
  - name: "Coffee Expense"    # More specific
    conditions:
      - field: "description"
        contains: ["咖啡", "coffee"]
    target_account: "Expenses:Food:Coffee"
    
  - name: "Food Expense"    # Most general
    conditions:
      - field: "description"
        contains: ["餐厅", "外卖"]
    target_account: "Expenses:Food"
```

### 2. Performance Optimization

```yaml
# Use indexed fields to improve matching performance
indexed_fields:
  - "description"
  - "amount"
  - "type"

# Avoid overly complex regular expressions
rules:
  - name: "Simple Matching"
    conditions:
      - field: "description"
        contains: ["关键词1", "关键词2"]  # Recommended
    # Instead of using complex regular expressions
```

### 3. Testing and Debugging

```yaml
# Add debugging information
debug: true
log_level: "info"

rules:
  - name: "Debug Rule"
    conditions:
      - field: "description"
        contains: ["测试"]
    target_account: "Expenses:Test"
    tags: ["debug", "test"]
    debug_info: "This is a test rule"
```

## Common Questions

### Q: How to handle duplicate rules?

A: Rules are matched in order, and the first matching rule will be used. It's recommended to place more specific rules first.

### Q: How to optimize regular expression performance?

A: Avoid using overly complex regular expressions, prioritize simple string matching.

### Q: How to debug rule matching issues?

A: Enable debug mode to view detailed matching logs.

### Q: How to handle special characters?

A: Properly escape special characters in regular expressions, or use string matching.
