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
  -> 模板解析字段
  -> 统一中间结构
  -> 模板规则
  -> 个人规则
  -> beancount / ledger
```

模板规则用于解释账单本身，例如收支方向、退款、手续费、状态等。个人规则用于表达用户自己的账户分类习惯，例如某个商户对应哪个支出账户。个人规则后执行，因此可以覆盖模板规则的结果。

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
- `>`、`>=`、`<`、`<=`
- `~` 包含
- `!~` 不包含
- `&&` 与
- `||` 或

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

## 使用建议

- 模板规则负责解释账单格式，个人规则负责账户分类。
- 规则 id 应描述“这段规则做什么”，例如 `午餐支出`、`HS300ETF账户覆盖`。
- 普通收支优先使用 `from` / `to`。
- 投资、手续费、换汇、币币交易等复杂场景优先使用 `postings`。
- 不要把 provider 私有概念放进 runtime；能用 `vars` 和 `postings` 表达的逻辑，应写在规则里。
