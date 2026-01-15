---
title: WeChat Pay
layout: default
parent: Provider Support
nav_order: 8
lang: en
---

# WeChat Pay Provider

The WeChat Provider supports converting WeChat bills to Beancount/Ledger format.

## Supported File Formats

- CSV format (default)
- XLSX format (newly supported)

## Bill Download Method

WeChat Pay download method can be found [here](https://blog.triplez.cn/posts/bills-export-methods/#%e5%be%ae%e4%bf%a1%e6%94%af%e4%bb%98).

## Usage

### Basic Command

```bash
# Convert CSV format WeChat bills
double-entry-generator translate -p wechat -t beancount wechat_records.csv

# Convert XLSX format WeChat bills
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat
defaultCurrency: CNY
title: WeChat Bill Conversion
layout: default

wechat:
  rules:
    - item: 三快
      targetAccount: Expenses:Food
    - item: 电子商务,天猫,京东,特约商户
      targetAccount: Expenses:Shopping
    - item: 电费,网上国网
      targetAccount: Expenses:Electricity
    - item: 滴滴出行,嘀嘀,中国石油
      targetAccount: Expenses:Transport
    - item: 现金奖励
      targetAccount: Income:Rewards
    - item: 财付通(银联云闪付)
      ignore: true
    - item: 财付通还款
      targetAccount: Assets:WeChat
```

## Configuration Explanation

### Global Configuration

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCashAccount`: WeChat account (equivalent to `methodAccount` in Alipay)
- `defaultCurrency`: Default currency

### Rule Configuration

The WeChat Provider provides rule-based matching, you can specify:

- `item` (Transaction description) exact/contains matching
- `peer` (Transaction counterpart) exact/contains matching
- `type` (Transaction type) exact/contains matching
- `time` (Transaction time) range matching
- `minPrice` (Minimum amount) and `maxPrice` (Maximum amount) range matching

### Rule Options

- `sep`: Separator, default is `,`
- `fullMatch`: Whether to use exact match, default is `false`
- `tag`: Set transaction Tag
- `ignore`: Whether to ignore matched transactions, default is `false`

## Account Relationships

`targetAccount` and `defaultCashAccount` increase/decrease account relationships:

| Income/Expense | minusAccount | plusAccount |
|-------|-------------|-------------|
| Income | targetAccount | defaultCashAccount |
| Expense | defaultCashAccount | targetAccount |

## Example Files

- [WeChat Bill Example (CSV)](../../example/wechat/example-wechat-records.csv)
- [WeChat Bill Example (XLSX)](../../example/wechat/example-wechat-records.xlsx)
- [Configuration Example](../../example/wechat/config.yaml)
