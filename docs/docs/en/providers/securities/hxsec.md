---
title: Huaxi Securities (HXSEC)
---


# Huaxi Securities (HXSEC) Provider

The HXSEC Provider supports converting Huaxi Securities (Tongdaxin) trading records to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Usage

### Basic Command

```bash
# Convert HXSEC trading records
double-entry-generator translate -p hxsec -t beancount hxsec_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Broker:HXSEC
defaultPlusAccount: Assets:Broker:HXSEC
defaultCashAccount: Assets:Broker:HXSEC
defaultCurrency: CNY
title: HXSEC Trading Conversion
layout: default

hxsec:
  rules:
    - type: 证券买入
      targetAccount: Assets:Stocks:CN
    - type: 证券卖出
      targetAccount: Assets:Stocks:CN
      pnlAccount: Income:Broker:PnL
    - type: 股息红利
      targetAccount: Income:Dividend
```

## Configuration Explanation

### Transaction Types

Huaxi Securities (Tongdaxin system) supports:
- Stock trading
- Dividend distribution
- Transaction fee deduction

### Tongdaxin Features

Huaxi Securities uses Tongdaxin trading system, file format is similar to other Tongdaxin brokers.

## Example Files

- [HXSEC Example](../../example/hxsec/example-hxsec-output.beancount)
- [Configuration Example](../../example/hxsec/config.yaml)
