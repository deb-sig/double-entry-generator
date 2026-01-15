---
title: Alipay
layout: default
parent: Provider Support
nav_order: 7
lang: en
---

# Alipay Provider

The Alipay Provider supports converting Alipay bills to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

For double-entry generator `v1.0.0` and above, please refer to [this article](https://blog.triplez.cn/posts/bills-export-methods/#%e6%94%af%e4%bb%98%e5%ae%9d) to obtain Alipay bills.

For double-entry generator `v0.2.0` and below, use this method: Log in to PC Alipay, visit [here](https://consumeprod.alipay.com/record/standard.htm), select a time range, scroll to the bottom of the page, and click "Download Query Results". Note: Please download the query results, not the [Income and Expense Details](https://cshall.alipay.com/lab/help_detail.htm?help_id=212688).

## Usage

### Basic Command

```bash
# Convert Alipay bills
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# Specify configuration file
double-entry-generator translate -p alipay -t beancount -c config.yaml alipay_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: Alipay Bill Conversion
layout: default

alipay:
  rules:
    # Income transactions
    - type: 收入
      item: 商品
      targetAccount: Income:Alipay:ShouKuanMa
      methodAccount: Assets:Alipay
    
    # Categorize expenses by category
    - category: 日用百货
      minPrice: 10
      targetAccount: Expenses:Groceries
    - category: 日用百货
      maxPrice: 9.99
      targetAccount: Expenses:Food:Drink
    
    # Categorize food by time
    - category: 餐饮美食
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: 16:00-22:00
      targetAccount: Expenses:Food:Dinner
    
    # Match by merchant
    - peer: 滴滴出行
      targetAccount: Expenses:Transport
    - peer: 苏宁
      targetAccount: Expenses:Electronics
    
    # Payment method matching
    - method: 余额
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 余额宝
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 交通银行信用卡(7449)
      fullMatch: true
      methodAccount: Liabilities:CC:COMM:7449
    
    # Investment related
    - peer: 基金
      type: 其他
      item: 买入
      methodAccount: Assets:Alipay
      targetAccount: Assets:Alipay:Invest:Fund
    - peer: 基金
      type: 其他
      item: 黄金-买入
      methodAccount: Assets:Alipay
      targetAccount: Assets:Alipay:Invest:Gold
```

## Configuration Explanation

### Global Configuration

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCurrency`: Default currency

### Rule Configuration

The Alipay Provider provides rule-based matching, you can specify:

- `type` (Transaction type) - Expense/Income/Other
- `category` (Consumption category) - 餐饮美食/日用百货/交通出行, etc.
- `peer` (Transaction counterpart) - Merchant name
- `item` (Product description) - Specific product description
- `method` (Payment method) - 余额/余额宝/Credit card, etc.
- `time` (Transaction time) - Time range matching
- `minPrice`/`maxPrice` - Amount range matching

### Rule Options

- `sep`: Separator, default is `,`, used for multi-keyword matching
- `fullMatch`: Whether to use exact match, default is `false`
- `tag`: Set transaction Tag
- `methodAccount`: Specify payment account (e.g., balance, credit card, etc.)
- `targetAccount`: Specify target account
- `pnlAccount`: Investment profit and loss account (for fund, gold trading)

## Account Relationships

Alipay's account relationships are relatively complex because they involve multiple payment methods:

### Basic Expenses
- **Expense transaction**: `methodAccount` → `targetAccount`
- **Income transaction**: `targetAccount` → `methodAccount`

### Investment Transactions
- **Buy**: `methodAccount` → `targetAccount`
- **Sell**: `targetAccount` → `methodAccount` + `pnlAccount` (profit/loss)

## Special Features

### 1. Payment Method Identification
Alipay bills contain detailed payment method information, which can be precisely mapped to different accounts:
- 余额 → `Assets:Alipay`
- 余额宝 → `Assets:Alipay:YuEBao`
- Credit card → `Liabilities:CC:Bank:Number`

### 2. Time-Based Categorization
Supports automatic food categorization by time:
```yaml
- category: 餐饮美食
  time: 11:00-14:00
  targetAccount: Expenses:Food:Lunch
```

### 3. Investment Transaction Support
Built-in support for profit and loss accounting for fund and gold trading.

## Example Files

- [Alipay Bill Example](../../example/alipay/example-alipay-records.csv)
- [Configuration Example](../../example/alipay/config.yaml)
- [Output Example](../../example/alipay/example-alipay-output.beancount)
