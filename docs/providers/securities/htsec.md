---
title: 海通证券 (HTSEC)
parent: 提供商支持
nav_order: 9
---

# 海通证券 (HTSEC) Provider

海通证券 Provider 支持将海通证券交易记录转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换海通证券交易记录
double-entry-generator translate -p htsec -t beancount htsec_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Broker:HTSEC
defaultPlusAccount: Assets:Broker:HTSEC
defaultCashAccount: Assets:Broker:HTSEC
defaultCurrency: CNY
title: 海通证券交易转换

htsec:
  rules:
    - type: 买入
      targetAccount: Assets:Stocks:CN
    - type: 卖出
      targetAccount: Assets:Stocks:CN
      pnlAccount: Income:Broker:PnL
    - type: 分红
      targetAccount: Income:Dividend
```

## 配置说明

### 交易类型

海通证券支持多种交易类型：
- 股票买入/卖出
- 分红派息
- 手续费记录

### 账户设置

- `Assets:Broker:HTSEC`: 券商资金账户
- `Assets:Stocks:CN`: 股票持仓账户
- `Income:Broker:PnL`: 交易损益账户
- `Income:Dividend`: 股息收入账户

## 示例文件

- [海通证券示例](../../example/htsec/example-htsec-output.beancount)
- [配置示例](../../example/htsec/config.yaml)