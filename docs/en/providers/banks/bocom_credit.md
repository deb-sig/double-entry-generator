---
title: Bank of Communications Credit Card (BOCOM Credit)
layout: default
parent: Provider Support
nav_order: 9
lang: en
---

# Bank of Communications Credit Card (BOCOM Credit) Provider

The BOCOM Credit Provider converts BOCOM credit card transaction details (EML needs to be converted to CSV first) exported from BOCOM credit card app to Beancount/Ledger entries.

## Supported File Formats

- CSV (obtained by converting EML email exported from BOCOM credit card app using [bill-file-converter](https://github.com/deb-sig/bill-file-converter))

## Download Method

1. Open BOCOM credit card app, search for "账单补发" (Bill Reissue)
2. Select the bill month to reissue, click "下一步" (Next) at the bottom
3. Confirm email address and confirm export
4. Download the received email as EML file and convert to CSV using bill-file-converter

## Usage

### Basic Command

```bash
double-entry-generator translate \
  --config ./example/bocom_credit/config.yaml \
  --provider bocom_credit \
  --output ./example/bocom_credit/example-bocom_credit-output.beancount \
  ./example/bocom_credit/example-bocom_credit-records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultCurrency: CNY
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:BOCOM:CreditCard
bocom_credit:
  rules:
    - item: 信用卡还款
      targetAccount: Equity:Transfers
    - item: 美团,饿了么
      targetAccount: Expenses:Food
```

## Example Files

- [Transaction Detail CSV Example](../../example/bocom_credit/example-bocom_credit-records.csv)
- [Converted Beancount Example](../../example/bocom_credit/example-bocom_credit-output.beancount)
- [Configuration File Example](../../example/bocom_credit/config.yaml)
