# Bank of China

The `boc` provider supports Bank of China debit card statements and credit card statements. Debit card statements are usually exported as Bank of China transaction detail statements, while credit card statements are exported as Bank of China credit card bills.

## Configuration Example

```yaml
defaultMinusAccount: Assets:Bank:CN:BOC:Savings
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: Bank of China
boc:
  rules:
    - method: "1528"
      methodAccount: Assets:Bank:CN:BOC:Savings
    - item: 结息,利息
      targetAccount: Income:Bank:CN:Interest
    - peer: 支付宝,微信,京东
      sep: ","
      ignore: true
```

## Rule Fields

- `peer`: counterparty. For debit card statements, the provider combines the counterparty name and counterparty card/account number for matching, so either a name or account-number fragment can be used.
- `item`: transaction name or summary.
- `method`: last four digits of the source card number.
- `type`: transaction direction, such as income or expense.
- `time` / `timestamp_range`: transaction-time matching.
- `minPrice` / `maxPrice`: amount-range matching.

## Convert Command

```bash
double-entry-generator translate \
  --config ./config.yaml \
  --provider boc \
  --target beancount \
  --output output.beancount \
  statement.csv
```
