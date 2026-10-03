---
title: Interactive Brokers (IBKR)
---

# Interactive Brokers (IBKR) Provider

The IBKR provider converts Interactive Brokers Flex Query XML exports into Beancount or Ledger output.

## Supported File Formats

- Flex Query XML

## Supported Record Types

The provider reads these Flex XML nodes:

- `Trade` with `assetCategory="STK"`: stock, ETF, and ADR trades, emitted as securities trades.
- `Trade` with `assetCategory="CASH"`: FX trades such as `USD.HKD`, emitted as currency-exchange transactions.
- `CashTransaction`: deposits, withdrawals, dividends, withholding tax, interest, and fees.

`StatementOfFundsLine` is a cash-flow view and can duplicate `Trade` / `CashTransaction` data, so it is not used as the primary conversion source.

## Usage

### Basic Command

```bash
double-entry-generator translate -p ibkr -t beancount \
  --config example/ibkr/config.yaml \
  example/ibkr/example-ibkr-records.xml
```

Ledger output:

```bash
double-entry-generator translate -p ibkr -t ledger \
  --config example/ibkr/config.yaml \
  example/ibkr/example-ibkr-records.xml
```

### Configuration

```yaml
defaultMinusAccount: Equity:Opening-Balances
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:IBKR:Cash
defaultPositionAccount: Assets:IBKR:Positions
defaultCommissionAccount: Expenses:IBKR:Fees
defaultPnlAccount: Income:IBKR:PnL
defaultCurrency: USD
bookingMethod: FIFO
title: IBKR

ibkr:
  rules:
    - item: "TSLA,NIO"
      sep: ","
      positionAccount: Assets:IBKR:Positions:Stocks
    - type: "Withholding Tax"
      commissionAccount: Expenses:IBKR:Tax
```

## Configuration Notes

### Global Accounts

- `defaultCashAccount`: IBKR cash account.
- `defaultPositionAccount`: default securities position account.
- `defaultCommissionAccount`: default commission/tax account.
- `defaultPnlAccount`: default income/PnL account for dividends, interest, and securities sell balancing.
- `defaultMinusAccount`: default source account for deposits.
- `bookingMethod`: optional Beancount booking method. The example uses `FIFO` to avoid ambiguous lot matching for repeated securities sells.

### Rule Fields

IBKR rules support:

- `item`: security name or cash transaction description
- `type`: original type, such as `BUY`, `SELL`, or `Withholding Tax`
- `sep`: multi-value separator, default `,`
- `fullMatch`: use exact matching
- `ignore`: skip matched transactions
- `cashAccount`: cash account
- `positionAccount`: position account
- `commissionAccount`: commission/tax account
- `pnlAccount`: PnL account

## Output Details

The IBKR provider preserves source metadata where present, including `account_id`, `trade_id`, `transaction_id`, `ib_order_id`, `ib_exec_id`, `exchange`, `isin`, `security_id`, `report_date`, and `settle_date`.

Stock buys emit cash decrease, position increase, and commission postings. Stock sells emit position decrease, cash increase, commission postings, and a PnL balancing posting. Cash FX trades emit `CurrencyExchange` transactions.

## Example Files

- [Configuration example](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/config.yaml)
- [Flex XML input example](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-records.xml)
- [Beancount output example](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-output.beancount)
- [Ledger output example](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-output.ledger)
