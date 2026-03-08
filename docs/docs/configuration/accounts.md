---
title: 账户映射
description: 账户设置与最佳实践
---

# 账户映射

良好的账户结构是复式记账的基础。本文说明账户设置的最佳实践。

## 账户类型

### 资产 (Assets)

#### 现金与银行
```yaml
Assets:Cash                    # 现金
Assets:Bank:CN:ICBC           # 工商银行
Assets:Bank:CN:CCB            # 建设银行
Assets:Bank:US:Chase          # Chase
Assets:Bank:CA:TD             # 道明银行
```

#### 电子钱包
```yaml
Assets:Digital:Alipay:Cash    # 支付宝余额
Assets:Digital:Alipay:YuEBao  # 余额宝
Assets:Digital:WeChat:Cash    # 微信零钱
Assets:Digital:WeChat:LiCai   # 微信理财通
```

#### 投资账户
```yaml
Assets:Invest:Stocks:CN       # A 股
Assets:Invest:Stocks:US       # 美股
Assets:Invest:Fund            # 基金
Assets:Invest:Gold            # 黄金
Assets:Crypto:BTC             # 比特币
Assets:Crypto:ETH             # 以太坊
```

### 负债 (Liabilities)

#### 信用卡
```yaml
Liabilities:CreditCard:ICBC:1234   # 工商银行信用卡
Liabilities:CreditCard:CCB:5678    # 建设银行信用卡
```

#### 贷款
```yaml
Liabilities:Loan:Mortgage     # 房贷
Liabilities:Loan:Car          # 车贷
Liabilities:Loan:Student      # 助学贷
```

### 支出 (Expenses)

#### 日常
```yaml
Expenses:Food:Groceries       # 日用/超市
Expenses:Food:Restaurant      # 餐饮
Expenses:Food:Delivery        # 外卖
Expenses:Food:Lunch           # 午餐
Expenses:Food:Dinner          # 晚餐

Expenses:Housing:Rent         # 房租
Expenses:Housing:Utilities    # 水电
Expenses:Housing:Internet     # 网络
Expenses:Housing:Maintenance  # 维修
```

#### 交通
```yaml
Expenses:Transport:Taxi       # 打车
Expenses:Transport:Subway     # 地铁
Expenses:Transport:Bus        # 公交
Expenses:Transport:Gas        # 油费
Expenses:Transport:Flight     # 机票
```

#### 购物
```yaml
Expenses:Shopping:Clothing    # 服饰
Expenses:Shopping:Electronics # 数码
Expenses:Shopping:Books       # 图书
Expenses:Shopping:Online      # 网购
```

#### 娱乐与教育
```yaml
Expenses:Entertainment:Movie  # 电影
Expenses:Entertainment:Game   # 游戏
Expenses:Education:Course     # 课程
Expenses:Education:Books      # 教材
```

#### 健康与保险
```yaml
Expenses:Health:Medical       # 医疗
Expenses:Health:Insurance     # 保险
Expenses:Health:Gym          # 健身
```

### 收入 (Income)

#### 工作收入
```yaml
Income:Salary                 # 工资
Income:Bonus                  # 奖金
Income:Freelance              # 兼职
```

#### 投资收入
```yaml
Income:Interest               # 利息
Income:Dividend               # 股息
Income:Investment:PnL         # 投资盈亏
Income:Crypto:PnL             # 加密货币盈亏
```

#### 其他收入
```yaml
Income:Gift                   # 礼金
Income:Refund                 # 退款
Income:Cashback               # 返现
```

## 设计原则

### 1. 层级清晰
```yaml
# 推荐
Assets:Bank:CN:ICBC:Checking
Assets:Bank:CN:ICBC:Savings

# 不推荐
Assets:ICBCChecking
Assets:ICBCSavings
```

### 2. 区分地区
```yaml
# 国内银行
Assets:Bank:CN:ICBC
Assets:Bank:CN:CCB

# 美国银行
Assets:Bank:US:Chase

# 加拿大银行
Assets:Bank:CA:TD
Assets:Bank:CA:BMO
```

### 3. 可选货币标注
```yaml
Assets:Bank:CN:ICBC:CNY       # 人民币
Assets:Bank:US:Chase:USD      # 美元
Assets:Bank:HK:HSBC:HKD       # 港币
```

## 各 Provider 建议

### 支付宝
```yaml
defaultCashAccount: Assets:Digital:Alipay:Cash

# 支付方式映射
- method: 余额
  methodAccount: Assets:Digital:Alipay:Cash
- method: 余额宝
  methodAccount: Assets:Digital:Alipay:YuEBao
- method: 工商银行(1234)
  methodAccount: Assets:Bank:CN:ICBC:1234
```

### 微信
```yaml
defaultCashAccount: Assets:Digital:WeChat:Cash
# 主要使用零钱与理财通
```

### 银行
```yaml
# 建设银行
defaultCashAccount: Assets:Bank:CN:CCB:Checking

# 工商银行信用卡
defaultCashAccount: Liabilities:CreditCard:ICBC:1234
```

## 调整建议

1. **从简开始**：先设大类，再按需细化
2. **命名一致**：账户层级与命名风格保持一致
3. **定期整理**：定期检查、合并或拆分账户
4. **避免过细**：少建几乎用不到的账户
