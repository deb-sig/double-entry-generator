---
title: 基本使用示例
layout: default
parent: 示例
nav_order: 1
description: "通过实际例子展示如何使用 Double Entry Generator"
permalink: /examples/basic-usage/
---

# 基本使用示例

本文档通过实际例子展示如何使用 Double Entry Generator。

## 场景1：支付宝账单转换

### 1. 准备账单文件

从支付宝下载 CSV 格式的账单文件：`alipay_202501.csv`

### 2. 创建配置文件

创建 `alipay_config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 我的支付宝账单
layout: default

alipay:
  rules:
    # 餐饮按时间分类
    - category: 餐饮美食
      time: "07:00-11:00"
      targetAccount: Expenses:Food:Breakfast
    - category: 餐饮美食
      time: "11:00-15:00"
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: "17:00-22:00"
      targetAccount: Expenses:Food:Dinner
    
    # 交通出行
    - peer: "滴滴出行,高德打车"
      sep: ","
      targetAccount: Expenses:Transport:Taxi
    
    # 网购
    - peer: "天猫,京东"
      sep: ","
      targetAccount: Expenses:Shopping:Online
    
    # 支付方式
    - method: "余额宝"
      methodAccount: Assets:Alipay:YuEBao
    - method: "余额"
      methodAccount: Assets:Alipay:Cash
```

### 3. 执行转换

```bash
double-entry-generator translate \
  --provider alipay \
  --target beancount \
  --config alipay_config.yaml \
  --output my_alipay.beancount \
  alipay_202501.csv
```

### 4. 查看结果

生成的 `my_alipay.beancount` 文件：

```beancount
option "title" "我的支付宝账单"
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

## 场景2：微信账单转换（XLSX）

### 1. 准备文件

从微信下载 XLSX 格式的账单：`wechat_202501.xlsx`

### 2. 配置文件

创建 `wechat_config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat:Cash
defaultCurrency: CNY
title: 我的微信账单
layout: default

wechat:
  rules:
    # 外卖
    - item: "美团,饿了么"
      sep: ","
      targetAccount: Expenses:Food:Delivery
      tag: "food,delivery"
    
    # 生活缴费
    - item: "电费,水费,燃气费"
      sep: ","
      targetAccount: Expenses:Utilities
      tag: "utilities"
    
    # 红包收入
    - type: "收入"
      item: "微信红包"
      targetAccount: Income:Gift
```

### 3. 执行转换

```bash
double-entry-generator translate \
  --provider wechat \
  --target beancount \
  --config wechat_config.yaml \
  --output my_wechat.beancount \
  wechat_202501.xlsx
```

## 场景3：建设银行账单转换

### 1. 准备文件

从建设银行网银下载 XLS 文件：`ccb_202501.xls`

### 2. 配置文件

创建 `ccb_config.yaml`：

```yaml
defaultMinusAccount: Assets:Bank:CCB
defaultPlusAccount: Assets:Bank:CCB
defaultCashAccount: Assets:Bank:CCB
defaultCurrency: CNY
title: 建设银行账单
layout: default

ccb:
  rules:
    # 第三方支付平台 - 忽略避免重复记账
    - peer: "支付宝,微信支付,财付通"
      sep: ","
      ignore: true
    
    # ATM取现
    - txType: "ATM取现"
      targetAccount: Assets:Cash
    
    # 工资收入
    - peer: "我的公司"
      type: "收入"
      targetAccount: Income:Salary
    
    # 信用卡还款
    - peer: "信用卡中心"
      targetAccount: Liabilities:CreditCard:CCB
```

### 3. 执行转换

```bash
double-entry-generator translate \
  --provider ccb \
  --target beancount \
  --config ccb_config.yaml \
  --output my_ccb.beancount \
  ccb_202501.xls
```

## 场景4：合并多个账单

### 1. 分别转换各平台账单

```bash
# 转换支付宝
double-entry-generator translate -p alipay -c alipay_config.yaml -o alipay.beancount alipay_202501.csv

# 转换微信
double-entry-generator translate -p wechat -c wechat_config.yaml -o wechat.beancount wechat_202501.xlsx

# 转换建设银行
double-entry-generator translate -p ccb -c ccb_config.yaml -o ccb.beancount ccb_202501.xls
```

### 2. 合并文件

创建主账本文件 `main.beancount`：

```beancount
option "title" "我的个人账本"
option "operating_currency" "CNY"

; 引入各平台账单
include "alipay.beancount"
include "wechat.beancount" 
include "ccb.beancount"

; 期初余额
1970-01-01 open Equity:Opening-Balances

2025-01-01 pad Assets:Alipay:Cash Equity:Opening-Balances
2025-01-01 pad Assets:WeChat:Cash Equity:Opening-Balances
2025-01-01 pad Assets:Bank:CCB Equity:Opening-Balances
```

### 3. 验证账本

```bash
# 使用 beancount 验证
bean-check main.beancount

# 生成报表
bean-report main.beancount balances
```

## 常见问题和解决方案

### 问题1：账户名称不规范

**问题**：输出的账户名称包含 FIXME

**解决**：检查配置文件中的 defaultMinusAccount 和 defaultPlusAccount 设置

### 问题2：交易没有正确分类

**解决**：
1. 检查规则的匹配字段是否正确
2. 注意规则的优先级顺序
3. 使用更具体的匹配条件

### 问题3：重复记账

**解决**：
1. 对第三方支付平台的银行端交易设置 `ignore: true`
2. 避免同一笔交易在多个平台重复导入

## 下一步

- [高级规则配置](../advanced-rules/) - 学习更复杂的规则配置技巧
- [配置总览](../../configuration/) - 了解配置文件的详细说明