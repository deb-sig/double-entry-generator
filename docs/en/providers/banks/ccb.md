---
title: China Construction Bank (CCB)
layout: default
parent: Provider Support
nav_order: 1
lang: en
---

# China Construction Bank (CCB) Provider

The CCB Provider supports converting China Construction Bank bills to Beancount/Ledger format.

## Supported File Formats

- CSV format
- XLS format
- XLSX format

## Usage

### Basic Command

```bash
# Convert CCB bills
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:CCB
defaultCurrency: CNY
title: CCB Bill Conversion
layout: default

ccb:
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
- `defaultCashAccount`: CCB account
- `defaultCurrency`: Default currency

### Rule Configuration

The CCB Provider provides rule-based matching, you can specify:

- `item` (Transaction description) exact/contains matching
- `peer` (Transaction counterpart) exact/contains matching
- `type` (Transaction type) exact/contains matching
- `status` (Transaction status) exact/contains matching
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

## File Format Explanation

CCB bill files typically contain the following information:

1. **File Header**: Contains bank information and account information
2. **Data Rows**: Contains transaction records
3. **File Footer**: Contains statistical information

### Data Fields

- **Transaction Date**: Transaction posting date
- **Trade Date**: Actual transaction date
- **Transaction Amount**: Transaction amount (positive for income, negative for expense)
- **Transaction Type**: Transaction type description
- **Transaction Counterpart**: Transaction counterpart information
- **Transaction Status**: Transaction status (success, failure, etc.)

## Example Files

- [CCB Bill Example](../../example/ccb/建设银行_xxxx_2025xxxx_2025xxxx.xls)
- [Configuration Example](../../example/ccb/config.yaml)
- [Output Example](../../example/ccb/example-ccb-output.beancount)
