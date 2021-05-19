# double-entry-generator

根据账单生成复式记账语言的代码。目前账单支持：

- 支付宝
- 微信

目前记账语言支持：

- BeanCount

架构支持扩展，如需支持新的账单（如银行账单等），可添加 [provider](pkg/provider)。如需支持新的记账语言，可添加 [compiler](pkg/compiler)。

```
┌───────────┐  ┌──────────┐  ┌────┐  ┌──────────┐  ┌──────────┐
│ translate │─▶│ provider │─▶│ IR │─▶│ compiler │─▶│ analyser │
└───────────┘  └──────────┘  └────┘  └──────────┘  └──────────┘
```

## 安装

```bash
go get github.com/gaocegege/double-entry-generator
```

## 使用

请见[使用文档](doc/double-entry-generator_translate.md)

## 示例

### 支付宝

```bash
double-entry-generator translate --config ./example/alipay/config.yaml ./example/alipay/example-alipay-records.csv
```

其中 `--config` 是配置文件，默认情况下，使用支付宝作为提供方，也可手动指定 `--provider`。具体参考[使用文档](doc/double-entry-generator_translate.md)。默认生成的文件是 `default_output.beancount`:

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

### 微信

```bash
double-entry-generator translate --config ./example/wechat/config.yaml --provider wechat ./example/wechat/example-wechat-records.csv
```

## 账单下载与格式问题

### 支付宝

#### 下载方式

登录 PC 支付宝后，访问 [这里](https://consumeprod.alipay.com/record/standard.htm)，选择时间区间，下拉到页面底端，点击下载查询结果。

注意：请下载查询结果，而非[收支明细](https://cshall.alipay.com/lab/help_detail.htm?help_id=212688)。

#### 格式示例

[example-alipay-records.csv](./example/alipay/example-alipay-records.csv)

### 微信

#### 下载方式

参考[百度经验](https://jingyan.baidu.com/article/1974b28941f977f4b0f7747b.html)。

#### 格式示例

[example-wechat-records.csv](./example/wechat/example-wechat-records.csv)

## 配置

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

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

`alipay` is the provider-specific configuration. Alipay provider has rules matching mechanism.

`alipay` 是蚂蚁账单相关的配置。它提供基于规则的匹配。可以指定 peer（交易对方）和 item（商品名称）的包含匹配。匹配成功则使用规则中定义的 `plusAccount` 和 `minusAccount` 覆盖默认定义。

## Special Thanks

- [dilfish/atb](https://github.com/dilfish/atb) convert alipay bill to beancount version
