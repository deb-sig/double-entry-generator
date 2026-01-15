---
title: Huobi
layout: default
parent: Provider Support
nav_order: 11
lang: en
---

# Huobi Provider

The Huobi Provider supports converting Huobi spot trading records to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

Log in to [Huobi Global website](https://www.huobi.com/), go to [Spot Order Transaction Details](https://www.huobi.com/zh-cn/transac/?tab=2&type=0) page, select appropriate time range, then click the export button at the top right of transaction details.

## Usage

### Basic Command

```bash
# Convert Huobi trading records
double-entry-generator translate -p huobi -t beancount huobi_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:Crypto:Huobi
defaultPlusAccount: Assets:Crypto:Huobi
defaultCashAccount: Assets:Crypto:Huobi
defaultCurrency: USDT
title: Huobi Trading Conversion
layout: default

huobi:
  rules:
    - type: 买入
      symbol: BTC
      targetAccount: Assets:Crypto:BTC
    - type: 卖出
      symbol: BTC
      targetAccount: Assets:Crypto:BTC
      pnlAccount: Income:Crypto:PnL
    - type: 买入
      symbol: ETH
      targetAccount: Assets:Crypto:ETH
```

## Configuration Explanation

### Transaction Types

Huobi supports various spot trading:
- Spot buy/sell
- Exchange between different currencies
- Transaction fee records

### Account Settings

- `Assets:Crypto:Huobi`: Huobi account (usually priced in USDT)
- `Assets:Crypto:BTC`: BTC position account
- `Assets:Crypto:ETH`: ETH position account
- `Income:Crypto:PnL`: Trading profit and loss account

### Currency Support

Supports mainstream cryptocurrencies:
- BTC (Bitcoin)
- ETH (Ethereum)
- USDT (Tether)
- And other currencies supported by Huobi

## Example Files

- [Huobi Trading Example](../../example/huobi/example-huobi-records.csv)
- [Configuration Example](../../example/huobi/config.yaml)
- [Output Example](../../example/huobi/example-huobi-output.beancount)
