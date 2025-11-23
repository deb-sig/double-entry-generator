---
title: 交通银行信用卡 (BOCOM Credit)
layout: default
parent: 提供商支持
nav_order: 9
---

# 交通银行信用卡 (BOCOM Credit) Provider

交通银行信用卡 Provider 将交通银行信用卡买单吧 App 导出的交易明细（EML 需先转换成 CSV）转换为 Beancount/Ledger 记账条目。

## 支持的文件格式

- CSV（通过交通银行信用卡买单吧 App 导出的邮件 EML 使用 [bill-file-converter](https://github.com/deb-sig/bill-file-converter) 转换得到）

## 下载方式

1. 打开交通银行信用卡买单吧 App，搜索“账单补发”
2. 选择补发账单月份，点击底部“下一步”
3. 确认邮箱地址，并确认导出
5. 将收到的邮件下载为 EML 文件，并使用 bill-file-converter 转换为 CSV

## 使用方法

### 基本命令

```bash
double-entry-generator translate \
  --config ./example/bocomcredit/config.yaml \
  --provider bocomcredit \
  --output ./example/bocomcredit/example-bocomcredit-output.beancount \
  ./example/bocomcredit/example-bocomcredit-records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultCurrency: CNY
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:BOCOM:CreditCard
bocomcredit:
  rules:
    - item: 信用卡还款
      targetAccount: Equity:Transfers
    - item: 美团,饿了么
      targetAccount: Expenses:Food
```

## 示例文件

- [交易明细 CSV 示例](../../example/bocomcredit/example-bocomcredit-records.csv)
- [转换后 Beancount 示例](../../example/bocomcredit/example-bocomcredit-output.beancount)
- [配置文件示例](../../example/bocomcredit/config.yaml)
