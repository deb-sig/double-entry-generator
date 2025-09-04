# 支付宝 (Alipay) Provider

支付宝 Provider 支持将支付宝账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换支付宝账单
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# 指定配置文件
double-entry-generator translate -p alipay -t beancount -c config.yaml alipay_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 支付宝账单转换

alipay:
  rules:
    # 收入类交易
    - type: 收入
      item: 商品
      targetAccount: Income:Alipay:ShouKuanMa
      methodAccount: Assets:Alipay
    
    # 按类别分类消费
    - category: 日用百货
      minPrice: 10
      targetAccount: Expenses:Groceries
    - category: 日用百货
      maxPrice: 9.99
      targetAccount: Expenses:Food:Drink
    
    # 按时间分类餐饮
    - category: 餐饮美食
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: 16:00-22:00
      targetAccount: Expenses:Food:Dinner
    
    # 按商家匹配
    - peer: 滴滴出行
      targetAccount: Expenses:Transport
    - peer: 苏宁
      targetAccount: Expenses:Electronics
    
    # 支付方式匹配
    - method: 余额
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 余额宝
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 交通银行信用卡(7449)
      fullMatch: true
      methodAccount: Liabilities:CC:COMM:7449
    
    # 投资相关
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

## 配置说明

### 全局配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCurrency`: 默认货币

### 规则配置

支付宝 Provider 提供基于规则的匹配，可以指定：

- `type`（交易类型）- 支出/收入/其他
- `category`（消费类别）- 餐饮美食/日用百货/交通出行等
- `peer`（交易对方）- 商家名称
- `item`（商品说明）- 具体商品描述
- `method`（支付方式）- 余额/余额宝/信用卡等
- `time`（交易时间）- 时间区间匹配
- `minPrice`/`maxPrice` - 金额区间匹配

### 规则选项

- `sep`: 分隔符，默认为 `,`，用于多关键字匹配
- `fullMatch`: 是否使用完全匹配，默认为 `false`
- `tag`: 设置流水的 Tag
- `methodAccount`: 指定支付账户（如余额、信用卡等）
- `targetAccount`: 指定目标账户
- `pnlAccount`: 投资收益账户（用于基金、黄金交易）

## 账户关系

支付宝的账户关系相对复杂，因为涉及多种支付方式：

### 基本消费
- **支出交易**: `methodAccount` → `targetAccount`
- **收入交易**: `targetAccount` → `methodAccount`

### 投资交易
- **买入**: `methodAccount` → `targetAccount`
- **卖出**: `targetAccount` → `methodAccount` + `pnlAccount`（损益）

## 特色功能

### 1. 支付方式识别
支付宝账单包含详细的支付方式信息，可以精确映射到不同账户：
- 余额 → `Assets:Alipay`
- 余额宝 → `Assets:Alipay:YuEBao`
- 信用卡 → `Liabilities:CC:Bank:Number`

### 2. 时间段分类
支持按时间自动分类餐饮消费：
```yaml
- category: 餐饮美食
  time: 11:00-14:00
  targetAccount: Expenses:Food:Lunch
```

### 3. 投资交易支持
内置支持基金和黄金交易的损益记账。

## 示例文件

- [支付宝账单示例](../../example/alipay/example-alipay-records.csv)
- [配置示例](../../example/alipay/config.yaml)
- [输出示例](../../example/alipay/example-alipay-output.beancount)