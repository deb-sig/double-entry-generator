---
title: Meituan (MT)
---


# Meituan (MT) Provider

The MT Provider supports converting Meituan delivery and dine-in consumption records to Beancount/Ledger format.

## Supported File Formats

- CSV format

## Usage

### Basic Command

```bash
# Convert Meituan consumption records
double-entry-generator translate -p mt -t beancount mt_records.csv
```

### Configuration File

Create configuration file `config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat  # Usually paid via WeChat Pay
defaultCurrency: CNY
title: Meituan Consumption Conversion
layout: default

mt:
  rules:
    - type: 外卖订单
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
      tag: food,delivery,lunch
    - type: 外卖订单
      time: 17:00-21:00
      targetAccount: Expenses:Food:Dinner
      tag: food,delivery,dinner
    - type: 到店消费
      category: 餐饮
      targetAccount: Expenses:Food:Restaurant
      tag: food,restaurant
    - type: 到店消费
      category: 休闲娱乐
      targetAccount: Expenses:Entertainment
      tag: entertainment
```

## Configuration Explanation

### Consumption Types

Meituan supports various consumption scenarios:
- **外卖订单**: Meituan delivery
- **到店消费**: Offline consumption paid via Meituan
- **优惠券**: Promotional activity records
- **退款**: Order refund records

### Time-Based Categorization

Supports automatic food categorization by time:
- 11:00-14:00 → Lunch
- 17:00-21:00 → Dinner
- Other times → General food

### Account Settings

- `Assets:WeChat`: WeChat Pay account (main payment method for Meituan)
- `Expenses:Food:*`: Food-related expense accounts
- `Expenses:Entertainment`: Entertainment consumption account

## Special Features

### 1. Automatic Time Categorization
Automatically determines whether it's lunch or dinner based on consumption time.

### 2. Dine-in vs Delivery Distinction
Can distinguish between delivery and dine-in consumption for separate accounting.

### 3. Promotional Records
Supports records of coupons, discounts, and other promotional activities.

## Example Files

- [Meituan Consumption Example](../../example/mt/example-mt-output.bean)
- [Configuration Example](../../example/mt/config.yaml)
