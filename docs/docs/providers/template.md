---
title: 通用模板 Provider
description: 使用运行时模板和规则导入账单
---

# 通用模板 Provider

传统 provider 需要在 DEG 内部新增 Go 包、注册 provider、维护 analyser/config 逻辑。随着 provider 数量增加，这种方式会让贡献和维护成本持续升高：每个 PR 都可能带来一套新逻辑，维护者需要重新 review、测试和理解 provider 私有行为，开发者也需要先理解 DEG 内部结构才能贡献。

通用模板 Provider 的目标是把“账单格式”和“导入规则”从 DEG 本体中拆出来。DEG 本体只负责读取账单、解释模板、执行规则并输出 beancount/ledger；具体账单格式由模板仓库维护。这样新增和更新 provider 时，更多工作可以集中在模板和规则文件上，DEG 开发可以更专注于规则引擎、导入体验和输出能力本身。

## 基本流程

```text
账单文件
  -> 槽位和元数据映射
  -> 个人规则（账户、元数据增删改）
  -> 没写到的一侧补 FIXME
  -> beancount / ledger
```

模板只声明 Beancount 已有的槽位，以及这份账单自己的元数据键。账户名只写在个人规则里。下面的「槽位契约」是现在要写的格式。更早的 `templateRules` 仍然可以导入旧模板。

## 查看模板

查看模板仓库中可用的模板：

```bash
double-entry-generator template list
```

按关键字搜索模板：

```bash
double-entry-generator template search wechat
double-entry-generator template search 支付
```

如果搜索结果只有一个，命令会展开显示模板分类、标签和可 pin 的版本。

## 生成个人规则骨架

从模板仓库生成个人规则文件：

```bash
double-entry-generator config init wechat -o wechat-rules.yaml
```

也可以 pin 到指定模板版本：

```bash
double-entry-generator config init wechat@2026-04-28 -o wechat-rules.yaml
```

## 导入账单

只使用模板规则导入：

```bash
double-entry-generator import wechat bill.csv -o output.bean
```

使用个人规则导入：

```bash
double-entry-generator import wechat bill.csv --rules wechat-rules.yaml -o output.bean
```

使用指定版本的模板：

```bash
double-entry-generator import wechat@2026-04-28 bill.csv --rules wechat-rules.yaml -o output.bean
```

使用本地模板文件：

```bash
double-entry-generator import ./wechat.yaml bill.csv --rules wechat-rules.yaml -o output.bean
```

## 补全

生成 shell 补全脚本：

```bash
double-entry-generator completion zsh
double-entry-generator completion bash
double-entry-generator completion powershell
```

补全支持：

- `double-entry-generator import <TAB>`：查询线上模板。
- `double-entry-generator import wechat@<TAB>`：查询模板版本。
- `double-entry-generator import wechat <TAB>`：补全账单文件。
- `double-entry-generator import wechat --rules <TAB>`：补全个人规则 YAML。
- `double-entry-generator template <TAB>`：补全模板命令。

`--rules` 是选填项。只有在用户输入 `-` 或 `--` 并触发补全时，才会作为 flag 候选出现。

## 模板文件

模板文件描述账单如何被读取。常见字段如下：

```yaml
schema: https://deg.dev/template-profile/v2
id: htsec
name: htsec
template:
    fileFormat: xlsx          # csv / xlsx / xls
    encoding: gb18030         # 可选，CSV/XLS 文本编码
    delimiter: ','            # CSV 分隔符；tab 可写 "\t"
    skipLeadingRows: 1        # 可选，跳过说明行
    skipInvalidRows: true     # 可选，跳过无法解析的汇总/空行
    sourceHeaders:
        - 交易日期
        - 交易时间
        - 金额
        - 摘要
    defaultCurrency: CNY
```

导入时，`sourceHeaders` 会成为规则可引用的字段。字段引用写作 `<字段名>`，例如 `<金额>`、`<交易日期>`。

### reader 块

`template.fileFormat / encoding / delimiter` 只够描述表格文件。账单不是表格时（JSON、XML、PDF 转出的文本），用 `reader:` 块说明字节怎么变成表格。有 `reader:` 时它优先；没有时沿用 `template.*` 的旧字段，旧模板不用改。

Reader 只认结构，不认语义：无论什么格式，产出都是一张字符串表格，后面的规则一视同仁。

```yaml
reader:
  format: csv            # csv | xlsx | xls | json | xml | text
  encoding: gb18030      # utf-8（默认）/ gbk / gb18030 / utf-16le / utf-16be
  delimiter: ","
  stripTabs: true
  sheet: 交易明细          # xlsx / xls：工作表名或 0 起的序号，缺省第一张
```

**json / xml**：`records` 是一个 XPath，每个命中的节点是一行；`columns` 把列名映射到相对于该节点的 XPath。两种格式用同一种查询语法。

```yaml
reader:
  format: xml
  records: "//Ntry"                       # camt.053 的每条分录
  columns:
    date: BookgDt/Dt
    amount: Amt
    currency: Amt/@Ccy                    # 属性
    direction: CdtDbtInd
    narration: NtryDtls/TxDtls/RmtInf/Ustrd
```

```yaml
reader:
  format: json
  records: "//result/*"                   # 数组的每个元素
  columns:
    hash: hash
    time: timeStamp
    value: value
```

`columns` 里声明的名字就是 `sourceHeaders`，规则里用 `<date>`、`<hash>` 引用。

**text**：面向版面文本，主要是 PDF 对账单经 `pdftotext -layout` 转出的结果。`record` 是带命名分组的正则，命中的行成为一行，分组名就是列名；`continuation` 匹配续行（PDF 里摘要换行很常见），其分组内容追加到上一行同名列。两者都不命中的行（页眉、页脚、合计）直接忽略。

```yaml
reader:
  format: text
  convert: pdftotext-layout               # 输入是 pdf 时先转文本；需要安装 poppler-utils
  record: '^(?P<date>\d{4}-\d{2}-\d{2})\s+(?P<narration>.+?)\s+(?P<amount>-?[\d,]+\.\d{2})\s+(?P<balance>[\d,]+\.\d{2})$'
  continuation: '^\s{10,}(?P<narration>\S.*)$'
```

没有 `pdftotext` 时导入会报错并给出手工转换命令；也可以自己转好后导入 `.txt`。浏览器端由前端用 pdf.js 转文本后交给同一个模板。

**html**：银行邮件账单、导出的网页对账单。和 json/xml 一样用 `records` + `columns`，XPath 作用在 DOM 上：

```yaml
reader:
  format: html
  records: "//table[@id='txns']//tr[position()>1]"
  columns: { date: "td[1]", narration: "td[2]", amount: "td[3]" }
```

**eml**：保存下来的邮件是容器，不是账单格式。`part` 选一个 MIME 部分（`text/html`、`text/plain` 或 `attachment:*.pdf` 这样的附件名通配），`inner` 是读那一部分用的 Reader，可以嵌套：

```yaml
reader:
  format: eml
  part: text/html                         # 交行：邮件正文是 HTML 表格
  inner:
    format: html
    records: "//table//tr[td]"
    columns: { date: "td[1]", narration: "td[3]", amount: "td[5]" }
```

```yaml
reader:
  format: eml
  part: "attachment:*.pdf"                # 招行：邮件附件是 PDF
  inner: { format: text, convert: pdftotext-layout, record: '...' }
```

**api**：账单不是文件而是接口（链上地址、交易所 API）。`import` 的第二个参数是来源（地址、账号），填进 `{source}`；`{env.NAME}` 读环境变量，密钥不进模板。响应按 `response`（默认 json，可选 xml）用 `records`/`columns` 读：

```yaml
reader:
  format: api
  url: "https://api.etherscan.io/v2/api?chainid=1&module=account&action=txlist&address={source}&apikey={env.ETHERSCAN_KEY}"
  records: "//result/*"
  columns: { hash: hash, time: timeStamp, from: from, to: to, value: value, gasUsed: gasUsed, gasPrice: gasPrice }
```

```bash
ETHERSCAN_KEY=... double-entry-generator import etherscan-eth 0x1234... --rules eth-rules.yaml
```

环境变量缺失时直接报错，不会带着空密钥去请求。

### shape 块

Reader 产出表格之后、规则运行之前，`shape:` 回答一个问题：哪些行是一笔交易。表头在第几行、说明行和合计行、失败的订单、一笔拆成两行、一行带两笔，都在这里处理完，规则层只会看到"一条记录就是一笔交易"。

`shape:` 是有序步骤列表，按写的顺序执行：

```yaml
shape:
  - locateHeader: { anchor: [交易时间, 金额(元)], scanRows: 50 }   # 用锚点列名找表头，不数行数
  - dropMatching: '^(共计|合计|导出说明)'                           # 整行文本匹配正则就丢，表头前后都可
  - dropIf: '<当前状态> ~ "失败|已关闭" || <金额(元)>.number == 0'  # 条件语法同规则的 when
  - merge:                                                        # 多行一笔：按 key 合并
      key: [<合同号>, <成交号>]
      take: { 成交金额: first-nonzero, 成交数量: first-nonzero, 手续费: sum, 备注: join }
  - split:                                                        # 一行多笔：拆成多条记录
      when: '<交易类型> == "零钱提现" && <备注> ~ "服务费"'
      into:
        - {}                                                      # 原行
        - { 交易类型: 手续费, 金额(元): '<备注>.extract("服务费.?([.0-9]+)")' }
```

- `locateHeader`：第一个包含全部 `anchor` 列名的行就是表头，它之前的行全部丢掉。微信、支付宝、建行历次改版多数只是说明行多了少了一行，用锚点就不受影响。
- `dropMatching`：正则对整行（单元格以空格连接）匹配，适合页眉页脚、合计行。
- `dropIf`：对列求条件，适合按状态、金额过滤。
- `merge`：`key` 相同的记录合成一条，`take` 决定每列怎么合：`first`（默认）、`last`、`first-nonzero`、`sum`、`join`。`key` 为空的记录不参与合并。
- `split`：命中 `when` 的记录替换成 `into` 里的若干条，每条是原记录加上覆盖的列；`{}` 表示原样复制。

没写 `locateHeader` 时，表头仍按 `template.skipLeadingRows / sourceHeaders / headerLocate` 的旧规则确定，所以可以只加一个 `dropIf` 而不改别的。

还有两个步骤：

```yaml
shape:
  - capture: { pattern: '账[\s\p{Zs}]*号[：:][\s\p{Zs}]*(?P<account>\S+)', scanRows: 6 }
  - capture: { pattern: '(?P<year>\d{4})年(?P<month>\d{2})月', scanRows: 3 }
  - locateHeader: { anchor: [交易日期, 金额] }
  - failIf: '<交易类型> != "消费" && <交易类型> != "还款"'
    message: 出现了模板不认识的交易类型，账单格式可能变了
```

- `capture`：从表格以外的行（通常是表头上方的说明区）取账单级的值，比如账号、卡别名、出账年月。命名分组就是名字，之后在任何地方用 `<file.account>` 引用。要放在 `locateHeader` 之前，否则说明行已经被丢掉了。中文账单常用全角空格，`\s` 不匹配全角空格，要写 `[\s\p{Zs}]`。找不到某个分组时导入直接报错。
- `failIf`：任意一条记录命中就停止导入，报出 `message` 和那一行的内容。用来在账单格式变化时尽早失败，而不是悄悄写出错账。

没有年份的日期（如信用卡账单的 `12/30`）不需要专门的语法，用 `capture` 加 `vars` 组合：

```yaml
template:
  vars:
    - vars: { prev: '<file.year> - 1' }
    - vars: { year: <file.year> }
    - when: '<交易日>.extract("^(\d+)") > <file.month>'   # 账单月之后的月份属于上一年
      vars: { year: '<var.prev>.format("%.0f")' }
  slots:
    date: <var.year>/<交易日>
```

## 规则文件

规则文件通常包含三块：

```yaml
options:
    title: 我的账本
    operatingCurrency: CNY

templateRules:
    - id: 模板基础交易
      actions:
          date: <交易日期> <交易时间>
          payee: <交易对方>
          narration: <摘要>

personalRules:
    - id: 餐饮支出
      when: <交易对方> ~ "美团"
      actions:
          to:
              account: Expenses:Food
```

- `options` 控制输出账本标题和本位币。
- `templateRules` 描述账单格式自身，通常由模板维护者提供。
- `personalRules` 描述个人账户分类，通常由用户维护。
- 规则按顺序执行；后命中的规则可以覆盖前面设置的字段或变量。

## 条件语法

`when` 用来判断规则是否命中：

```yaml
when: <收/支> == "支出"
when: <金额>.number >= 10
when: <交易对方> ~ "美团"
when: <交易对方> !~ "微信"
when: (<方向> == "买入" || <方向> == "卖出") && <交易类型> == "币币交易"
```

支持的比较符：

- `==`、`!=`
- `>`、`>=`、`<`、`<=`（两边都是数字时按数值比较）
- `~` 包含
- `!~` 不包含
- `^=` 以…开头，`$=` 以…结尾
- `=~` 整串匹配正则，例如 `<地址> =~ "(?i)0x1429.*"`（`(?i)` 忽略大小写）
- `&&` 与
- `||` 或

`<file.x>` 引用 `shape.capture` 从说明行取到的值，`<var.x>` 引用模板变量，`metadata.x` 引用映射后的元数据。

## 字段方法

字段可以串联方法：

```yaml
<金额>.number
<金额>.+
<金额>.-
<金额>.!
<证券代码>.format("%06.0f")
<手续费>.extract("^([.0-9]+)")
raw[交易创建时间].time
```

常用方法：

- `.number`：清理金额字符串，例如货币符号、千分位。
- `.+`：强制为正数。
- `.-`：强制为负数。
- `.!`：反转正负。
- `.extract("regex")`：用正则提取文本。
- `.format("...")`：使用格式模板输出，例如 `%.2f`、`%06.0f`。
- `.date`、`.time`、`.timestamp`：从时间文本中提取日期、时间或 Unix 时间戳。

金额表达式支持简单算术：

```yaml
<数量>.number * <价格>.number
<手续费>.number + <印花税>.number + <过户费>.number
```

## Actions

当前通用 runtime 的核心 actions 是：

```yaml
actions:
    date: <交易时间>
    payee: <交易对象>
    narration: <商品/说明>
    note: <备注>
    amount: <金额>.number
    currency: CNY
    from:
        account: Assets:Bank
    to:
        account: Expenses:Food
    metadata:
        orderId: <订单号>
    tags:
        - Food
    vars:
        cash: Assets:Broker:Cash
    postings:
        - <var.cash> -<金额>.format("%.2f") CNY
    ignore: true
```

普通流水可以用 `from` / `to` 快捷生成双分录：

```yaml
- id: 默认支出
  when: <收/支> == "支出"
  actions:
      from:
          account: Assets:FIXME
      to:
          account: Expenses:FIXME
      amount: <金额>.number
      currency: CNY
```

复杂交易建议直接写 `postings`，这样规则表达的是最终账本分录，而不是 provider 或 IR 私有字段：

```yaml
- id: 证券买入
  when: <业务类型> == "证券买入"
  actions:
      narration: 证券买入-<证券代码>.format("%06.0f")-<证券名称>
      postings:
          - <var.cash> -<成交金额>.format("%.2f") CNY
          - <var.position> <成交数量>.format("%.2f") <var.security> {<成交价格>.format("%.3f") CNY} @@ <成交金额>.format("%.2f") CNY
          - <var.cash> -<手续费>.format("%.2f") CNY
          - <var.feeExpense> <手续费>.format("%.2f") CNY
```

## Vars

`vars` 是规则变量，用来复用或覆盖账户、币种和中间值。它不是 IR 字段，只在规则渲染期间存在。

模板规则可以提供默认变量：

```yaml
templateRules:
    - id: 证券默认变量
      actions:
          vars:
              cash: Assets:Broker:Cash
              position: Assets:Broker:Positions
              feeExpense: Expenses:Broker:Commission
              security: SH<证券代码>.format("%06.0f")
```

个人规则可以覆盖同名变量：

```yaml
personalRules:
    - id: HS300ETF账户覆盖
      when: <证券名称> ~ "HS300ETF"
      actions:
          vars:
              cash: Assets:Rule1:Cash
              position: Assets:Broker:Positions:沪深300
              feeExpense: Expenses:Rule1:Commission
```

后续 postings 中的 `<var.cash>`、`<var.position>` 会使用覆盖后的值。

## 忽略和辅助行

辅助行可以用 `ignore: true` 跳过。例如某些账单把一笔业务拆成“金额行”和“数量价格行”，可以把金额-only 行忽略，在数量价格行上用规则计算金额：

```yaml
- id: 拆分成交目标
  when: <成交数量> != "0" && <成交价格> != "0" && <成交金额> == "0"
  actions:
      vars:
          amount: <成交数量>.number * <成交价格>.number

- id: 拆分成交金额来源
  when: <成交金额> != "0" && <成交数量> == "0" && <成交价格> == "0"
  actions:
      ignore: true
```

## 槽位契约

槽位只有 Beancount 交易本身有的字段：`date`、`payee`、`narration`、`amount`、`currency`、`flag`、`tags`、`links`，以及 `metadata` 里的键。不要增加 `method`、`type` 这种槽位。支付方式、收/支写成元数据。没有该列的账单就不声明这个键。

```yaml
template:
  fileFormat: csv
  sourceHeaders: [交易时间, 交易对方, 商品, 收/支, 金额(元), 支付方式, 交易单号]
  slots:
    date: <交易时间>
    payee: <交易对方>
    narration: <商品>
    amount: <金额(元)>.number
    currency: CNY
    metadata:
      method: <支付方式>
      type: <收/支>
      orderId: <交易单号>
  amountSign:
    metadata: type
    negate: [支出, 支]
```

`amountSign` 是旧写法，仍然可用；新模板请写 `direction`（见下）。个人规则是唯一写 `from` / `to` 的地方。某一侧没写时，支出用 `Assets:FIXME` 和 `Expenses:FIXME`，收入用 `Income:FIXME` 和 `Assets:FIXME`。

### direction：金额方向

方向是输入，不是输出，所以单独声明，不从 metadata 反推。三种形态任选一种：

```yaml
template:
  direction:                       # 一列的取值决定方向
    column: <收/支>
    outflow: [支出, /]              # 命中为流出（支出）
    inflow: [收入]                  # 命中为流入；都没命中时看金额自身正负
```

```yaml
template:
  direction:                       # 支出、收入各一列（银行流水常见）；此时 slots.amount 可省略
    outflowColumn: <支出金额>
    inflowColumn: <收入金额>
```

```yaml
template:
  direction: {}                    # 不声明：金额自带正负，负数为流出
  direction: { invert: true }      # 信用卡账单常见：正数是消费（流出）
```

列值两边都没命中时默认看金额正负；加 `default: outflow` 可以把「不计收支」这类值一律按流出处理。

### 槽位模式下的 templateRules

声明了 `slots` 之后，`templateRules` 仍然可以用，但只能整理**字段**：按条件改 `payee`、`narration`、`date`、`currency`、增删 `metadata`、设置 `vars`、`ignore`。写 `from`/`to`/`postings`/`accounts` 会被拒绝。它们在槽位映射之后、legs 分支选择之前执行：

```yaml
templateRules:
  - id: 对手方为空时用银行名
    when: <对手信息> == "" || <对手信息> == "--"
    actions:
      payee: ABC Debit
  - id: 外币
    when: <币别> != "人民币"
    actions:
      currency: USD
```

元数据键的输出顺序以 `slots.metadata` 的声明顺序为准；规则补上的键如果在声明里（值可以写空串占位），会落到声明的位置。

### vars：模板变量

`template.vars` 是模板级变量，规则和 legs 里用 `<var.名字>` 引用。它只能放币种符号、计算出的金额这类值，**不能放账户名**（校验会拒绝）。可以带条件，后面的覆盖前面的：

```yaml
template:
  vars:
    - vars:
        security: SZ<证券代码>.format("%06.0f")
        fee: <手续费>.number + <印花税>.number
    - when: <股东账号> ~ "A"
      vars:
        security: SH<证券代码>.format("%06.0f")
```

### legs：多腿交易按角色声明

证券、交易所、链上转账一笔不止两条腿。模板只声明每条腿的**角色**和金额，账户留给用户绑定：

```yaml
template:
  legs:
    - id: 买入
      when: <操作> == "买"
      legs:
        - { role: cash,     amount: "-<var.amount>", currency: CNY }
        - { role: position, amount: '<成交数量>.format("%.2f")', currency: <var.security>, cost: '<成交价格>.format("%.3f") CNY', price: "@@ <var.amount> CNY" }
        - { role: cash,     amount: "-<var.fee>", currency: CNY }
        - { role: fee,      amount: <var.fee>,  currency: CNY }
    - id: 卖出
      when: <操作> == "卖"
      narration: 卖出-<证券名称>          # 分支可以覆盖 payee / narration / metadata
      legs:
        - { role: position, amount: '-<成交数量>.format("%.2f")', currency: <var.security>, cost: "{}", price: '@ <成交价格>.format("%.3f") CNY' }
        - { role: cash,     amount: <var.amount> }
        - { role: pnl }                   # 不写 amount：交给 Beancount 自动配平
```

- 分支按顺序取第一个 `when` 成立的；没有 `when` 的分支是默认分支；都不命中就退回普通的 `from`/`to` 两腿。
- `role` 是封闭核心集：`self` `other` `from` `to` `cash` `custody` `position` `fee` `gas` `pnl`。`other` 是对手方，方向与 `self` 相反。`self` 是这份账单自己的账户（银行卡、信用卡、钱包）：支出时它是 `from`，收入时它是 `to`，相当于 hledger 的 `account1`。普通账单在 `accounts:` 里只需绑 `self` 一个。机构特有的腿用 `x-` 前缀（如 `x-margin`），这类角色必须在规则文件里绑定，引擎不会补 FIXME。
- 腿里写 `account:` 会被拒绝。
- `cost` 自动加 `{}`，`price` 不带 `@` 时自动加 `@ `；写 `{}` 和 `@@ …` 都按原样保留。

### timezone：时区

账单里的时间通常是当地时间，没有时区。默认按运行机器的时区解析；写上 `timezone` 后，日期和 `.timestamp` 比较都按这个时区算，换台机器结果也一样：

```yaml
template:
  timezone: Asia/Shanghai
```

### 用户侧：accounts 绑定 + rules

用户的规则文件只需要两块。`accounts:` 把角色一次绑到自己的账户；`rules:` 用 `when` 覆盖个别交易（`personalRules:` 是同义写法）。

```yaml
template: abc_debit            # 普通银行卡：只绑 self

accounts:
  self: Assets:ABC:DebitCard
personalRules:
  - id: 房租
    when: narration ~ "房租"
    actions:
      to: Expenses:Housing:Rent
```

```yaml
template: htsec@2026-10-08     # 证券：按角色绑

accounts:
  cash: Assets:Htsec:Cash
  position: Assets:Htsec:Positions
  fee: Expenses:Htsec:Commission
  pnl: Income:Htsec:PnL

rules:
  - id: 忽略新增证券
    when: <证券名称> == "新增证券"
    actions:
      ignore: true
  - id: 兴业转债单独记
    when: <证券名称> ~ "兴业转债"
    actions:
      accounts:                        # 只对命中的交易改绑
        cash: Assets:Rule1:Cash
        position: Assets:Rule1:Positions
```

### 输出设置：自己的摘要、不要的元数据

规则文件的 `output:` 块决定每笔交易的默认写法，不用逐条规则去改：

```yaml
output:
  payee: <交易对方>
  narration: <交易类型>｜<商品>        # 用 <列名> 引用账单列
  metadata:
    drop: [orderId, merchantId]        # 不输出这些键
    # keep: [method, status]           # 或者只保留这些键（keep 优先于 drop）
```

- `payee` / `narration` 在模板映射之后、个人规则之前生效，所以个人规则仍然可以逐笔覆盖。
- `metadata` 过滤在最后执行，规则补上的键也会被过滤。
- `config init` 生成的文件里带着注释掉的 `output:` 示例，模板站的「输出设置」表单改的也是这一块。

### 对账：余额断言、去重、复核标记

交易都生成之后、写进账本之前，还有一道关。

**余额断言**（模板侧）：账单有余额列时声明出来，引擎逐行核对「上一行余额 ± 本行金额 = 本行余额」，账单按时间正序或倒序都能识别。版面文本正则漏读一行，这里立刻报错而不是默默出错账：

```yaml
template:
  balance:
    column: <余额>
```

**去重与标记**（用户侧，写在规则文件）：

```yaml
reconcile:
  dedupe:
    key: [metadata.orderId]        # 识别同一笔的字段：date / payee / narration / amount / metadata.<键>
    against: ./main.bean           # 已有账本；按 key 命中的交易不再写入
    window: 2d                     # 账本里 ±2 天内有同金额交易：标为 ! 供复核，不丢
  flagEngineFilled: "!"            # 引擎补过 FIXME 的交易也标 !
```

- `key` 的任一字段为空时不参与去重，不会因为信息不全而误删。
- 账单内部重复（同一 key 出现两次）只保留第一条。
- 导入结束会打印 `reconcile: N repeated in bill, N already in ledger, N flagged for review`。

跨来源的重复（微信里付的京东订单，京东账单里又出现一次）没有共同的订单号，只能靠 `window` 模糊匹配加人工复核，所以标 `!` 而不是直接丢。

普通收支用 `other:` 最省事：它是对手方，支出时记在 `to`，收入（含退款）时记在 `from`，一条规则同时覆盖消费和它的退款。`from:` / `to:` 是按方向写死的两侧，显式写了就优先于 `self` / `other`。

```yaml
accounts:
  self: Assets:Bank:CMB
personalRules:
  - id: 咖啡
    when: payee ~ "咖啡"
    actions:
      other: Expenses:Food:Coffee      # 消费：Coffee 增加；退款：Coffee 减少
  - id: 我发出的转账
    when: metadata.from == "0x1429…"
    actions:
      direction: outflow               # 只有用户知道哪个地址是自己的，规则可以指定方向
```没绑到的核心角色由引擎补 FIXME：`cash/custody/position → Assets:FIXME`，`fee/gas → Expenses:FIXME`，`pnl → Income:FIXME`，`from/to` 按方向。补的那一侧在来源记录（`Sources`）里标为 `engine`，网页端可以据此高亮。

```yaml
template: wechat@2026-04-28
personalRules:
  - id: 一卡通
    when: payee ~ "一卡通"
    actions:
      from: Assets:Current:零钱
  - id: 去掉订单号
    actions:
      metadataDrop: [orderId]
  - id: 补充备注
    actions:
      metadata:
        note: <备注>
```

`when` 里的 `payee`、`narration`、`amount` 和 `metadata.method` 读取的是映射之后的值。`<交易对方>` 仍然可以直接点名列。列名改了之后，只引用槽位的规则不用改；引用了被删列名或被删元数据键的规则，导入时会提示规则 `id`。

`double-entry-generator config init <template>` 会把槽位映射和元数据键写进个人文件的注释，并记下模板版本。规则 `id` 可以用中文。

## 使用建议

- 模板规则负责解释账单格式，个人规则负责账户分类。
- 规则 id 应描述“这段规则做什么”，例如 `午餐支出`、`HS300ETF账户覆盖`。
- 普通收支优先使用 `from` / `to`。
- 投资、手续费、换汇、币币交易等复杂场景优先使用 `postings`。
- 不要把 provider 私有概念放进 runtime；能用 `vars` 和 `postings` 表达的逻辑，应写在规则里。
