---
title: 配置总览
description: 配置与规则编写说明
---

# 配置总览

Double Entry Generator 使用 YAML 格式的配置文件定义转换规则。

## 基本结构

```yaml
# 全局配置
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:Example
defaultCurrency: CNY
title: 示例配置
layout: default

# 各 provider 配置
providerName:
  rules:
    - # 规则 1
    - # 规则 2
```

## 全局配置字段

### 必填字段

- `defaultMinusAccount`: 金额减少时的默认账户
- `defaultPlusAccount`: 金额增加时的默认账户
- `defaultCurrency`: 默认货币代码（如 CNY、USD、EUR）
- `title`: 配置文件标题

### 可选字段

- `defaultCashAccount`: 默认现金/资金账户（部分 provider 使用）

## Provider 配置

每个 provider 有独立配置段，段名与 provider 名称一致：

```yaml
alipay:
  rules: [...]

wechat:
  rules: [...]

ccb:
  rules: [...]
```

## 规则优先级

规则按配置中出现的顺序匹配：

1. **自上而下匹配**
2. **后出现的规则优先级更高**
3. **一笔交易可匹配多条规则**
4. **最后匹配到的规则覆盖之前的设置**

## 常见配置方式

### 按时间分类

```yaml
rules:
  - category: 餐饮美食
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch
  - category: 餐饮美食
    time: "17:00-21:00"
    targetAccount: Expenses:Food:Dinner
```

### 按金额分类

```yaml
rules:
  - category: 日用百货
    maxPrice: 10
    targetAccount: Expenses:Food:Snacks
  - category: 日用百货
    minPrice: 10
    targetAccount: Expenses:Groceries
```

### 按商户分类

```yaml
rules:
  - peer: "滴滴,高德,出租车"
    sep: ","
    targetAccount: Expenses:Transport:Taxi
  - peer: "美团,饿了么"
    sep: ","
    targetAccount: Expenses:Food:Delivery
```

## 下一步

- [规则配置详解](rules.md) - 编写匹配规则
- [账户映射](accounts.md) - 账户设置与最佳实践
