---
title: 兴业银行借记卡 (CIB Debit)
---

# 兴业银行借记卡 (CIB Debit) Provider

CIB Debit Provider 支持将兴业银行借记卡的 Excel 交易明细转换为 Beancount/Ledger 格式。

## 支持的文件格式

- XLSX 格式
- 支持同一次转换传入多个币种/子账户文件

## 使用方法

### 基本命令

```bash
double-entry-generator translate -p cib_debit -t beancount \
  --config example/cib_debit/config.yaml \
  example/cib_debit/example-cib_debit-cny-records.xlsx \
  example/cib_debit/example-cib_debit-hkd-records.xlsx \
  example/cib_debit/example-cib_debit-usd-records.xlsx
```

Ledger 输出：

```bash
double-entry-generator translate -p cib_debit -t ledger \
  --config example/cib_debit/config.yaml \
  example/cib_debit/example-cib_debit-cny-records.xlsx \
  example/cib_debit/example-cib_debit-hkd-records.xlsx \
  example/cib_debit/example-cib_debit-usd-records.xlsx
```

### 配置文件

```yaml
defaultMinusAccount: Income:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:CIB
defaultCurrency: CNY
title: 兴业银行借记卡示例

cib_debit:
  rules:
    - txType: 存款利息
      targetAccount: Income:Interest
    - txType: 转账转出
      targetAccount: Equity:Transfers
    - txType: 跨行消费,快捷支付
      targetAccount: Expenses:Daily
    - txType: 购汇
      targetAccount: Equity:Transfers
```

## 配置说明

### 全局账户

- `defaultCashAccount`: 兴业银行借记卡现金账户。
- `defaultMinusAccount`: 收入类交易未匹配规则时使用的对方账户。
- `defaultPlusAccount`: 支出类交易未匹配规则时使用的对方账户。
- `defaultCurrency`: 默认币种，通常为 `CNY`。

### 规则字段

CIB Debit 规则支持以下匹配字段：

- `peer`: 对方户名
- `peerBank`: 对方银行
- `peerAccount`: 对方账号
- `item`: 摘要/用途组合后的描述
- `type`: 收支方向
- `txType`: 交易摘要，例如 `存款利息`、`转账转出`、`购汇`
- `minPrice` / `maxPrice`: 金额区间
- `sep`: 多值分隔符，默认为 `,`
- `fullMatch`: 是否使用完整匹配
- `ignore`: 匹配后忽略该交易
- `tag`: 输出标签
- `methodAccount`: 银行卡账户
- `targetAccount`: 对方账户

## 多币种和购汇

同一次命令传入多个 CIB 子账户文件时，Provider 会按交易时间和金额自动合并可配对的购汇记录，输出为 `CurrencyExchange` 交易。例如 CNY 转 USD 会生成一条带 `@@` 总价的 Beancount/Ledger 分录，而不是两个互不关联的普通流水。

## 示例文件

- [配置示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/config.yaml)
- [CNY 输入示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-cny-records.xlsx)
- [HKD 输入示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-hkd-records.xlsx)
- [USD 输入示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-usd-records.xlsx)
- [Beancount 输出示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-output.beancount)
- [Ledger 输出示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/cib_debit/example-cib_debit-output.ledger)
