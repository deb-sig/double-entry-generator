---
title: Rules Configuration
---


# Rules Configuration Details

Rules are the core functionality of Double Entry Generator, used to automatically categorize transactions into different accounts.

## Rule Matching Fields

The following are common rule-matching fields. **Support varies by provider**; see each provider’s documentation for details.

- `type`: Transaction type (e.g., expense, income, other)
- `peer`: Transaction counterpart (merchant name, personal name, etc.)
- `item`: Product/service description
- `time`: Time range matching
- `minPrice`: Minimum amount
- `maxPrice`: Maximum amount

### Provider-Specific Fields

Special fields supported by different providers:

#### Alipay (alipay)
- `category`: Consumption category (餐饮美食, 日用百货, etc.)
- `method`: Payment method (余额, 余额宝, 信用卡, etc.)

#### WeChat (wechat)
- `status`: Transaction status

#### China Construction Bank (ccb)
- `txType`: Summary information
- `status`: Transaction status

## Matching Options

### Basic Options

```yaml
- peer: "美团"
  fullMatch: true        # Exact match, default is false
  targetAccount: Expenses:Food
```

### Multi-Keyword Matching

```yaml
- peer: "美团,饿了么,肯德基"
  sep: ","              # Separator, default is ","
  targetAccount: Expenses:Food
```

### Time Range Matching

```yaml
- category: 餐饮美食
  time: "11:00-14:00"   # 11:00 to 14:00
  targetAccount: Expenses:Food:Lunch
```

### Amount Range Matching

```yaml
- category: 日用百货
  minPrice: 0           # Minimum amount
  maxPrice: 10          # Maximum amount
  targetAccount: Expenses:Food:Snacks
```

## Account Settings

### Basic Account Settings

```yaml
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi    # Target account
  methodAccount: Assets:Alipay             # Payment account (optional)
```

### Investment Transactions

```yaml
- peer: "基金"
  type: "其他"
  item: "买入"
  targetAccount: Assets:Alipay:Invest:Fund
  pnlAccount: Income:Alipay:Invest:PnL     # Profit and loss account
```

## Special Features

### Ignore Transactions

```yaml
- peer: "财付通"
  ignore: true          # Ignore matched transactions
```

### Add Tags

Alipay uses `tags`; WeChat uses `tag`. Both are optional; `sep` is the separator.

```yaml
# Alipay
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi
  tags: "transport,taxi"
  sep: ","

# WeChat
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi
  tag: "transport,taxi"
  sep: ","
```

## Rule Priority and Override

### Priority Rules

1. Rules are matched in the order they appear in the configuration file
2. Later rules override previous settings
3. A transaction can match multiple rules

### Example

Rules are listed under the provider’s `rules` key (e.g. `alipay.rules`, `wechat.rules`):

```yaml
alipay:
  rules:
    - peer: "美团"
      targetAccount: Expenses:Food
    - peer: "美团"
      time: "11:00-14:00"
      targetAccount: Expenses:Food:Lunch  # Later rule overrides
```

## Best Practices

### 1. From General to Specific

```yaml
alipay:
  rules:
    - category: 餐饮美食
      targetAccount: Expenses:Food
    - category: 餐饮美食
      time: "11:00-14:00"
      targetAccount: Expenses:Food:Lunch
```

### 2. Use Multiple Keywords

Use one of the match fields (e.g. `peer`) with multiple values separated by `sep`; a transaction matching any of them triggers the rule:

```yaml
- peer: "美团,饿了么,肯德基,麦当劳"
  sep: ","
  targetAccount: Expenses:Food:FastFood
```

### 3. Set Priority Reasonably

Place special rules after general rules.

## Debugging Tips

### 1. Add Rules Gradually
Start with simple rules and gradually add complex ones.

### 2. Use Tags for Tracking
Add tags to important rules for easier analysis later.

### 3. Regularly Check Output
Regularly check the generated ledger and adjust rule settings.
