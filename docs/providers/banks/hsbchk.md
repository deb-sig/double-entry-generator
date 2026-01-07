---
title: 汇丰银行香港 (HSBC HK)
layout: default
parent: 提供商支持
nav_order: 4
---

# 汇丰银行香港 (HSBC HK) Provider

汇丰银行香港 Provider 支持将香港汇丰银行账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 账单下载方式

1. 登录HSBC HK网上银行
2. 访问账户概览页面
3. 选择所需的账户（借记卡或信用卡）
4. 在交易明细页面，选择想要导出的时间段
5. 点击"导出"按钮，选择 CSV 格式导出

## 使用方法

### 基本命令

```bash
# 转换汇丰银行香港账单
double-entry-generator translate -p hsbchk -t beancount hsbchk_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Bank:HSBC:HK
defaultPlusAccount: Assets:Bank:HSBC:HK
defaultCashAccount: Assets:Bank:HSBC:HK
defaultCurrency: HKD
title: 汇丰银行香港账单转换
layout: default

hsbchk:
  rules:
    - peer: 支付宝
      targetAccount: Expenses:Payment:Alipay
    - peer: 八达通
      targetAccount: Expenses:Transport:Octopus
    - peer: 便利店
      targetAccount: Expenses:Food:Convenience
```

## 配置说明

支持香港常见的支付方式和商户分类，包括八达通、便利店等香港特色消费场景。

### 货币支持

默认支持港币 (HKD)，也可配置其他货币。

## 示例文件

- [汇丰银行香港借记卡示例](../../example/hsbchk/debit/example-hsbchk-debit-records.csv)
- [汇丰银行香港信用卡示例](../../example/hsbchk/credit/example-hsbchk-credit-records.csv)
- [配置示例](../../example/hsbchk/credit/config.yaml)