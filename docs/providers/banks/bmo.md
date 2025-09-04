---
title: 蒙特利尔银行 (BMO)
parent: 提供商支持
nav_order: 5
---

# 加拿大银行 (BMO) Provider

BMO (Bank of Montreal) Provider 支持将加拿大蒙特利尔银行账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换BMO银行账单
double-entry-generator translate -p bmo -t beancount bmo_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Bank:BMO
defaultPlusAccount: Assets:Bank:BMO
defaultCashAccount: Assets:Bank:BMO
defaultCurrency: CAD
title: BMO银行账单转换

bmo:
  rules:
    - peer: GROCERY
      targetAccount: Expenses:Groceries
    - peer: GAS STATION
      targetAccount: Expenses:Transport:Gas
    - peer: RESTAURANT
      targetAccount: Expenses:Food:Restaurant
```

## 配置说明

### 全局配置

- 默认货币为加元 (CAD)
- 支持加拿大常见的商户类型识别

### 规则配置

BMO Provider 基于北美银行账单格式，支持英文商户名称匹配。

## 示例文件

- [BMO借记卡示例](../../example/bmo/debit/example-bmo-records.csv)
- [BMO信用卡示例](../../example/bmo/credit/example-bmo-records.csv)
- [配置示例](../../example/bmo/credit/config.yaml)