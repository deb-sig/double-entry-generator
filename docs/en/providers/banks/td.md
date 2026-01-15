---
title: Toronto-Dominion Bank (TD)
layout: default
parent: Provider Support
nav_order: 6
lang: en
---

# Toronto-Dominion Bank (TD) Provider

The TD (Toronto-Dominion Bank) Provider supports converting TD Bank bills to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

1. Log in to TD web version: https://easyweb.td.com/
2. Click on the specified account
3. Select bill range -> "Select Download Format" -> Spreadsheet(.csv) -> Download

## Usage

### Basic Command

```bash
# Convert TD bank bills
double-entry-generator translate -p td -t beancount td_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Bank:TD
defaultPlusAccount: Assets:Bank:TD
defaultCashAccount: Assets:Bank:TD
defaultCurrency: CAD
title: TD Bank Bill Conversion
layout: default

td:
  rules:
    - peer: LOBLAWS
      targetAccount: Expenses:Groceries
    - peer: CANADIAN TIRE
      targetAccount: Expenses:Shopping
    - peer: TIM HORTONS
      targetAccount: Expenses:Food:Coffee
```

## Configuration Explanation

### Global Configuration

- Default currency is Canadian Dollar (CAD)
- Supports identification of major Canadian chain stores

### Rule Configuration

The TD Provider is optimized for Canadian local merchants, such as Loblaws, Canadian Tire, Tim Hortons, etc.

## Example Files

- [TD Bank Example](../../example/td/example-td-records.csv)
- [Configuration Example](../../example/td/config.yaml)
- [Output Example](../../example/td/example-td-output.beancount)
