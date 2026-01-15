---
title: JD.com (JD)
---


# JD.com (JD) Provider

The JD Provider supports converting JD.com shopping orders to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Bill Download Method

1. Open JD.com mobile APP
2. Go to My -> My Wallet -> Bills
3. Click the icon at top right (three horizontal lines)
4. Select "账单导出（仅限个人对账）" (Bill Export - Personal Reconciliation Only)

## Usage

### Basic Command

```bash
# Convert JD.com shopping records
double-entry-generator translate -p jd -t beancount jd_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat  # WeChat Pay or other payment methods
defaultCurrency: CNY
title: JD.com Shopping Conversion
layout: default

jd:
  rules:
    - category: 食品饮料
      targetAccount: Expenses:Food:Groceries
      tag: food,groceries
    - category: 家居家装
      targetAccount: Expenses:Home:Furniture
      tag: home,furniture
    - category: 数码
      targetAccount: Expenses:Electronics
      tag: electronics
    - category: 服装
      targetAccount: Expenses:Clothing
      tag: clothing
    - category: 图书
      targetAccount: Expenses:Books
      tag: books,education
```

## Configuration Explanation

### Product Categories

JD.com supports rich product categories:
- **食品饮料**: Fresh food, snacks, beverages, etc.
- **家居家装**: Furniture, home appliances, decoration supplies
- **数码**: Mobile phones, computers, digital accessories
- **服装**: Men's wear, women's wear, shoes and bags
- **图书**: Books, e-books, educational supplies
- **母婴**: Formula, toys, baby supplies

### Payment Methods

JD.com supports multiple payment methods:
- WeChat Pay
- Alipay
- JD Baitiao (JD Credit)

## Example Files

- [JD.com Shopping Example](../../example/jd/example-jd-records.csv)
- [Configuration Example](../../example/jd/config.yaml)
- [Output Example](../../example/jd/example-jd-output.beancount)
