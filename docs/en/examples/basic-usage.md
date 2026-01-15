---
title: Basic Usage Examples
layout: default
parent: Examples
nav_order: 1
description: "Learn how to use Double Entry Generator through practical examples"
permalink: /en/examples/basic-usage/
lang: en
---

# Basic Usage Examples

This document demonstrates how to use Double Entry Generator through practical examples.

## Scenario 1: Alipay Bill Conversion

### 1. Prepare Bill File

Download CSV format bill file from Alipay: `alipay_202501.csv`

### 2. Create Configuration File

Create `alipay_config.yaml`:

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: My Alipay Bills
layout: default

alipay:
  rules:
    # Categorize food by time
    - category: 餐饮美食
      time: "07:00-11:00"
      targetAccount: Expenses:Food:Breakfast
    - category: 餐饮美食
      time: "11:00-15:00"
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: "17:00-22:00"
      targetAccount: Expenses:Food:Dinner
    
    # Transportation
    - peer: "滴滴出行,高德打车"
      sep: ","
      targetAccount: Expenses:Transport:Taxi
    
    # Online shopping
    - peer: "天猫,京东"
      sep: ","
      targetAccount: Expenses:Shopping:Online
    
    # Payment methods
    - method: "余额宝"
      methodAccount: Assets:Alipay:YuEBao
    - method: "余额"
      methodAccount: Assets:Alipay:Cash
```

### 3. Execute Conversion

```bash
double-entry-generator translate \
  --provider alipay \
  --target beancount \
  --config alipay_config.yaml \
  --output my_alipay.beancount \
  alipay_202501.csv
```

### 4. View Results

Generated `my_alipay.beancount` file:

```beancount
option "title" "My Alipay Bills"
option "operating_currency" "CNY"

1970-01-01 open Assets:Alipay:Cash
1970-01-01 open Assets:Alipay:YuEBao
1970-01-01 open Expenses:Food:Lunch
1970-01-01 open Expenses:Transport:Taxi

2025-01-15 * "滴滴出行" "快车" 
    Expenses:Transport:Taxi     23.50 CNY
    Assets:Alipay:Cash         -23.50 CNY

2025-01-15 * "某餐厅" "午餐" 
    Expenses:Food:Lunch         35.00 CNY
    Assets:Alipay:YuEBao       -35.00 CNY
```

## Scenario 2: WeChat Bill Conversion (XLSX)

### 1. Prepare Files

Download XLSX format bill from WeChat: `wechat_202501.xlsx`

### 2. Configuration File

Create `wechat_config.yaml`:

```yaml
defaultMinusAccount: Assets:WeChat:Cash
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: My WeChat Bills

wechat:
  rules:
    - category: 餐饮美食
      targetAccount: Expenses:Food
    - peer: "滴滴"
      targetAccount: Expenses:Transport:Taxi
```

### 3. Execute Conversion

```bash
double-entry-generator translate \
  --provider wechat \
  --target beancount \
  --config wechat_config.yaml \
  --output my_wechat.beancount \
  wechat_202501.xlsx
```

## Scenario 3: Bank Statement Conversion

### 1. Prepare Bank Statement

Download bank statement file: `ccb_202501.xls`

### 2. Configuration File

Create `ccb_config.yaml`:

```yaml
defaultMinusAccount: Assets:Bank:CCB
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: My CCB Bills

ccb:
  rules:
    - txType: "消费"
      targetAccount: Expenses:FIXME
```

### 3. Execute Conversion

```bash
double-entry-generator translate \
  --provider ccb \
  --target beancount \
  --config ccb_config.yaml \
  --output my_ccb.beancount \
  ccb_202501.xls
```

## Scenario 4: Merging Multiple Bills

### 1. Convert Each Platform's Bills Separately

```bash
# Convert Alipay
double-entry-generator translate -p alipay -c alipay_config.yaml -o alipay.beancount alipay_202501.csv

# Convert WeChat
double-entry-generator translate -p wechat -c wechat_config.yaml -o wechat.beancount wechat_202501.xlsx

# Convert CCB
double-entry-generator translate -p ccb -c ccb_config.yaml -o ccb.beancount ccb_202501.xls
```

### 2. Merge Files

Create main ledger file `main.beancount`:

```beancount
option "title" "My Personal Ledger"
option "operating_currency" "CNY"

; Include bills from each platform
include "alipay.beancount"
include "wechat.beancount" 
include "ccb.beancount"

; Opening balances
1970-01-01 open Equity:Opening-Balances

2025-01-01 pad Assets:Alipay:Cash Equity:Opening-Balances
2025-01-01 pad Assets:WeChat:Cash Equity:Opening-Balances
2025-01-01 pad Assets:Bank:CCB Equity:Opening-Balances
```

### 3. Verify Ledger

```bash
# Verify using beancount
bean-check main.beancount

# Generate reports
bean-report main.beancount balances
```

## Common Issues and Solutions

### Issue 1: Account Names Not Standardized

**Problem**: Output account names contain FIXME

**Solution**: Check `defaultMinusAccount` and `defaultPlusAccount` settings in the configuration file

### Issue 2: Transactions Not Correctly Categorized

**Solution**:
1. Check if rule matching fields are correct
2. Pay attention to rule priority order
3. Use more specific matching conditions

### Issue 3: Duplicate Bookkeeping

**Solution**:
1. Set `ignore: true` for bank-side transactions from third-party payment platforms
2. Avoid importing the same transaction from multiple platforms

## Next Steps

- [Advanced Rules Configuration](../advanced-rules/) - Learn more complex rule configuration techniques
- [Configuration Overview](../../configuration/) - Understand detailed configuration file instructions
