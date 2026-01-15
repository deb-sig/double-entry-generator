---
title: Agricultural Bank of China Debit Card (ABC Debit)
layout: default
parent: Provider Support
nav_order: 10
lang: en
---

# Agricultural Bank of China Debit Card (ABC Debit) Provider

The ABC Debit Provider supports converting ABC App exported debit card transaction detail CSV to Beancount/Ledger entries.

## Supported File Formats

- CSV (obtained by converting PDF exported from ABC App using [bill-file-converter](https://github.com/deb-sig/bill-file-converter))

## Download Method

1. Open ABC App, go to homepage "我的账户" (My Account)
2. Click debit card "明细查询" (Transaction Details)
3. Click export button at top right of "明细查询" page
4. Complete account, currency, time, and email form fields, click confirm
5. Convert the received PDF to CSV using bill-file-converter

## Usage

### Beancount

```bash
double-entry-generator translate \
  --config ./example/abc_debit/config.yaml \
  --provider abc_debit \
  --output ./example/abc_debit/example-abc_debit-output.beancount \
  ./example/abc_debit/example-abc_debit-records.csv
```

### Ledger

```bash
double-entry-generator translate \
  --config ./example/abc_debit/config.yaml \
  --provider abc_debit \
  --target ledger \
  --output ./example/abc_debit/example-abc_debit-output.ledger \
  ./example/abc_debit/example-abc_debit-records.csv
```

## Configuration File Example

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:ABC:DebitCard
defaultCurrency: CNY
title: abc_debit
abc_debit:
  rules:
    - item: 转存
      targetAccount: Equity:Transfers
    - item: 财付通
      targetAccount: Expenses:Transport:Transit
      tag: transport
    - item: 正常还款
      targetAccount: Liabilities:Loans:Personal
    - item: 结息
      targetAccount: Income:Interest
    - item: 利息税
      targetAccount: Expenses:Tax
```

## Example Files

- [Transaction Detail CSV Example](../../example/abc_debit/example-abc_debit-records.csv)
- [Converted Beancount Example](../../example/abc_debit/example-abc_debit-output.beancount)
- [Converted Ledger Example](../../example/abc_debit/example-abc_debit-output.ledger)
- [Configuration File Example](../../example/abc_debit/config.yaml)
