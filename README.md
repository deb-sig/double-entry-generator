# double-entry-generator

根据账单生成复式记账语言的代码。目前账单支持：

- 支付宝
- 微信
- 火币-币币交易

目前记账语言支持：

- BeanCount

架构支持扩展，如需支持新的账单（如银行账单等），可添加 [provider](pkg/provider)。如需支持新的记账语言，可添加 [compiler](pkg/compiler)。

```
┌───────────┐  ┌──────────┐  ┌────┐  ┌──────────┐  ┌──────────┐
│ translate │->│ provider │->│ IR │->│ compiler │->│ analyser │
└───────────┘  └──────────┘  └────┘  └──────────┘  └──────────┘
                  alipay               beancount      alipay
                  wechat                              wechat
                  huobi                               huobi
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

### Huobi Global (Crypto)

```bash
double-entry-generator translate --config ./example/huobi/config.yaml --provider huobi ./example/huobi/example-huobi-records.csv
```

## 账单下载与格式问题

### 支付宝

#### 下载方式

支付宝账单获取方式[见此](https://blog.triplez.cn/posts/bills-export-methods/#支付宝)。

`v0.2.0` 及以下版本请使用如下方式获取账单：

登录 PC 支付宝后，访问 [这里](https://consumeprod.alipay.com/record/standard.htm)，选择时间区间，下拉到页面底端，点击下载查询结果。

注意：请下载查询结果，而非[收支明细](https://cshall.alipay.com/lab/help_detail.htm?help_id=212688)。

#### 格式示例

[example-alipay-records.csv](./example/alipay/example-alipay-records.csv)

> 此示例为从「支付宝」APP 获取的账单格式。

### 微信

#### 下载方式

参考[百度经验](https://jingyan.baidu.com/article/1974b28941f977f4b0f7747b.html)。

#### 格式示例

[example-wechat-records.csv](./example/wechat/example-wechat-records.csv)

### Huobi Global (Crypto)

目前该项目只保证币币交易订单的转换，暂未测试合约、杠杆等交易订单。

> PR welcome :)

#### 下载方式

登录[火币 Global 网站](https://www.huobi.com/)，进入[币币订单的成交明细](https://www.huobi.com/zh-cn/transac/?tab=2&type=0)页面，选择合适的时间区间后，点击成交明细右上角的导出按钮即可。

#### 格式示例

[exmaple-huobi-records.csv](./example/huobi/example-huobi-records.csv)

转换后的结果示例：[exmaple-huobi-output.bean](./example/huobi/example-huobi-output.bean).

## 配置

### 支付宝
```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 测试
alipay:
  rules:
    - peer: 肯德基|麦当劳
      sep: '|'
      targetAccount: Expenses:Food
    - peer: 滴露
      targetAccount: Expenses:Groceries
    - peer: 苏宁
      targetAccount: Expenses:Electronics
    - item: 相互宝
      targetAccount: Expenses:Insurance

    - method: 余额 # 余额/余额宝
      methodAccount: Assets:Alipay
    - method: 招商银行(9876)
      methodAccount: Assets:Bank:CN:CMB-9876:Savings

    # 交易类型为其他
    - type: 其他
      item: 收益发放
      methodAccount: Income:Alipay:YuEBao:PnL
      targetAccount: Assets:Alipay
    - type: 其他
      peer: 蚂蚁财富
      item: 买入
      targetAccount: Assets:Alipay:Invest
      methodAccount: Assets:Alipay
    - type: 其他
      peer: 蚂蚁财富
      item: 卖出至余额宝
      targetAccount: Assets:Alipay
      methodAccount: Assets:Alipay:Invest
      pnlAccount: Income:Alipay:Invest:PnL
    - type: 其他
      item: 余额宝-单次转入
      targetAccount: Assets:Alipay
      methodAccount: Assets:Alipay
```

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

`alipay` is the provider-specific configuration. Alipay provider has rules matching mechanism.

`alipay` 是蚂蚁账单相关的配置。它提供基于规则的匹配。可以指定 type（收/支）、 method（支付方式）、peer（交易对方）和 item（商品名称）的包含匹配。匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` 覆盖默认定义。

支付宝提供了“交易方式”字段来标识资金出入账户。这样就可以直接通过“交易方式”，并辅以“收/支”字段确认该账户为增加账户还是减少账户。而复式记账法每笔交易至少需要两个账户，另一个账户则可通过“交易对方”（peer）、“商品”（item）、“收/支”（type）以及“交易方式”（method）的多种包含匹配得出。匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `methodAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `methodAccount` 的增减账户关系如下表：

|收/支|methodAccount|targetAccount|
|----|----|----|
|收入|plusAccount|minusAccount|
|支出|minusAccount|plusAccount|
|其他|minusAccount|plusAccount|

> 当交易类型为「其他」时，需要自行手动定义借贷账户。此时本软件会认为 `methodAccount` 是贷账户，`targetAccount` 是借账户。

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

微信账单提供了“交易方式”字段来标识资金出入账户。这样就可以直接通过“交易方式”，并辅以“收/支”字段确认该账户为增加账户还是减少账户。而复式记账法每笔交易至少需要两个账户，另一个账户则可通过“交易对方”（peer）、“商品”（item）、“收/支”（type）以及“交易方式”（method）的多种包含匹配得出。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `methodAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `methodAccount` 的增减账户关系如下表：

|收/支|methodAccount|targetAccount|
|----|----|----|
|收入|plusAccount|minusAccount|
|支出|minusAccount|plusAccount|

### Huobi Global (Crypto)

```yaml
defaultCashAccount: Assets:Huobi:Cash
defaultPositionAccount: Assets:Huobi:Positions
defaultCommissionAccount: Expenses:Huobi:Commission
defaultPnlAccount: Income:Huobi:PnL
defaultCurrency: USDT
title: 测试
huobi:
  rules:
    - item: BTC/USDT,BTC1S/USDT  # multiple keywords with separator
      type: 币币交易
      txType: 买入
      sep: ','  # define separator as a comma
      cashAccount: Assets:Rule1:Cash
      positionAccount: Assets:Rule1:Positions
      CommissionAccount: Expenses:Rule1:Commission
      pnlAccount: Income:Rule1:PnL
```

`defaultCashAccount`, `defaultPositionAccount`, `defaultCommissionAccount`, `defaultPnlAccount` 和 `defaultCurrency` 是全局的必填默认值。

`huobi` is the provider-specific configuration. Huobi provider has rules matching mechanism.

`huobi` 是火币相关的配置。它提供基于规则的匹配。可以指定：
- item（交易对）的包含匹配。
- type（交易类型）的包含匹配。
- txType（交易方向）的包含匹配。

在单条规则中可以使用分隔符（sep）填写多个关键字，在同一对象中，每个关键字之间是或的关系。

匹配成功则使用规则中定义的 `cashAccount`, `positionAccount`, `commissionAccount` 和 `pnlAccount` 覆盖默认定义。

其中：
- `defaultCashAccount` 是默认资本账户，一般用于存储 USDT。
- `defaultPositionAccount` 是默认持仓账户。
- `defaultCommissionAccount` 是默认手续费账户。
- `defaultPnlAccount` 是默认损益账户。
- `defaultCurrency` 是默认货币。

## Special Thanks

- [dilfish/atb](https://github.com/dilfish/atb) convert alipay bill to beancount version
