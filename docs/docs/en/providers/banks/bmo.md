---
title: Bank of Montreal (BMO)
---


# Bank of Montreal (BMO) Provider

The BMO (Bank of Montreal) Provider supports converting BMO bills to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

1. Log in to BMO web version: https://www.bmo.com/en-ca/main/personal/
2. Select the specified account
3. Transactions -> Download, select time range

## Usage

### Basic Command

```bash
# Convert BMO bank bills
double-entry-generator translate -p bmo -t beancount bmo_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Bank:BMO
defaultPlusAccount: Assets:Bank:BMO
defaultCashAccount: Assets:Bank:BMO
defaultCurrency: CAD
title: BMO Bank Bill Conversion
layout: default

bmo:
  rules:
    - peer: GROCERY
      targetAccount: Expenses:Groceries
    - peer: GAS STATION
      targetAccount: Expenses:Transport:Gas
    - peer: RESTAURANT
      targetAccount: Expenses:Food:Restaurant
```

## Configuration Explanation

### Global Configuration

- Default currency is Canadian Dollar (CAD)
- Supports identification of common Canadian merchant types

### Rule Configuration

The BMO Provider is based on North American bank bill format, supporting English merchant name matching.

## Example Files

- [BMO Debit Card Example](../../example/bmo/debit/example-bmo-records.csv)
- [BMO Credit Card Example](../../example/bmo/credit/example-bmo-records.csv)
- [Configuration Example](../../example/bmo/credit/config.yaml)
