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
