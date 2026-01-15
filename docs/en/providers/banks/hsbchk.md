---
title: HSBC Hong Kong (HSBC HK)
layout: default
parent: Provider Support
nav_order: 4
lang: en
---

# HSBC Hong Kong (HSBC HK) Provider

The HSBC HK Provider supports converting HSBC Hong Kong bills to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

1. Log in to HSBC HK online banking
2. Access the account overview page
3. Select the desired account (debit card or credit card)
4. On the transaction details page, select the time period you want to export
5. Click the "Export" button and select CSV format

## Usage

### Basic Command

```bash
# Convert HSBC HK bills
double-entry-generator translate -p hsbchk -t beancount hsbchk_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Bank:HSBC:HK
defaultPlusAccount: Assets:Bank:HSBC:HK
defaultCashAccount: Assets:Bank:HSBC:HK
defaultCurrency: HKD
title: HSBC HK Bill Conversion
layout: default

hsbchk:
  rules:
    - peer: 支付宝
      targetAccount: Expenses:Payment:Alipay
    - peer: 八达通
      targetAccount: Expenses:Transport:Octopus
    - peer: 便利店
      targetAccount: Expenses:Food:Convenience
```

## Configuration Explanation

Supports common Hong Kong payment methods and merchant categories, including Octopus, convenience stores, and other Hong Kong-specific consumption scenarios.

### Currency Support

Default support for Hong Kong Dollar (HKD), other currencies can also be configured.

## Example Files

- [HSBC HK Debit Card Example](../../example/hsbchk/debit/example-hsbchk-debit-records.csv)
- [HSBC HK Credit Card Example](../../example/hsbchk/credit/example-hsbchk-credit-records.csv)
- [Configuration Example](../../example/hsbchk/credit/config.yaml)
