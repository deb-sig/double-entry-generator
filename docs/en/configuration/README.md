---
title: Configuration Guide
layout: default
nav_order: 2
has_children: true
permalink: /en/configuration/
description: "Detailed configuration and rule setup guide"
lang: en
---

# Configuration Overview

Double Entry Generator uses YAML format configuration files to define conversion rules.

## Basic Structure

```yaml
# Global configuration
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:Example
defaultCurrency: CNY
title: Example Configuration
layout: default

# Provider-specific configuration
providerName:
  rules:
    - # Rule 1
    - # Rule 2
```

## Global Configuration Fields

### Required Fields

- `defaultMinusAccount`: Default account for amount decrease
- `defaultPlusAccount`: Default account for amount increase
- `defaultCurrency`: Default currency code (e.g., CNY, USD, EUR)
- `title`: Configuration file title

### Optional Fields

- `defaultCashAccount`: Default cash account (used by some providers)

## Provider Configuration

Each provider has its own configuration section, with the section name matching the provider name:

```yaml
alipay:
  rules: [...]

wechat:
  rules: [...]

ccb:
  rules: [...]
```

## Rule Priority

Rules are matched in the order they appear in the configuration file:

1. **Match from top to bottom**
2. **Later rules have higher priority**
3. **A transaction can match multiple rules**
4. **The last matched rule overrides previous settings**

## Common Configuration Patterns

### Categorize by Time

```yaml
rules:
  - category: 餐饮美食
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch
  - category: 餐饮美食
    time: "17:00-21:00"
    targetAccount: Expenses:Food:Dinner
```

### Categorize by Amount

```yaml
rules:
  - category: 日用百货
    maxPrice: 10
    targetAccount: Expenses:Food:Snacks
  - category: 日用百货
    minPrice: 10
    targetAccount: Expenses:Groceries
```

### Categorize by Merchant

```yaml
rules:
  - peer: "滴滴,高德,出租车"
    sep: ","
    targetAccount: Expenses:Transport:Taxi
  - peer: "美团,饿了么"
    sep: ","
    targetAccount: Expenses:Food:Delivery
```

## Next Steps

- [Rules Configuration Details](rules.html) - Learn how to write matching rules
- [Account Mapping](accounts.html) - Learn account setup best practices
