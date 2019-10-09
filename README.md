# double-entry-generator

Generate Double-Entry Bookkeeping Code From Bills

根据账单生成复式记账语言的代码。目前账单支持：

- 支付宝

目前记账语言支持：

- BeanCount

架构支持扩展，如需支持新的账单（如银行账单等），可添加 [provider](pkg/provider)。如需支持新的记账语言，可添加 [compiler](pkg/compiler)。

## Installation

```bash
go get github.com/gaocegege/double-entry-generator
```

## Usage

Please see the [doc](doc/double-entry-generator_translate.md)

## Example

```bash
double-entry-generator translate --config ./config.yaml ./example-alipay-records.csv
```

The result will be generated in `default_output.beancount`:

```
option "title" "测试"
option "operating_currency" "CNY"

1970-01-01 open Expenses:Food
1970-01-01 open Income:Earnings
1970-01-01 open Assets:Alipay
1970-01-01 open Expenses:Test
1970-01-01 open Liabilities:CreditCard:Test

2019-09-30 * "肯德基(张江高科餐厅)" "张江高科餐厅"
	Expenses:Food 27.00 CNY
	Liabilities:CreditCard:Test -27.00 CNY

2019-09-30 * "中欧基金管理有限公司" "余额宝-2019.09.29-收益发放"
	Assets:Alipay 0.01 CNY
	Income:Earnings -0.01 CNY
```

## Configuration

```
defaultMinusAccount: Liabilities:CreditCard:Test
defaultPlusAccount: Expenses:Test
defaultCurrency: CNY
title: 测试
alipay:
  rules:
    - peer: 餐厅
      plusAccount: Expenses:Food
    - item: 收益
      plusAccount: Assets:Alipay
      minusAccount: Income:Earnings
```

`defaultMinusAccount`, `defaultPlusAccount` and `defaultCurrency` are global default options. `defaultMinusAccount` is the default account which amount is the minuend.

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

`alipay` is the provider-specific configuration. Alipay provider has rules matching mechanism.

`alipay` 是蚂蚁账单相关的配置。它提供基于规则的匹配。可以指定 peer（交易对方）和 item（商品名称）的包含匹配。匹配成功则使用规则中定义的 `plusAccount` 和 `minusAccount` 覆盖默认定义。

## Special Thanks

- [dilfish/atb](https://github.com/dilfish/atb) convert alipay bill to beancount version
