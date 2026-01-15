---
title: 配置指南
description: 详细的配置和规则设置指南
---


# 配置总览

Double Entry Generator 使用 YAML 格式的配置文件来定义转换规则。

## 基本结构

```yaml
# 全局配置
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:Example
defaultCurrency: CNY
title: 示例配置
layout: default

# Provider 特定配置
providerName:
  rules:
    - # 规则1
    - # 规则2
```

## 全局配置字段

### 必需字段

- `defaultMinusAccount`: 默认的金额减少账户
- `defaultPlusAccount`: 默认的金额增加账户
- `defaultCurrency`: 默认货币代码（如 CNY, USD, EUR）
- `title`: 配置文件标题

### 可选字段

- `defaultCashAccount`: 默认现金账户（部分 provider 使用）

## Provider 配置

每个 provider 都有自己的配置节，节名与 provider 名称相同：

```yaml
alipay:
  rules: [...]

wechat:
  rules: [...]

ccb:
  rules: [...]
```

## 规则优先级

规则按照在配置文件中的顺序进行匹配：

1. **从上到下依次匹配**
2. **后面的规则优先级更高**
3. **一笔交易可以匹配多个规则**
4. **最后匹配的规则会覆盖前面的设置**

## 常见配置模式

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

### 按商家分类

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

- [规则配置详解](rules.html) - 学习如何编写匹配规则
- [账户映射](accounts.html) - 了解账户设置最佳实践