---
title: Industrial and Commercial Bank of China (ICBC)
---


# Industrial and Commercial Bank of China (ICBC) Provider

The ICBC Provider supports converting ICBC bills to Beancount/Ledger format, and can automatically identify debit card and credit card bills.

## Supported File Formats

- CSV format

## Bill Download Method

ICBC bill download method can be found [here](https://blog.triplez.cn/posts/bills-export-methods/#%e4%b8%ad%e5%9b%bd%e5%b7%a5%e5%95%86%e9%93%b6%e8%a1%8c).

## Usage

### Basic Command

```bash
# Convert ICBC bills
double-entry-generator translate -p icbc -t beancount icbc_records.csv

# Specify configuration file
double-entry-generator translate -p icbc -t beancount -c config.yaml icbc_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Bank:ICBC
defaultPlusAccount: Assets:Bank:ICBC
defaultCashAccount: Assets:Bank:ICBC
defaultCurrency: CNY
title: ICBC Bill Conversion
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

## Configuration Explanation

### Global Configuration

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCashAccount`: ICBC account
- `defaultCurrency`: Default currency

### Rule Configuration

The ICBC Provider provides rule-based matching, you can specify:

- `peer` (Transaction counterpart) exact/contains matching
- `type` (Income/Expense) exact/contains matching
- `txType` (Summary) exact/contains matching

### Rule Options

- `sep`: Separator, default is `,`
- `fullMatch`: Whether to use exact match, default is `false`
- `tag`: Set transaction Tag
- `ignore`: Whether to ignore matched transactions, default is `false`

## Account Type Identification

The ICBC Provider can automatically identify bill types:

- **Debit Card Bills**: Savings card transaction records
- **Credit Card Bills**: Credit card transaction records

## Example Files

- [ICBC Debit Card Example](../../example/icbc/debit-v1/example-icbc-debit-v1-records.csv)
- [ICBC Credit Card Example](../../example/icbc/credit/example-icbc-credit-records.csv)
- [Configuration Example](../../example/icbc/credit/config.yaml)
