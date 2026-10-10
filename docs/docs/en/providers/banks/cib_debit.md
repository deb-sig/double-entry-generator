---
title: Industrial Bank Debit Card (CIB Debit)
---

# Industrial Bank Debit Card (CIB Debit) Provider

The CIB Debit provider converts Industrial Bank debit-card Excel exports into Beancount or Ledger output.

## Supported File Formats

- XLSX files
- Multiple currency/subaccount files in one conversion command

## Usage

### Basic Command

```bash
double-entry-generator translate -p cib_debit -t beancount \
  --config example/cib_debit/config.yaml \
  example/cib_debit/example-cib_debit-cny-records.xlsx \
  example/cib_debit/example-cib_debit-hkd-records.xlsx \
  example/cib_debit/example-cib_debit-usd-records.xlsx
```

Ledger output:

```bash
double-entry-generator translate -p cib_debit -t ledger \
  --config example/cib_debit/config.yaml \
  example/cib_debit/example-cib_debit-cny-records.xlsx \
  example/cib_debit/example-cib_debit-hkd-records.xlsx \
  example/cib_debit/example-cib_debit-usd-records.xlsx
```

### Configuration

```yaml
defaultMinusAccount: Income:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:CIB
defaultCurrency: CNY
title: Industrial Bank debit example

cib_debit:
  rules:
    - txType: 存款利息
      targetAccount: Income:Interest
    - txType: 转账转出
      targetAccount: Equity:Transfers
    - txType: 跨行消费,快捷支付
      targetAccount: Expenses:Daily
    - txType: 购汇
      targetAccount: Equity:Transfers
```

## Configuration Notes

### Global Accounts

- `defaultCashAccount`: the Industrial Bank debit-card cash account.
- `defaultMinusAccount`: fallback account for unmatched incoming transactions.
- `defaultPlusAccount`: fallback account for unmatched outgoing transactions.
- `defaultCurrency`: default currency, usually `CNY`.

### Rule Fields

CIB Debit rules support these match fields:

- `peer`: counterparty name
- `peerBank`: counterparty bank
- `peerAccount`: counterparty account number
- `item`: combined summary/purpose description
- `type`: debit/credit direction
- `txType`: transaction summary, such as `存款利息`, `转账转出`, or `购汇`
- `minPrice` / `maxPrice`: amount range
- `sep`: multi-value separator, default `,`
- `fullMatch`: use exact matching
- `ignore`: skip matched transactions
- `tag`: output tags
- `methodAccount`: bank-card account
- `targetAccount`: counterparty account

## Multi-Currency FX Handling

When multiple CIB subaccount files are passed to the same command, the provider matches paired foreign-exchange rows by time and amount. Matched rows are emitted as a single `CurrencyExchange` transaction with a total price (`@@`) in Beancount/Ledger instead of unrelated normal cash-flow rows.

## Example Files

- [Configuration example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/config.yaml)
- [CNY input example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-cny-records.xlsx)
- [HKD input example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-hkd-records.xlsx)
- [USD input example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-usd-records.xlsx)
- [Beancount output example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-output.beancount)
- [Ledger output example](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-output.ledger)
