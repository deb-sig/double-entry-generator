---
title: 账户映射  
parent: 配置指南
nav_order: 2
---

# 账户映射

合理的账户设计是复式记账的基础。本文档提供账户设置的最佳实践。

## 账户类型

### 资产账户 (Assets)

#### 现金和银行
```yaml
Assets:Cash                    # 现金
Assets:Bank:CN:ICBC           # 工商银行
Assets:Bank:CN:CCB            # 建设银行
Assets:Bank:US:Chase          # 大通银行
Assets:Bank:CA:TD             # 道明银行
```

#### 数字钱包
```yaml
Assets:Digital:Alipay:Cash    # 支付宝余额
Assets:Digital:Alipay:YuEBao  # 余额宝
Assets:Digital:WeChat:Cash    # 微信零钱
Assets:Digital:WeChat:LiCai   # 微信理财通
```

#### 投资账户
```yaml
Assets:Invest:Stocks:CN       # 中国股票
Assets:Invest:Stocks:US       # 美国股票
Assets:Invest:Fund            # 基金
Assets:Invest:Gold            # 黄金
Assets:Crypto:BTC             # 比特币
Assets:Crypto:ETH             # 以太坊
```

### 负债账户 (Liabilities)

#### 信用卡
```yaml
Liabilities:CreditCard:ICBC:1234   # 工商银行信用卡
Liabilities:CreditCard:CCB:5678    # 建设银行信用卡
```

#### 贷款
```yaml
Liabilities:Loan:Mortgage     # 房贷
Liabilities:Loan:Car          # 车贷
Liabilities:Loan:Student      # 学生贷款
```

### 支出账户 (Expenses)

#### 生活必需
```yaml
Expenses:Food:Groceries       # 日用百货
Expenses:Food:Restaurant      # 餐厅用餐
Expenses:Food:Delivery        # 外卖
Expenses:Food:Lunch           # 午餐
Expenses:Food:Dinner          # 晚餐

Expenses:Housing:Rent         # 房租
Expenses:Housing:Utilities    # 水电费
Expenses:Housing:Internet     # 网费
Expenses:Housing:Maintenance  # 房屋维护
```

#### 交通出行
```yaml
Expenses:Transport:Taxi       # 出租车/网约车
Expenses:Transport:Subway     # 地铁
Expenses:Transport:Bus        # 公交
Expenses:Transport:Gas        # 汽油费
Expenses:Transport:Flight     # 机票
```

#### 购物消费
```yaml
Expenses:Shopping:Clothing    # 服装
Expenses:Shopping:Electronics # 电子产品
Expenses:Shopping:Books       # 图书
Expenses:Shopping:Online      # 网购
```

#### 娱乐教育
```yaml
Expenses:Entertainment:Movie  # 电影
Expenses:Entertainment:Game   # 游戏
Expenses:Education:Course     # 课程费用
Expenses:Education:Books      # 教育书籍
```

#### 健康保险
```yaml
Expenses:Health:Medical       # 医疗费用
Expenses:Health:Insurance     # 保险费用
Expenses:Health:Gym          # 健身费用
```

### 收入账户 (Income)

#### 工作收入
```yaml
Income:Salary                 # 工资
Income:Bonus                  # 奖金
Income:Freelance              # 自由职业收入
```

#### 投资收入
```yaml
Income:Interest               # 利息收入
Income:Dividend               # 股息收入
Income:Investment:PnL         # 投资损益
Income:Crypto:PnL             # 加密货币损益
```

#### 其他收入
```yaml
Income:Gift                   # 礼金收入
Income:Refund                 # 退款
Income:Cashback               # 返现
```

## 账户设计原则

### 1. 层次清晰
```yaml
# 好的设计
Assets:Bank:CN:ICBC:Checking
Assets:Bank:CN:ICBC:Savings

# 不好的设计
Assets:ICBCChecking
Assets:ICBCSavings
```

### 2. 地区区分
```yaml
# 中国银行
Assets:Bank:CN:ICBC
Assets:Bank:CN:CCB

# 美国银行
Assets:Bank:US:Chase
Assets:Bank:US:BankOfAmerica

# 加拿大银行
Assets:Bank:CA:TD
Assets:Bank:CA:BMO
```

### 3. 货币标注（可选）
```yaml
Assets:Bank:CN:ICBC:CNY       # 人民币账户
Assets:Bank:US:Chase:USD      # 美元账户
Assets:Bank:HK:HSBC:HKD       # 港币账户
```

## Provider 特定建议

### 支付宝配置
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

### 微信配置
```yaml
defaultCashAccount: Assets:Digital:WeChat:Cash

# 主要用微信余额和零钱通
```

### 银行配置
```yaml
# 建设银行
defaultCashAccount: Assets:Bank:CN:CCB:Checking

# 工商银行信用卡
defaultCashAccount: Liabilities:CreditCard:ICBC:1234
```

## 常见模式

### 餐饮分类
```yaml
Expenses:Food:Breakfast       # 早餐
Expenses:Food:Lunch           # 午餐  
Expenses:Food:Dinner          # 晚餐
Expenses:Food:Snacks          # 零食
Expenses:Food:Delivery        # 外卖
Expenses:Food:Restaurant      # 餐厅
```

### 购物分类
```yaml
Expenses:Shopping:Groceries   # 日用百货
Expenses:Shopping:Clothing    # 服装
Expenses:Shopping:Electronics # 电子产品
Expenses:Shopping:Books       # 图书
Expenses:Shopping:Home        # 家居用品
```

### 投资分类
```yaml
Assets:Invest:Stocks:CN       # 中国股票
Assets:Invest:Stocks:US       # 美国股票
Assets:Invest:Fund:Index      # 指数基金
Assets:Invest:Fund:Active     # 主动基金
Assets:Invest:Bond            # 债券
Assets:Invest:Gold            # 黄金
Assets:Crypto:BTC             # 比特币
Assets:Crypto:ETH             # 以太坊
```

## 调整建议

1. **从简单开始**: 先设置大类，后续根据需要细分
2. **保持一致**: 账户命名保持一致的层次结构
3. **定期整理**: 定期检查和整理账户结构
4. **避免过度细分**: 不要创建太多很少使用的账户