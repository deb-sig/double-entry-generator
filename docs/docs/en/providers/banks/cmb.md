---
title: China Merchants Bank (CMB)
---


# China Merchants Bank (CMB) Provider

The CMB Provider supports converting CMB bills to Beancount/Ledger format, supporting both savings card and credit card bills.

## Supported File Formats

- CSV format

## Usage

### Basic Command

```bash
# Convert CMB savings card bills
double-entry-generator translate -p cmb -t beancount -c config.yaml cmb_records.csv

# Convert CMB credit card bills
double-entry-generator translate -p cmb -t beancount -c config.yaml cmb_records.csv
```

### Configuration File

#### Savings Card Configuration Example

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:CMB
defaultCurrency: CNY
title: CMB Savings Card Bill Conversion

cmb:
  rules:
    # Expenses
    - peer: 电费,网上国网,国网
      targetAccount: Expenses:Electricity
    - peer: 中国移动
      targetAccount: Expenses:Mobile
    # Insurance claims
    - peer: 太平洋健康保险股份有限公司
      item: 汇入汇款
      targetAccount: Income:Insurance
```

#### Credit Card Configuration Example

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CreditCard:CMB
defaultCurrency: CNY
title: CMB Credit Card Bill Conversion

cmb:
  rules:
    - item: 掌上生活影票
      targetAccount: Expenses:Movie
    - item: 手机银行饭票
      targetAccount: Expenses:Food
    - item: 中国移动
      targetAccount: Expenses:Mobile
    - item: 财付通
      ignore: true
```

## Configuration Explanation

### Global Configuration

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCashAccount`: CMB account
  - Savings card: `Assets:DebitCard:CMB`
  - Credit card: `Liabilities:CreditCard:CMB`
- `defaultCurrency`: Default currency

### Rule Configuration

The CMB Provider provides rule-based matching, you can specify:

- `peer` (Transaction counterpart) exact/contains matching
- `item` (Product description) exact/contains matching
- `type` (Transaction type) exact/contains matching
- `txType` (Transaction type) exact/contains matching

### Rule Options

- `sep`: Separator, default is `,`
- `fullMatch`: Whether to use exact match, default is `false`
- `tag`: Set transaction Tag
- `ignore`: Whether to ignore matched transactions, default is `false`
- `methodAccount`: Payment account (optional)
- `targetAccount`: Target account

## Account Relationships

`targetAccount` and `defaultCashAccount` increase/decrease account relationships:

| Income/Expense | minusAccount       | plusAccount        |
|-------|-------------------|-------------------|
| Income  | targetAccount     | defaultCashAccount |
| Expense  | defaultCashAccount | targetAccount      |

## Bill Download Method

### Savings Card Bills

1. Open CMB App
2. Search for "流水打印" (Transaction Print)
3. Switch to "高级筛选" (Advanced Filter) at bottom right
4. Select card number, start date, end date
5. Set bill format
   - "展示摘要类型" (Display Summary Type) select "全部" (All)
   - "展示交易对手信息" (Display Transaction Counterpart Info) select "开启" (On)
   - "展示完整卡号" (Display Full Card Number) select "开启" (On)
   - "展示收入及支出汇总金额" (Display Income and Expense Summary) select "关闭" (Off)
   - "交易币种" (Transaction Currency) select "全部" (All)
   - "金额区间" (Amount Range) select "关闭" (Off)
   - "交易类型" (Transaction Type) select "全部" (All)
   - "仅展示活期户流水" (Only Show Current Account Transactions) select "关闭" (Off)
6. Fill in receiving email address, confirm export
7. Convert the exported PDF file to CSV using [bill-file-converter](https://github.com/deb-sig/bill-file-converter)

### Credit Card Bills

1. Open CMB Life App
2. Search for "账单补寄" (Bill Reissue)
3. Select billing cycle
4. Submit application, confirm export
5. Convert the exported PDF file to CSV using [bill-file-converter](https://github.com/deb-sig/bill-file-converter)

## Example Files

- [CMB Savings Card Example](../../example/cmb/debit/example-cmb-records.csv)
- [CMB Credit Card Example](../../example/cmb/credit/example-cmb-records.csv)
- [Savings Card Configuration Example](../../example/cmb/debit/config.yaml)
- [Credit Card Configuration Example](../../example/cmb/credit/config.yaml)
- [Output Example](../../example/cmb/debit/example-cmb-output.beancount)
