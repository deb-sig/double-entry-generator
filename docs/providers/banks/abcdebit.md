---
title: 中国农业银行储蓄卡 (ABC Debit)
layout: default
parent: 提供商支持
nav_order: 10
---

# 中国农业银行储蓄卡 (ABC Debit) Provider

中国农业银行储蓄卡 Provider 支持将中国农业银行 App 导出的储蓄卡交易明细 CSV 转换为 Beancount/Ledger 记账条目。

## 支持的文件格式

- CSV（通过中国农业银行储蓄卡 App 导出的 PDF 使用 [bill-file-converter](https://github.com/deb-sig/bill-file-converter) 转换得到）

## 下载方式
1. 打开中国农业银行 App，进入首页“我的账户”
2. 点击借记卡“明细查询”
3. 点击“明细查询”页面右上方导出按钮
3. 完善账户、币种、时间和邮箱表单项，点击确定
5. 将收到的 PDF 使用 bill-file-converter 转换为 CSV

## 使用方法

### Beancount

```bash
double-entry-generator translate \
  --config ./example/abcdebit/config.yaml \
  --provider abcdebit \
  --output ./example/abcdebit/example-abcdebit-output.beancount \
  ./example/abcdebit/example-abcdebit-records.csv
```

### Ledger

```bash
double-entry-generator translate \
  --config ./example/abcdebit/config.yaml \
  --provider abcdebit \
  --target ledger \
  --output ./example/abcdebit/example-abcdebit-output.ledger \
  ./example/abcdebit/example-abcdebit-records.csv
```

## 配置文件示例

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:ABC:DebitCard
defaultCurrency: CNY
title: abcdebit
abcdebit:
  rules:
    - item: 转存
      targetAccount: Equity:Transfers
    - item: 财付通
      targetAccount: Expenses:Transport:Transit
      tag: transport
    - item: 正常还款
      targetAccount: Liabilities:Loans:Personal
    - item: 结息
      targetAccount: Income:Interest
    - item: 利息税
      targetAccount: Expenses:Tax
```

## 示例文件

- [交易明细 CSV 示例](../../example/abcdebit/example-abcdebit-records.csv)
- [转换后 Beancount 示例](../../example/abcdebit/example-abcdebit-output.beancount)
- [转换后 Ledger 示例](../../example/abcdebit/example-abcdebit-output.ledger)
- [配置文件示例](../../example/abcdebit/config.yaml)
