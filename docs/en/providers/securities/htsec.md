---
title: Haitong Securities (HTSEC)
layout: default
parent: Provider Support
nav_order: 9
lang: en
---

# Haitong Securities (HTSEC) Provider

The HTSEC Provider supports converting Haitong Securities trading records to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

Log in to Haitong Securities PC trading client, select Query -> Delivery Slip from the left navigation bar, click Query button on the right to export delivery slip Excel file.

## Usage

### Basic Command

```bash
# Convert HTSEC trading records
double-entry-generator translate -p htsec -t beancount htsec_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Broker:HTSEC
defaultPlusAccount: Assets:Broker:HTSEC
defaultCashAccount: Assets:Broker:HTSEC
defaultCurrency: CNY
title: HTSEC Trading Conversion
layout: default

htsec:
  rules:
    - type: 买入
      targetAccount: Assets:Stocks:CN
    - type: 卖出
      targetAccount: Assets:Stocks:CN
      pnlAccount: Income:Broker:PnL
    - type: 分红
      targetAccount: Income:Dividend
```

## Configuration Explanation

### Transaction Types

HTSEC supports various transaction types:
- Stock buy/sell
- Dividend distribution
- Transaction fee records

### Account Settings

- `Assets:Broker:HTSEC`: Broker capital account
- `Assets:Stocks:CN`: Stock position account
- `Income:Broker:PnL`: Trading profit and loss account
- `Income:Dividend`: Dividend income account

## Example Files

- [HTSEC Example](../../example/htsec/example-htsec-output.beancount)
- [Configuration Example](../../example/htsec/config.yaml)
