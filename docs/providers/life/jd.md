---
title: 京东 (JD)
layout: default
parent: 提供商支持
nav_order: 12
---

# 京东 (JD) Provider

京东 Provider 支持将京东购物订单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 账单下载方式

1. 打开京东手机 APP
2. 前往我的 -> 我的钱包 -> 账单
3. 点击右上角 Icon(三条横杠)
4. 选择"账单导出（仅限个人对账）"

## 使用方法

### 基本命令

```bash
# 转换京东购物记录
double-entry-generator translate -p jd -t beancount jd_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat  # 微信支付或其他支付方式
defaultCurrency: CNY
title: 京东购物转换
layout: default

jd:
  rules:
    - category: 食品饮料
      targetAccount: Expenses:Food:Groceries
      tag: food,groceries
    - category: 家居家装
      targetAccount: Expenses:Home:Furniture
      tag: home,furniture
    - category: 数码
      targetAccount: Expenses:Electronics
      tag: electronics
    - category: 服装
      targetAccount: Expenses:Clothing
      tag: clothing
    - category: 图书
      targetAccount: Expenses:Books
      tag: books,education
```

## 配置说明

### 商品分类

京东支持丰富的商品分类：
- **食品饮料**: 生鲜、零食、饮料等
- **家居家装**: 家具、家电、装修用品
- **数码**: 手机、电脑、数码配件
- **服装**: 男装、女装、鞋包
- **图书**: 图书、电子书、教育用品
- **母婴**: 奶粉、玩具、婴儿用品

### 支付方式

京东支持多种支付方式：
- 微信支付
- 支付宝
- 京东白条
- 银行卡
- 京东余额

### 订单状态

支持不同订单状态的处理：
- 已支付订单
- 退款订单
- 部分退款

## 特色功能

### 1. 智能分类
根据京东的商品分类自动归类到对应的账户。

### 2. 白条支持
支持京东白条（消费信贷）的记账处理。

### 3. 促销活动
支持优惠券、满减、秒杀等促销活动记录。

## 示例文件

- [京东购物示例](../../example/jd/example-jd-records.csv)
- [配置示例](../../example/jd/config.yaml)
- [输出示例](../../example/jd/example-jd-output.beancount)