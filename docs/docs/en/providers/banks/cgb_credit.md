---
title: CGB Credit Card (CGB Credit)
---

# CGB Credit Card (CGB Credit) Provider

The CGB Credit provider converts CSV transaction details generated from CGB credit card PDF statements into Beancount/Ledger entries.

## Supported File Formats

- CSV generated from CGB credit card PDF statements by [bill-file-converter](https://github.com/deb-sig/bill-file-converter)

## Usage

```bash
double-entry-generator translate \
  --config ./example/cgb_credit/config.yaml \
  --provider cgb_credit \
  --output ./example/cgb_credit/example-cgb_credit-output.beancount \
  ./example/cgb_credit/example-cgb_credit-records.csv
```

## Configuration

```yaml
title: cgb_credit
defaultCurrency: CNY
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CGB:CreditCard
cgb_credit:
  rules:
    - type: 还款
      targetAccount: Equity:Transfers
    - item: 美团
      targetAccount: Expenses:Food
```

## Notes

- Positive amounts are treated as card spending or installment charges.
- Negative amounts are treated as refunds or repayments, reducing the credit card liability.
- If transaction currency differs from settlement currency, the original transaction amount is preserved as metadata.
