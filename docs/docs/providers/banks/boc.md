# 中国银行

`boc` provider 支持中国银行借记卡和信用卡账单。借记卡账单通常来自“中国银行交易流水明细单”，信用卡账单通常来自“中国银行信用卡账单”。

## 配置示例

```yaml
defaultMinusAccount: Assets:Bank:CN:BOC:Savings
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 中国银行
boc:
  rules:
    - method: "1528"
      methodAccount: Assets:Bank:CN:BOC:Savings
    - item: 结息,利息
      targetAccount: Income:Bank:CN:Interest
    - peer: 支付宝,微信,京东
      sep: ","
      ignore: true
```

## 规则字段

- `peer`：交易对手。借记卡流水会把对方户名和对方卡号合并用于匹配，因此可以用对方名称或账号片段匹配。
- `item`：交易名称或商品/摘要。
- `method`：本方卡号后四位。
- `type`：交易方向，例如收入或支出。
- `time` / `timestamp_range`：按交易时间匹配。
- `minPrice` / `maxPrice`：按金额区间匹配。

## 转换命令

```bash
double-entry-generator translate \
  --config ./config.yaml \
  --provider boc \
  --target beancount \
  --output output.beancount \
  statement.csv
```
