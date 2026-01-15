---
title: 美团 (MT)
---


# 美团 (MT) Provider

美团 Provider 支持将美团外卖和到店消费记录转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换美团消费记录
double-entry-generator translate -p mt -t beancount mt_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat  # 通常通过微信支付
defaultCurrency: CNY
title: 美团消费转换
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

## 配置说明

### 消费类型

美团支持多种消费场景：
- **外卖订单**: 美团外卖配送
- **到店消费**: 美团支付的线下消费
- **优惠券**: 优惠活动记录
- **退款**: 订单退款记录

### 时间分类

支持按时间自动分类餐饮消费：
- 11:00-14:00 → 午餐
- 17:00-21:00 → 晚餐
- 其他时间 → 一般餐饮

### 账户设置

- `Assets:WeChat`: 微信支付账户（美团主要支付方式）
- `Expenses:Food:*`: 餐饮相关支出账户
- `Expenses:Entertainment`: 娱乐消费账户

## 特色功能

### 1. 自动时间分类
根据消费时间自动判断是午餐还是晚餐。

### 2. 到店vs外卖区分  
可以区分外卖配送和到店消费，便于分别记账。

### 3. 优惠记录
支持优惠券、满减等促销活动的记录。

## 示例文件

- [美团消费示例](../../example/mt/example-mt-output.bean)
- [配置示例](../../example/mt/config.yaml)