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

### 支付宝
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

### 微信

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: Test WeChat bills config
wechat:
  rules:
    - type: 收入 # 微信红包
      method: /
      item: /
      targetAccount: Income:Wechat:RedPacket
    - type: / # 转入零钱通
      peer: /
      item: /
      targetAccount: Assets:Digital:WeChat:Cash

    - peer: 滴滴
      targetAccount: Expenses:Transport:Taxi
    - peer: 公交
      targetAccount: Expenses:Transport:Bus
    - peer: 地铁
      targetAccount: Expenses:Transport:Metro
    - peer: 中国铁路
      targetAccount: Expenses:Transport:Train
    
    - method: / # 一般为收入，存入零钱
      methodAccount: Assets:Digital:Wechat:Cash
    - method: 零钱 # 零钱/零钱通
      methodAccount: Assets:Digital:Wechat:Cash
    - method: 工商银行
      methodAccount: Assets:Bank:CN:ICBC:Savings
```

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

`wechat` is the provider-specific configuration. WeChat provider has rules matching mechanism.

微信账单与支付宝不同的是，它提供了“交易方式”字段来标识资金出入账户。这样就可以直接通过“交易方式”，并辅以“收/支”字段确认该账户为增加账户还是减少账户。而复式记账法每笔交易至少需要两个账户，另一个账户则可通过“交易对方”（peer）、“商品”（item）、“收/支”（type）以及“交易方式”（method）的多种包含匹配得出。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `methodAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `methodAccount` 的增减账户关系如下表：

|收/支|methodAccount|targetAccount|
|----|----|----|
|收入|plusAccount|minusAccount|
|支出|minusAccount|plusAccount|

## Special Thanks

- [dilfish/atb](https://github.com/dilfish/atb) convert alipay bill to beancount version
