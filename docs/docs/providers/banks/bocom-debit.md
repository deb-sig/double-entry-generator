---
title: 交通银行储蓄卡 (BOCOM Debit)
---


# 交通银行储蓄卡 (BOCOM Debit) Provider

交通银行储蓄卡 Provider 将交通银行 App 导出的交易明细（PDF 需先转换成 CSV）转换为 Beancount/Ledger 记账条目。

## 支持的文件格式

- CSV（通过交通银行 App 导出的 PDF 使用 [bill-file-converter](https://github.com/deb-sig/bill-file-converter) 转换得到）

## 下载方式

1. 打开交通银行 App，搜索“交易明细”
2. 点击底部“导出交易明细”，选择电子版
3. 选择卡号并设置自定义时间范围后点击“去开立”
4. 设置账单格式并全部开启，填写接收邮箱地址并确认导出
5. 将收到的 PDF 文件使用 bill-file-converter 转换为 CSV

## 使用方法

### 基本命令

```bash
double-entry-generator translate \
  --config ./example/bocom_debit/config.yaml \
  --provider bocom_debit \
  --output ./example/bocom_debit/example-bocom-debit-output.beancount \
  ./example/bocom_debit/example-bocom-debit-records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:BOCOM
defaultCurrency: CNY
title: 交通银行借记卡示例

bocom_debit:
  rules:
    - peer: 应付个人活期储蓄存款利息
      targetAccount: Income:Interest
    - peer: 信用卡还款
      targetAccount: Equity:Transfers
    - txType: 信用卡转账还款
      targetAccount: Equity:Transfers
    - peer: 网上国网
      targetAccount: Expenses:Electricity
    - item: 基金理财产品申购
      targetAccount: Assets:Funds
    - item: 财付通
      ignore: true
```

## 示例文件

- [交易明细 CSV 示例](../../example/bocom_debit/example-bocom-debit-records.csv)
- [转换后 Beancount 示例](../../example/bocom_debit/example-bocom-debit-output.beancount)
- [配置文件示例](../../example/bocom_debit/config.yaml)
