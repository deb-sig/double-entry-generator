---
title: 工商银行 (ICBC)
layout: default
parent: 提供商支持
nav_order: 2
---

# 工商银行 (ICBC) Provider

工商银行 Provider 支持将工商银行账单转换为 Beancount/Ledger 格式，可自动识别借记卡和信用卡账单。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换工商银行账单
double-entry-generator translate -p icbc -t beancount icbc_records.csv

# 指定配置文件
double-entry-generator translate -p icbc -t beancount -c config.yaml icbc_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Bank:ICBC
defaultPlusAccount: Assets:Bank:ICBC
defaultCashAccount: Assets:Bank:ICBC
defaultCurrency: CNY
title: 工商银行账单转换
layout: default

icbc:
  rules:
    - peer: 支付宝
      methodAccount: Assets:Bank:ICBC
      targetAccount: Expenses:Payment:Alipay
      tag: alipay,payment
    - peer: 微信
      methodAccount: Assets:Bank:ICBC
      targetAccount: Expenses:Payment:WeChat
      tag: wechat,payment
    - peer: 滴滴
      methodAccount: Assets:Bank:ICBC
      targetAccount: Expenses:Transport:Taxi
      tag: transport,taxi
```

## 配置说明

### 全局配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCashAccount`: 工商银行账户
- `defaultCurrency`: 默认货币

### 规则配置

工商银行 Provider 提供基于规则的匹配，可以指定：

- `peer`（交易对方）的完全/包含匹配
- `type`（收/支）的完全/包含匹配
- `txType`（摘要）的完全/包含匹配

### 规则选项

- `sep`: 分隔符，默认为 `,`
- `fullMatch`: 是否使用完全匹配，默认为 `false`
- `tag`: 设置流水的 Tag
- `ignore`: 是否忽略匹配的交易，默认为 `false`

## 账户类型识别

ICBC Provider 可以自动识别账单类型：

- **借记卡账单**: 储蓄卡消费记录
- **信用卡账单**: 信用卡消费记录

## 示例文件

- [工商银行借记卡示例](../../example/icbc/debit-v1/example-icbc-debit-v1-records.csv)
- [工商银行信用卡示例](../../example/icbc/credit/example-icbc-credit-records.csv)
- [配置示例](../../example/icbc/credit/config.yaml)