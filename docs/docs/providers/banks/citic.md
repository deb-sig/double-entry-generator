---
title: 中信银行 (CITIC)
---


# 中信银行 (CITIC) Provider

中信银行 Provider 支持将中信银行信用卡账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 账单下载方式

1. 打开中信银行信用卡 PC 官网
2. 使用手机 App 扫码登录
3. 选择账单查询标签页
4. 选择卡片及账单月份
5. 点击"账单下载"

## 使用方法

### 基本命令

```bash
# 转换中信银行信用卡账单
double-entry-generator translate -p citic -t beancount citic_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CC:CITIC
defaultCurrency: CNY
title: 中信银行信用卡账单转换
layout: default

citic:
  rules:
    - peer: 支付宝
      targetAccount: Expenses:Payment:Alipay
      tag: alipay,payment
    - peer: 微信支付
      targetAccount: Expenses:Payment:WeChat
      tag: wechat,payment
    - peer: 滴滴
      targetAccount: Expenses:Transport:Taxi
      tag: transport,taxi
```

## 配置说明

### 全局配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCashAccount`: 中信银行信用卡账户
- `defaultCurrency`: 默认货币

### 规则配置

中信银行 Provider 提供基于规则的匹配，支持按交易对方、类型等进行分类。

## 账户关系

作为信用卡账单，账户关系为：

| 交易类型 | minusAccount | plusAccount |
|----------|-------------|-------------|
| 消费 | defaultCashAccount | targetAccount |
| 还款 | targetAccount | defaultCashAccount |

## 示例文件

- [中信银行信用卡示例](../../example/citic/credit/example-citic-output.beancount)
- [配置示例](../../example/citic/credit/config.yaml)