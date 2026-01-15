---
title: Rules Configuration
---


# Rules Configuration Details

Rules are the core functionality of Double Entry Generator, used to automatically categorize transactions into different accounts.

## Rule Matching Fields

### Common Fields

Fields supported by all providers:

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

```yaml
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi
  tag: "transport,taxi"  # Add tags
  sep: ","              # Tag separator
```

## Rule Priority and Override

### Priority Rules

1. Rules are matched in the order they appear in the configuration file
2. Later rules override previous settings
3. A transaction can match multiple rules

### Example

```yaml
rules:
  # General rule (lower priority)
  - peer: "美团"
    targetAccount: Expenses:Food
  
  # Specific time rule (higher priority)
  - peer: "美团"
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch  # Overrides the above setting
```

## Best Practices

### 1. From General to Specific

```yaml
rules:
  # Set general category first
  - category: 餐饮美食
    targetAccount: Expenses:Food
  
  # Then refine specific cases
  - category: 餐饮美食
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch
```

### 2. Use Multiple Keywords

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
