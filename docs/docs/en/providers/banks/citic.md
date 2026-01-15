---
title: China CITIC Bank (CITIC)
---


# China CITIC Bank (CITIC) Provider

The CITIC Provider supports converting CITIC Bank credit card bills to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

1. Open CITIC Bank credit card PC official website
2. Log in using mobile App QR code scanning
3. Select the bill query tab
4. Select card and bill month
5. Click "Download Bill"

## Usage

### Basic Command

```bash
# Convert CITIC Bank credit card bills
double-entry-generator translate -p citic -t beancount citic_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CC:CITIC
defaultCurrency: CNY
title: CITIC Bank Credit Card Bill Conversion
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

## Configuration Explanation

### Global Configuration

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCashAccount`: CITIC Bank credit card account
- `defaultCurrency`: Default currency

### Rule Configuration

The CITIC Provider provides rule-based matching, supporting categorization by transaction counterpart, type, etc.

## Account Relationships

As a credit card bill, the account relationships are:

| Transaction Type | minusAccount | plusAccount |
|----------|-------------|-------------|
| Expense | defaultCashAccount | targetAccount |
| Payment | targetAccount | defaultCashAccount |

## Example Files

- [CITIC Bank Credit Card Example](../../example/citic/credit/example-citic-output.beancount)
- [Configuration Example](../../example/citic/credit/config.yaml)
