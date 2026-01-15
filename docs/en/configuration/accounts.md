---
title: Account Mapping
layout: default
parent: Configuration Guide
nav_order: 2
lang: en
---

# Account Mapping

A well-designed account structure is the foundation of double-entry bookkeeping. This document provides best practices for account setup.

## Account Types

### Asset Accounts (Assets)

#### Cash and Banks
```yaml
Assets:Cash                    # Cash
Assets:Bank:CN:ICBC           # Industrial and Commercial Bank of China
Assets:Bank:CN:CCB            # China Construction Bank
Assets:Bank:US:Chase          # Chase Bank
Assets:Bank:CA:TD             # Toronto-Dominion Bank
```

#### Digital Wallets
```yaml
Assets:Digital:Alipay:Cash    # Alipay balance
Assets:Digital:Alipay:YuEBao  # Yu'e Bao
Assets:Digital:WeChat:Cash    # WeChat balance
Assets:Digital:WeChat:LiCai   # WeChat Wealth Management
```

#### Investment Accounts
```yaml
Assets:Invest:Stocks:CN       # Chinese stocks
Assets:Invest:Stocks:US       # US stocks
Assets:Invest:Fund            # Funds
Assets:Invest:Gold            # Gold
Assets:Crypto:BTC             # Bitcoin
Assets:Crypto:ETH             # Ethereum
```

### Liability Accounts (Liabilities)

#### Credit Cards
```yaml
Liabilities:CreditCard:ICBC:1234   # ICBC credit card
Liabilities:CreditCard:CCB:5678    # CCB credit card
```

#### Loans
```yaml
Liabilities:Loan:Mortgage     # Mortgage
Liabilities:Loan:Car          # Car loan
Liabilities:Loan:Student      # Student loan
```

### Expense Accounts (Expenses)

#### Daily Necessities
```yaml
Expenses:Food:Groceries       # Groceries
Expenses:Food:Restaurant      # Restaurant dining
Expenses:Food:Delivery        # Food delivery
Expenses:Food:Lunch           # Lunch
Expenses:Food:Dinner          # Dinner

Expenses:Housing:Rent         # Rent
Expenses:Housing:Utilities    # Utilities
Expenses:Housing:Internet     # Internet
Expenses:Housing:Maintenance  # Home maintenance
```

#### Transportation
```yaml
Expenses:Transport:Taxi       # Taxi/ride-hailing
Expenses:Transport:Subway     # Subway
Expenses:Transport:Bus        # Bus
Expenses:Transport:Gas        # Gas
Expenses:Transport:Flight     # Flight tickets
```

#### Shopping
```yaml
Expenses:Shopping:Clothing    # Clothing
Expenses:Shopping:Electronics # Electronics
Expenses:Shopping:Books       # Books
Expenses:Shopping:Online      # Online shopping
```

#### Entertainment & Education
```yaml
Expenses:Entertainment:Movie  # Movies
Expenses:Entertainment:Game   # Games
Expenses:Education:Course     # Course fees
Expenses:Education:Books      # Educational books
```

#### Health & Insurance
```yaml
Expenses:Health:Medical       # Medical expenses
Expenses:Health:Insurance     # Insurance
Expenses:Health:Gym          # Gym fees
```

### Income Accounts (Income)

#### Work Income
```yaml
Income:Salary                 # Salary
Income:Bonus                  # Bonus
Income:Freelance              # Freelance income
```

#### Investment Income
```yaml
Income:Interest               # Interest income
Income:Dividend               # Dividend income
Income:Investment:PnL         # Investment profit and loss
Income:Crypto:PnL             # Cryptocurrency profit and loss
```

#### Other Income
```yaml
Income:Gift                   # Gift money
Income:Refund                 # Refund
Income:Cashback               # Cashback
```

## Account Design Principles

### 1. Clear Hierarchy
```yaml
# Good design
Assets:Bank:CN:ICBC:Checking
Assets:Bank:CN:ICBC:Savings

# Poor design
Assets:ICBCChecking
Assets:ICBCSavings
```

### 2. Regional Distinction
```yaml
# Chinese banks
Assets:Bank:CN:ICBC
Assets:Bank:CN:CCB

# US banks
Assets:Bank:US:Chase
Assets:Bank:US:BankOfAmerica

# Canadian banks
Assets:Bank:CA:TD
Assets:Bank:CA:BMO
```

### 3. Currency Annotation (Optional)
```yaml
Assets:Bank:CN:ICBC:CNY       # CNY account
Assets:Bank:US:Chase:USD      # USD account
Assets:Bank:HK:HSBC:HKD       # HKD account
```

## Provider-Specific Recommendations

### Alipay Configuration
```yaml
defaultCashAccount: Assets:Digital:Alipay:Cash

# Payment method mapping
- method: 余额
  methodAccount: Assets:Digital:Alipay:Cash
- method: 余额宝
  methodAccount: Assets:Digital:Alipay:YuEBao
- method: 工商银行(1234)
  methodAccount: Assets:Bank:CN:ICBC:1234
```

### WeChat Configuration
```yaml
defaultCashAccount: Assets:Digital:WeChat:Cash

# Mainly uses WeChat balance and WeChat Wealth Management
```

### Bank Configuration
```yaml
# China Construction Bank
defaultCashAccount: Assets:Bank:CN:CCB:Checking

# ICBC Credit Card
defaultCashAccount: Liabilities:CreditCard:ICBC:1234
```

## Common Patterns

### Food Categories
```yaml
Expenses:Food:Breakfast       # Breakfast
Expenses:Food:Lunch           # Lunch
Expenses:Food:Dinner          # Dinner
Expenses:Food:Snacks          # Snacks
Expenses:Food:Delivery        # Delivery
Expenses:Food:Restaurant      # Restaurant
```

### Shopping Categories
```yaml
Expenses:Shopping:Groceries   # Groceries
Expenses:Shopping:Clothing    # Clothing
Expenses:Shopping:Electronics # Electronics
Expenses:Shopping:Books       # Books
Expenses:Shopping:Home        # Home goods
```

### Investment Categories
```yaml
Assets:Invest:Stocks:CN       # Chinese stocks
Assets:Invest:Stocks:US       # US stocks
Assets:Invest:Fund:Index      # Index funds
Assets:Invest:Fund:Active     # Active funds
Assets:Invest:Bond            # Bonds
Assets:Invest:Gold            # Gold
Assets:Crypto:BTC             # Bitcoin
Assets:Crypto:ETH             # Ethereum
```

## Adjustment Recommendations

1. **Start Simple**: Set general categories first, then refine as needed
2. **Stay Consistent**: Keep consistent hierarchical structure in account naming
3. **Regular Organization**: Regularly check and organize account structure
4. **Avoid Over-Segmentation**: Don't create too many rarely-used accounts
