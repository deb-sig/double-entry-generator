---
title: Bank of Communications Debit Card (BOCOM Debit)
layout: default
parent: Provider Support
nav_order: 8
lang: en
---

# Bank of Communications Debit Card (BOCOM Debit) Provider

The BOCOM Debit Provider converts BOCOM App exported transaction details (PDF needs to be converted to CSV first) to Beancount/Ledger entries.

## Supported File Formats

- CSV (obtained by converting PDF exported from BOCOM App using [bill-file-converter](https://github.com/deb-sig/bill-file-converter))

## Download Method

1. Open BOCOM App, search for "交易明细" (Transaction Details)
2. Click "导出交易明细" (Export Transaction Details) at the bottom, select electronic version
3. Select card number and set custom time range, then click "去开立" (Go to Open)
4. Set bill format and enable all options, fill in receiving email address and confirm export
5. Convert the received PDF file to CSV using bill-file-converter

## Usage

### Basic Command

```bash
double-entry-generator translate \
  --config ./example/bocom_debit/config.yaml \
  --provider bocom_debit \
  --output ./example/bocom_debit/example-bocom-debit-output.beancount \
  ./example/bocom_debit/example-bocom-debit-records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:BOCOM
defaultCurrency: CNY
title: BOCOM Debit Card Example

bocom_debit:
  rules:
    - peer: 应付个人活期储蓄存款利息
      targetAccount: Income:Interest
    - peer: 信用卡还款
      targetAccount: Equity:Transfers
    - txType: 信用卡转账还款
      targetAccount: Equity:Transfers
    - peer: 网上国网
      targetAccount: Expenses:Electricity
    - item: 基金理财产品申购
      targetAccount: Assets:Funds
    - item: 财付通
      ignore: true
```

## Example Files

- [Transaction Detail CSV Example](../../example/bocom_debit/example-bocom-debit-records.csv)
- [Converted Beancount Example](../../example/bocom_debit/example-bocom-debit-output.beancount)
- [Configuration File Example](../../example/bocom_debit/config.yaml)
