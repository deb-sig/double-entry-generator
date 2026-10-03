---
title: 广发银行信用卡 (CGB Credit)
---

# 广发银行信用卡 (CGB Credit) Provider

广发银行信用卡 Provider 将广发信用卡 PDF 账单转换后的 CSV 明细转换为 Beancount/Ledger 记账条目。

## 支持的文件格式

- CSV（通过 [bill-file-converter](https://github.com/deb-sig/bill-file-converter) 将广发信用卡 PDF 转换得到）

## 使用方法

```bash
double-entry-generator translate \
  --config ./example/cgb_credit/config.yaml \
  --provider cgb_credit \
  --output ./example/cgb_credit/example-cgb_credit-output.beancount \
  ./example/cgb_credit/example-cgb_credit-records.csv
```

## 配置文件

```yaml
title: cgb_credit
defaultCurrency: CNY
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CGB:CreditCard
cgb_credit:
  rules:
    - type: 还款
      targetAccount: Equity:Transfers
    - item: 示例商户A
      targetAccount: Expenses:Food
    - item: 示例商户B
      targetAccount: Expenses:Shopping:Online
```

## 解析说明

- 正数金额按信用卡消费或分期记为支出。
- 负数金额按退款或还款记为收入，用于减少信用卡负债。
- `交易货币` 与 `入账货币` 不一致时，会在元数据中保留原交易金额。

## 示例文件

- [交易明细 CSV 示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cgb_credit/example-cgb_credit-records.csv)
- [转换后 Beancount 示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cgb_credit/example-cgb_credit-output.beancount)
- [配置文件示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cgb_credit/config.yaml)
