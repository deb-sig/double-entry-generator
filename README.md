# double-entry-generator

根据账单生成复式记账语言的代码。目前账单支持：

- 支付宝
- 微信
- 火币-币币交易
- 海通证券
- 中国工商银行
- Toronto-Dominion Bank
- Bank of Montreal

目前记账语言支持：

- BeanCount
- Ledger

架构支持扩展，如需支持新的账单（如银行账单等），可添加 [provider](pkg/provider)。如需支持新的记账语言，可添加 [compiler](pkg/compiler)。

```
┌───────────┐  ┌──────────┐  ┌────┐  ┌──────────┐  ┌──────────┐
│ translate │->│ provider │->│ IR │->│ compiler │->│ analyser │
└───────────┘  └──────────┘  └────┘  └──────────┘  └──────────┘
                  alipay               beancount      alipay
                  wechat               ledger         wechat
                  huobi                               huobi
                  htsec                               htsec
                  icbc                                icbc
                  td                                  td
                  bmo                                 bmo
```

## 安装

### Homebrew

使用 Homebrew 安装：

```bash
brew install deb-sig/tap/double-entry-generator
```

使用 Homebrew 更新软件：

```bash
brew upgrade deb-sig/tap/double-entry-generator
```

### 二进制安装

在 [GitHub Release](https://github.com/deb-sig/double-entry-generator/releases) 页面中下载相应架构的二进制文件到本地即可。

### 源码安装

```bash
go get -u github.com/deb-sig/double-entry-generator
```

## 使用

请见[使用文档](doc/double-entry-generator_translate.md)

## 示例

### Beancount

#### 支付宝

```bash
double-entry-generator translate \
  --config ./example/alipay/config.yaml \
  --output ./example/alipay/example-alipay-output.beancount \
  ./example/alipay/example-alipay-records.csv
```

其中 `--config` 是配置文件，默认情况下，使用支付宝作为提供方，也可手动指定 `--provider`。具体参考[使用文档](doc/double-entry-generator_translate.md)。默认生成的文件是 `default_output.beancount`，若有 `--output` 或 `-o` 指定输出文件，则会输出到指定的文件中。如上述例子会将转换结果输出至 `./example/alipay/example-alipay-output.beancount` 文件中。

#### 微信

```bash
double-entry-generator translate \
  --config ./example/wechat/config.yaml \
  --provider wechat \
  --output ./example/wechat/example-wechat-output.beancount \
  ./example/wechat/example-wechat-records.csv
```

#### Huobi Global (Crypto)

```bash
double-entry-generator translate \
  --config ./example/huobi/config.yaml \
  --provider huobi \
  --output ./example/huobi/example-huobi-output.beancount \
  ./example/huobi/example-huobi-records.csv
```

#### 海通证券

```bash
double-entry-generator translate \
  --config ./example/htsec/config.yaml \
  --provider htsec \
  --output ./example/htsec/example-htsec-output.beancount \
  ./example/htsec/example-htsec-records.xlsx
```

#### 中国工商银行

```bash
double-entry-generator translate \
  --config ./example/icbc/credit/config.yaml \
  --provider icbc \
  --output ./example/icbc/credit/example-icbc-credit-output.beancount \
  ./example/icbc/credit/example-icbc-credit-records.csv
```

#### Toronto-Dominion Bank

```bash
double-entry-generator translate \
  --config ./example/td/config.yaml \
  --provider td \
  --output ./example/td/example-td-output.beancount \
  ./example/td/example-td-records.csv
```

#### Bank of Montreal

```bash
double-entry-generator translate \
  --config ./example/bmo/credit/config.yaml \
  --provider bmo \
  --output ./example/bmo/credit/example-bmo-output.beancount \
  ./example/bmo/credit/example-bmo-records.csv
```

### Ledger

#### 支付宝

```bash
double-entry-generator translate \
  --config ./example/alipay/config.yaml \
  --target ledger \
  --output ./example/alipay/example-alipay-output.ledger \
  ./example/alipay/example-alipay-records.csv
```

#### 微信

```bash
double-entry-generator translate \
  --config ./example/wechat/config.yaml \
  --provider wechat \
  --target ledger \
  --output ./example/wechat/example-wechat-output.ledger \
  ./example/wechat/example-wechat-records.csv
```

#### Huobi Global (Crypto)

```bash
double-entry-generator translate \
  --config ./example/huobi/config.yaml \
  --provider huobi \
  --target ledger \
  --output ./example/huobi/example-huobi-output.ledger \
  ./example/huobi/example-huobi-records.csv
```

#### 海通证券

```bash
double-entry-generator translate \
  --config ./example/htsec/config.yaml \
  --provider htsec \
  --target ledger \
  --output ./example/htsec/example-htsec-output.ledger \
  ./example/htsec/example-htsec-records.xlsx
```

#### 中国工商银行

```bash
double-entry-generator translate \
  --config ./example/icbc/credit/config.yaml \
  --provider icbc \
  --target ledger \
  --output ./example/icbc/credit/example-icbc-credit-output.ledger \
  ./example/icbc/credit/example-icbc-credit-records.csv
```

#### Toronto-Dominion Bank

```bash
double-entry-generator translate \
  --config ./example/td/config.yaml \
  --provider td \
  --target ledger \
  --output ./example/td/example-td-output.ledger \
  ./example/td/example-td-records.csv
```

#### Bank of Montreal

```bash
double-entry-generator translate \
  --config ./example/bmo/debit/config.yaml \
  --provider bmo \
  --target ledger \
  --output ./example/bmo/debit/example-bmo-output.ledger \
  ./example/bmo/debit/example-bmo-records.csv
```

## 账单下载与格式问题

### 支付宝

#### 下载方式

`v1.0.0` 及以上的版本请参考[此文章](https://blog.triplez.cn/posts/bills-export-methods/#%e6%94%af%e4%bb%98%e5%ae%9d)获取支付宝账单。

`v0.2.0` 及以下版本请使用此方式获取账单：登录 PC 支付宝后，访问 [这里](https://consumeprod.alipay.com/record/standard.htm)，选择时间区间，下拉到页面底端，点击下载查询结果。注意：请下载查询结果，而非[收支明细](https://cshall.alipay.com/lab/help_detail.htm?help_id=212688)。

#### 格式示例

[example-alipay-records.csv](./example/alipay/example-alipay-records.csv)

> 此示例为从「支付宝」APP 获取的账单格式。

转换后的结果示例：[exmaple-alipay-output.beancount](./example/alipay/example-alipay-output.beancount).

### 微信

#### 下载方式

微信支付的下载方式[见此](https://blog.triplez.cn/posts/bills-export-methods/#%e5%be%ae%e4%bf%a1%e6%94%af%e4%bb%98)。

#### 格式示例

[example-wechat-records.csv](./example/wechat/example-wechat-records.csv)

转换后的结果示例：[exmaple-wechat-output.beancount](./example/wechat/example-wechat-output.beancount).

### Huobi Global (Crypto)

目前该项目只保证币币交易订单的转换，暂未测试合约、杠杆等交易订单。

> PR welcome :)

#### 下载方式

登录[火币 Global 网站](https://www.huobi.com/)，进入[币币订单的成交明细](https://www.huobi.com/zh-cn/transac/?tab=2&type=0)页面，选择合适的时间区间后，点击成交明细右上角的导出按钮即可。

#### 格式示例

[exmaple-huobi-records.csv](./example/huobi/example-huobi-records.csv)

转换后的结果示例：[exmaple-huobi-output.beancount](./example/huobi/example-huobi-output.beancount).

### 海通证券

#### 下载方式

登录e海通财PC独立交易版PC客户端，左侧导航栏选择查询-交割单，右侧点击查询按钮导出交割单excel文件。

#### 格式示例

[example-htsec-records.csv](./example/htsec/example-htsec-records.xlsx)

转换后的结果示例：[exmaple-htsec-output.beancount](./example/htsec/example-htsec-output.beancount).

### 中国工商银行

#### 下载方式

中国工商银行账单的下载方式[见此](https://blog.triplez.cn/posts/bills-export-methods/#%e4%b8%ad%e5%9b%bd%e5%b7%a5%e5%95%86%e9%93%b6%e8%a1%8c)。

#### 格式示例

> `double-entry-generator` 能够自动识别出中国工商银行的账单类型（借记卡/信用卡）。

借记卡账单示例： [example-icbc-debit-records.csv](example/icbc/debit/example-icbc-debit-records.csv)

借记卡账单转换后的结果示例：[example-icbc-debit-output.beancount](example/icbc/debit/example-icbc-debit-output.beancount).

信用卡账单示例： [example-icbc-credit-records.csv](example/icbc/credit/example-icbc-credit-records.csv)

信用卡账单转换后的结果示例：[example-icbc-credit-output.beancount](example/icbc/credit/example-icbc-credit-output.beancount).

### Toronto-Dominion Bank

1. 登录TD 网页版本: https://easyweb.td.com/
2. 点击指定的账户
3. 选择账单范围 -> "Select Download Format" -> Spreadsheet(.csv) -> Download

#### 格式示例

[example-td-records.csv](./example/td/example-td-records.csv)

+ Beancount 转换的结果示例: [example-td-out.beancount](./example/td/example-td-output.beancount)
+ Ledger 转换的结果示例: [example-td-out.ledger](./example/td/example-td-output.ledger)

### Bank of Montreal

1. 登录 BMO 网页版本: https://www.bmo.com/en-ca/main/personal/
2. 选择指定账户
3. Transactions -> Download 选择时间范围

#### 格式示例

[example-bmo-record.csv](./example/bmo/debit/example-bmo-records.csv)

+ Beancount 转换的结果示例: [example-bmo-out.beancount](./example/bmo/debit/example-bmo-output.beancount)
+ Ledger 转换的结果示例: [example-bmo-out.ledger](./example/bmo/debit/example-bmo-output.ledger)

## 配置

### 支付宝

<details>
<summary>
  支付宝配置文件示例
</summary>

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 测试
alipay:
  rules:
    - category: 日用百货
      targetAccount: Expenses:Groceries
    - category: 餐饮美食
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: 16:00-22:00
      targetAccount: Expenses:Food:Dinner

    - peer: 滴露
      targetAccount: Expenses:Groceries
    - peer: 苏宁
      targetAccount: Expenses:Electronics
    - item: 相互宝
      targetAccount: Expenses:Insurance

    - method: 余额
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 余额宝
      fullMatch: true
      methodAccount: Assets:Alipay
    - method: 招商银行(9876)
      fullMatch: true
      methodAccount: Assets:Bank:CN:CMB-9876:Savings

    - type: 收入 # 其他转账收款
      targetAccount: Income:FIXME
      methodAccount: Assets:Alipay
    - type: 收入 # 收款码收款
      item: 商品
      targetAccount: Income:Alipay:ShouKuanMa
      methodAccount: Assets:Alipay

    # 交易类型为其他
    - type: 其他
      item: 收益发放
      methodAccount: Income:Alipay:YuEBao:PnL
      targetAccount: Assets:Alipay
    - type: 其他
      item: 余额宝-单次转入
      targetAccount: Assets:Alipay
      methodAccount: Assets:Alipay

    - peer: 基金
      type: 其他
      item: 黄金-买入
      methodAccount: Assets:Alipay
      targetAccount: Assets:Alipay:Invest:Gold
    - peer: 基金
      type: 其他
      item: 黄金-卖出
      methodAccount: Assets:Alipay:Invest:Gold
      targetAccount: Assets:Alipay
      pnlAccount: Income:Alipay:Invest:PnL
    - peer: 基金
      type: 其他
      item: 买入
      methodAccount: Assets:Alipay
      targetAccount: Assets:Alipay:Invest:Fund
    - peer: 基金
      type: 其他
      item: 卖出
      methodAccount: Assets:Alipay:Invest:Fund
      targetAccount: Assets:Alipay
      pnlAccount: Income:Alipay:Invest:PnL
```

</details></br>

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

`alipay` is the provider-specific configuration. Alipay provider has rules matching mechanism.

`alipay` 蚂蚁账单相关的配置。它提供基于规则的匹配。可以指定：
- `peer`（交易对方）的完全/包含匹配。
- `item`（商品说明）的完全/包含匹配。
- `type`（收/支）的完全/包含匹配。
- `method`（收/付款方式）的完全/包含匹配。
- `category`（交易分类）的完全/包含匹配。
- `time`（交易时间）的区间匹配。
  > 交易时间可写为以下两种形式：
  > - `11:00-13:00`
  > - `11:00:00-13:00:00`
  > 24 小时制，起始时间和终止之间之间使用 `-` 分隔。

在单条规则中可以使用分隔符（sep）填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `tag` 来设置流水的 [Tag](https://beancount.github.io/docs/beancount_language_syntax.html#tags)，使用 `sep` 作为分隔符。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `targetAccount` 、 `methodAccount` 等账户覆盖默认定义账户。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

支付宝提供了“交易方式”字段来标识资金出入账户。这样就可以直接通过“交易方式”，并辅以“收/支”字段确认该账户为增加账户还是减少账户。而复式记账法每笔交易至少需要两个账户，另一个账户则可通过“交易对方”（peer）、“商品”（item）、“收/支”（type）以及“交易方式”（method）的多种包含匹配得出。匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `methodAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `methodAccount` 的增减账户关系如下表：


| 收/支 | 交易分类 | minusAccount  | plusAccount   |
| ----- | -------- | ------------- | ------------- |
| 收入  | *        | targetAccount | methodAccount |
| 收入  | 退款     | targetAccount | methodAccount |
| 支出  | *        | methodAccount | targetAccount |
| 其他  | *        | methodAccount | targetAccount |
| 其他  | 退款     | targetAccount | methodAccount |

> 当交易类型为「其他」时，需要自行手动定义借贷账户。此时本软件会认为 `methodAccount` 是贷账户，`targetAccount` 是借账户。

### 微信

<details>
<summary>
  微信配置文件示例
</summary>

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCommissionAccount: Expenses:Commission:FIXME
defaultCurrency: CNY
title: 测试
wechat:
  rules:
    # type (additional condition)
    - type: 收入 # 微信红包
      method: /
      item: /
      targetAccount: Income:Wechat:RedPacket
    - type: / # 转入零钱通
      txType: 转入零钱
      peer: /
      item: /
      targetAccount: Assets:Digital:Wechat:Cash
    - type: / # 零钱提现
      txType: 零钱提现
      targetAccount: Assets:Digital:Wechat:Cash
      commissionAccount: Expenses:Wechat:Commission
    - type: / # 零钱充值
      txType: 零钱充值
      targetAccount: Assets:Digital:Wechat:Cash
    - type: / # 零钱通转出-到工商银行(9876)
      txType: 零钱通转出-到工商银行(9876)
      targetAccount: Assets:Bank:CN:ICBC:Savings

    - peer: 云膳过桥米线,餐厅
      sep: ','
      time: 11:00-15:00
      targetAccount: Expenses:Food:Meal:Lunch
    - peer: 云膳过桥米线,餐厅
      sep: ','
      time: 16:30-21:30
      targetAccount: Expenses:Food:Meal:Dinner
    - peer: 餐厅
      time: 23:55-00:10 # test T+1
      targetAccount: Expenses:Food:Meal:MidNight
    - peer: 餐厅
      time: 23:50-00:05 # test T-1
      targetAccount: Expenses:Food:Meal:MidNight

    - peer: 房东
      type: 支出
      targetAccount: Expenses:Housing:Rent

    - peer: 用户
      type: 收入
      targetAccount: Income:Service

    - peer: 理财通
      type: /
      targetAccount: Assets:Trade:Tencent:LiCaiTong

    - peer: 建设银行
      txType: 信用卡还款
      targetAccount: Liabilities:Bank:CN:CCB

    - method: / # 一般为收入，存入零钱
      methodAccount: Assets:Digital:Wechat:Cash
    - method: 零钱
      fullMatch: true
      methodAccount: Assets:Digital:Wechat:Cash
    - method: 零钱通
      fullMatch: true
      methodAccount: Assets:Digital:Wechat:Cash
    - method: 工商银行
      methodAccount: Assets:Bank:CN:ICBC:Savings
    - method: 中国银行
      methodAccount: Assets:Bank:CN:BOC:Savings

```

</details></br>

`defaultMinusAccount`, `defaultPlusAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户。 `defaultCurrency` 是默认货币。

> `defaultCommissionAccount` 是默认的服务费账户，若无服务费相关交易，则不需要填写。但笔者还是建议填写一个占位 FIXME 账户，否则遇到带服务费的交易，转换器会报错退出。

`wechat` is the provider-specific configuration. WeChat provider has rules matching mechanism.

`wechat` 是微信相关的配置。它提供基于规则的匹配。可以指定：
- `peer`（交易对方）的完全/包含匹配。
- `item`（商品名称）的完全/包含匹配。
- `type`（收/支）的完全/包含匹配。
- `txType`（交易类型）的完全/包含匹配。
- `method`（支付方式）的完全/包含匹配。
- `time`（交易时间）的区间匹配。
  > 交易时间可写为以下两种形式：
  > - `11:00-13:00`
  > - `11:00:00-13:00:00`
  > 24 小时制，起始时间和终止之间之间使用 `-` 分隔。

在单条规则中可以使用分隔符（sep）填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `tag` 来设置流水的 [Tag](https://beancount.github.io/docs/beancount_language_syntax.html#tags)，使用 `sep` 作为分隔符。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `targetAccount` 、 `methodAccount` 等账户覆盖默认定义账户。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

微信账单提供了“交易方式”字段来标识资金出入账户。这样就可以直接通过“交易方式”，并辅以“收/支”字段确认该账户为增加账户还是减少账户。而复式记账法每笔交易至少需要两个账户，另一个账户则可通过“交易对方”（peer）、“商品”（item）、“收/支”（type）以及“交易方式”（method）的多种包含匹配得出。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和 `methodAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `methodAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `methodAccount` 的增减账户关系如下表：

| 收/支 | minusAccount  | plusAccount   |
| ----- | ------------- | ------------- |
| 收入  | targetAccount | methodAccount |
| 支出  | methodAccount | targetAccount |

### Huobi Global (Crypto)

<details>
<summary>
  火币-币币交易配置文件示例
</summary>

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
      type: 买入
      txType: 币币交易
      fullMatch: true
      sep: ','  # define separator as a comma
      cashAccount: Assets:Rule1:Cash
      positionAccount: Assets:Rule1:Positions
      CommissionAccount: Expenses:Rule1:Commission
      pnlAccount: Income:Rule1:PnL
```

</details></br>

`defaultCashAccount`, `defaultPositionAccount`, `defaultCommissionAccount`, `defaultPnlAccount` 和 `defaultCurrency` 是全局的必填默认值。

`huobi` is the provider-specific configuration. Huobi provider has rules matching mechanism.

`huobi` 是火币相关的配置。它提供基于规则的匹配。可以指定：
- `item`（交易对）的完全/包含匹配。
- `type`（交易方向）的完全/包含匹配。
- `txType`（交易类型）的完全/包含匹配。
- `time`（交易时间）的区间匹配。
  > 交易时间可写为以下两种形式：
  > - `11:00-13:00`
  > - `11:00:00-13:00:00`
  > 24 小时制，起始时间和终止之间之间使用 `-` 分隔。

在单条规则中可以使用分隔符（sep）填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `cashAccount`, `positionAccount`, `commissionAccount` 和 `pnlAccount` 覆盖默认定义。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

其中：
- `defaultCashAccount` 是默认资本账户，一般用于存储 USDT。
- `defaultPositionAccount` 是默认持仓账户。
- `defaultCommissionAccount` 是默认手续费账户。
- `defaultPnlAccount` 是默认损益账户。
- `defaultCurrency` 是默认货币。

### 海通证券

<details>
<summary>
  海通证券交割单配置文件示例
</summary>

```yaml
defaultCashAccount: Assets:Htsec:Cash
defaultPositionAccount: Assets:Htsec:Positions
defaultCommissionAccount: Expenses:Htsec:Commission
defaultPnlAccount: Income:Htsec:PnL
defaultCurrency: CNY
title: 测试
htsec:
  rules:
    - item: 兴业转债
      type: 卖
      sep: ','
      cashAccount: Assets:Rule1:Cash
      positionAccount: Assets:Rule1:Positions
      CommissionAccount: Expenses:Rule1:Commission
      pnlAccount: Income:Rule1:PnL
```

</details></br>

`defaultCashAccount`, `defaultPositionAccount`, `defaultCommissionAccount`, `defaultPnlAccount` 和 `defaultCurrency` 是全局的必填默认值。

`htsec` is the provider-specific configuration. Htsec provider has rules matching mechanism.

`htsec` 是海通证券相关的配置。它提供基于规则的匹配。可以指定：
- `item`（交易方向-证券编码-证券市值）的完全/包含匹配。
- `type`（交易方向）的完全/包含匹配。
- `time`（交易时间）的区间匹配。
  > 交易时间可写为以下两种形式：
  > - `11:00-13:00`
  > - `11:00:00-13:00:00`
      > 24 小时制，起始时间和终止之间之间使用 `-` 分隔。

在单条规则中可以使用分隔符（sep）填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `cashAccount`, `positionAccount`, `commissionAccount` 和 `pnlAccount` 覆盖默认定义。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

其中：
- `defaultCashAccount` 是默认资本账户，一般用于存储证券账户可用资金。
- `defaultPositionAccount` 是默认持仓账户。
- `defaultCommissionAccount` 是默认手续费账户。
- `defaultPnlAccount` 是默认损益账户。
- `defaultCurrency` 是默认货币。

### 中国工商银行

<details>
<summary>
  中国工商银行配置文件示例
</summary>

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:Bank:CN:ICBC
defaultCurrency: CNY
title: 测试
icbc:
  rules:
    - peer: 财付通,支付宝
      ignore: true
    - peer: 广东联合电子收费股份
      targetAccount: Expenses:Transport:Highway
    - txType: 人民币自动转帐还款
      targetAccount: Assets:Bank:CN:ICBC:Savings
    - peer: XX旗舰店
      targetAccount: Expenses:Joy
```

</details></br>

`defaultMinusAccount`, `defaultPlusAccount`, `defaultCashAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户， `defaultCashAccount` 是该配置中默认使用的银行卡账户（等同于支付宝/微信中的 `methodAccount` ）。 `defaultCurrency` 是默认货币。

`icbc` 是中国工商银行相关的配置。它提供基于规则的匹配。可以指定：
- `peer`（交易对方）的完全/包含匹配。
- `type`（收/支）的完全/包含匹配。
- `txType`（交易类型）的完全/包含匹配。

在单条规则中可以使用分隔符 `sep` 填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `tag` 来设置流水的 [Tag](https://beancount.github.io/docs/beancount_language_syntax.html#tags)，使用 `sep` 作为分隔符。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `targetAccount` 账户覆盖默认定义账户。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

中国工商银行账单中的记账金额中存在收入/支出之分，通过这个机制就可以判断银行卡账户在交易中的正负关系。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和全局值 `defaultCashAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `defaultCashAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `defaultCashAccount` 的增减账户关系如下表：

| 收/支 | minusAccount       | plusAccount        |
| ----- | ------------------ | ------------------ |
| 收入  | targetAccount      | defaultCashAccount |
| 支出  | defaultCashAccount | targetAccount      |

### Toronto-Dominion Bank

<details>
<summary>
  TD银行配置文件示例
</summary>

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:TDChequing
defaultCurrency: CAD
title: 测试
td:
  rules:
    - item: "T T"
      targetAccount: Expenses:Grocery
      tag: tt_tag
    - item: "DOLLARAMA"
      targetAccount: Expenses:Grocery
      tag: grocery_tag1,cheap_tag2
    - item: "DEVELOPM MSP"
      targetAccount: Income:Salary
    - type: 收入
      item: "SEND E-TFR"
      targetAccount: Income:FIXME

```

</details></br>

`defaultMinusAccount`, `defaultPlusAccount`, `defaultCashAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户， `defaultCashAccount` 是该配置中默认使用的银行卡账户（等同于支付宝/微信中的 `methodAccount` ）。 `defaultCurrency` 是默认货币。

`td` 是 Toronto-Dominion Bank相关的配置。它提供基于规则的匹配。因为TD本身的账单较简单，所以可以指定的规则不多：
- `item`:（交易商品）的完全/包含匹配。
- `type`:（收/支）的完全/包含匹配。

在单条规则中可以使用分隔符 `sep` 填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `tag` 来设置流水的 [Beancount Tag](https://beancount.github.io/docs/beancount_language_syntax.html#tags)或[Ledger Meta Tag](https://ledger-cli.org/doc/ledger3.html#Metadata-tags)，使用 `sep` 作为分隔符。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `targetAccount` 账户覆盖默认定义账户。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

TD账单中的记账金额中存在收入/支出之分，通过这个机制就可以判断银行卡账户在交易中的正负关系。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和全局值 `defaultCashAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `defaultCashAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `defaultCashAccount` 的增减账户关系如下表：

| 收/支 | minusAccount       | plusAccount        |
| ----- | ------------------ | ------------------ |
| 收入  | targetAccount      | defaultCashAccount |
| 支出  | defaultCashAccount | targetAccount      |

#### Bank of Montreal

<details>
<summary>
  BMO银行配置文件示例
</summary>

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:BMOChequing
defaultCurrency: CAD
title: 测试
bmo:
  rules:
    - item: "T T"
      targetAccount: Expenses:Grocery
      tag: tt_tag
    - item: "DOLLARAMA"
      targetAccount: Expenses:Grocery
      tag: grocery_tag1,cheap_tag2
    - item: "DEVELOPM MSP"
      targetAccount: Income:Salary
    - type: 收入
      item: "SEND E-TFR"
      targetAccount: Income:FIXME

```

</details></br>

`defaultMinusAccount`, `defaultPlusAccount`, `defaultCashAccount` 和 `defaultCurrency` 是全局的必填默认值。其中 `defaultMinusAccount` 是默认金额减少的账户，`defaultPlusAccount` 是默认金额增加的账户， `defaultCashAccount` 是该配置中默认使用的银行卡账户（等同于支付宝/微信中的 `methodAccount` ）。 `defaultCurrency` 是默认货币。

`bmo` 是 Toronto-Dominion Bank相关的配置。它提供基于规则的匹配。因为BMO本身的账单较简单，所以可以指定的规则不多：
- `item`:（交易商品）的完全/包含匹配。
- `type`:（收/支）的完全/包含匹配。

在单条规则中可以使用分隔符 `sep` 填写多个关键字，在同一对象中，每个关键字之间是或的关系。

在单条规则中可以使用 `fullMatch` 来设置字符匹配规则，`true` 表示使用完全匹配(full match)，`false` 表示使用包含匹配(partial match)，不设置该项则默认使用包含匹配。

在单条规则中可以使用 `tag` 来设置流水的 [Beancount Tag](https://beancount.github.io/docs/beancount_language_syntax.html#tags)或[Ledger Meta Tag](https://ledger-cli.org/doc/ledger3.html#Metadata-tags)，使用 `sep` 作为分隔符。

在单条规则中可以使用 `ignore` 来设置是否忽略匹配上该规则的交易，`true` 表示忽略匹配上该规则的交易，`fasle` 则为不忽略，缺省为 `false` 。

匹配成功则使用规则中定义的 `targetAccount` 账户覆盖默认定义账户。

规则匹配的顺序是：从 `rules` 配置中的第一条开始匹配，如果匹配成功仍继续匹配。也就是后面的规则优先级要**高于**前面的规则。

BMO账单中的记账金额中存在收入/支出之分，通过这个机制就可以判断银行卡账户在交易中的正负关系。如支付宝配置类似，匹配成功则使用规则中定义的 `targetAccount` 和全局值 `defaultCashAccount` ，并通过确认该笔交易是收入还是支出，决定 `targetAccount` 和 `defaultCashAccount` 的正负关系，来覆盖默认定义的增减账户。

`targetAccount` 与 `defaultCashAccount` 的增减账户关系如下表：

| 收/支 | minusAccount       | plusAccount        |
| ----- | ------------------ | ------------------ |
| 收入  | targetAccount      | defaultCashAccount |
| 支出  | defaultCashAccount | targetAccount      |

## Special Thanks

- [dilfish/atb](https://github.com/dilfish/atb) convert alipay bill to beancount version
