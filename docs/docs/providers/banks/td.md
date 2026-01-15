---
title: 道明银行 (TD)
---


# 道明银行 (TD) Provider

TD (Toronto-Dominion Bank) Provider 支持将加拿大道明银行账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 账单下载方式

1. 登录 TD 网页版本: https://easyweb.td.com/
2. 点击指定的账户
3. 选择账单范围 -> "Select Download Format" -> Spreadsheet(.csv) -> Download

## 使用方法

### 基本命令

```bash
# 转换TD银行账单
double-entry-generator translate -p td -t beancount td_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Bank:TD
defaultPlusAccount: Assets:Bank:TD
defaultCashAccount: Assets:Bank:TD
defaultCurrency: CAD
title: TD银行账单转换
layout: default

td:
  rules:
    - peer: LOBLAWS
      targetAccount: Expenses:Groceries
    - peer: CANADIAN TIRE
      targetAccount: Expenses:Shopping
    - peer: TIM HORTONS
      targetAccount: Expenses:Food:Coffee
```

## 配置说明

### 全局配置

- 默认货币为加元 (CAD)
- 支持加拿大主要连锁店识别

### 规则配置

TD Provider 针对加拿大本土商户进行了优化，如 Loblaws、Canadian Tire、Tim Hortons 等。

## 示例文件

- [TD银行示例](../../example/td/example-td-records.csv)
- [配置示例](../../example/td/config.yaml)
- [输出示例](../../example/td/example-td-output.beancount)