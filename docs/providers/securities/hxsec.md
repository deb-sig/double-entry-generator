---
title: 华西证券 (HXSEC)
layout: default
parent: 提供商支持
nav_order: 10
---

# 华西证券 (HXSEC) Provider

华西证券 Provider 支持将华西证券（通达信）交易记录转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换华西证券交易记录
double-entry-generator translate -p hxsec -t beancount hxsec_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Broker:HXSEC
defaultPlusAccount: Assets:Broker:HXSEC
defaultCashAccount: Assets:Broker:HXSEC
defaultCurrency: CNY
title: 华西证券交易转换
layout: default

hxsec:
  rules:
    - type: 证券买入
      targetAccount: Assets:Stocks:CN
    - type: 证券卖出
      targetAccount: Assets:Stocks:CN
      pnlAccount: Income:Broker:PnL
    - type: 股息红利
      targetAccount: Income:Dividend
```

## 配置说明

### 交易类型

华西证券（通达信系统）支持：
- 证券买卖交易
- 股息红利发放
- 交易手续费扣除

### 通达信特色

华西证券使用通达信交易系统，文件格式与其他通达信券商类似。

## 示例文件

- [华西证券示例](../../example/hxsec/example-hxsec-output.beancount)
- [配置示例](../../example/hxsec/config.yaml)