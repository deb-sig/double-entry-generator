# RFC: v3 导入器架构 —— 五层流水线

状态：已定稿两项决策，实施中。目标分支 `dev-v3`。

已决：

- `role` 用**封闭核心集** `from / to / cash / position / fee / pnl / gas / custody`，引擎校验；机构特有的腿用 `x-` 前缀扩展（如 `x-margin`），引擎不校验、不推 FIXME 类型，只按用户 `accounts:` 绑定。
- PDF 走 `pdftotext -layout`，CLI 接受这个系统依赖；浏览器端用 pdf.js 转文本。引擎不解析 PDF 二进制。
- json / xml 统一用 **XPath**（不用 JSONPath），一种查询语法覆盖两种格式。

进度：第 1 步 Reader 已落地（`pkg/reader`，csv/xlsx/xls/json/xml/text），模板 `reader:` 块可用，旧字段兼容。第 2 步 Shaper 已落地（`pkg/importer/shape.go`，`shape:` 列表：`locateHeader / dropMatching / dropIf / merge / split`），没写 `locateHeader` 时沿用旧表头字段。第 3、4 步已落地：PR #2/#3 的槽位与来源记录合入，加 `template.direction`（三形态）、`template.vars`（可带条件）、`template.legs`（角色分支）、用户侧 `accounts:` 绑定与 `rules:`；模板仓库 htsec（legs）和 wechat（direction）已迁移，输出与原 expected 一致。模板仓库 26 家全部迁到槽位契约（新 pin `2026-10-08`，旧 pin 保留旧格式仍可导入），迁移过程中补了 `self` 角色（账单自己的账户，按方向落到 from/to）、`direction.invert/default`、槽位模式下仅限字段的 `templateRules`。第 5–7 步已落地：`html` / `eml`（容器，`part` + `inner`）/ `api`（`{source}`、`{env.X}`）Reader；`template.balance.column` 余额断言；`reconcile.dedupe`（精确 key、已有账本、`window` 模糊标记）与 `flagEngineFilled`。

## 要解决什么

把上游 82 个 issue 按原因归类，账单相关的只有五种：

| 原因 | 代表 issue | 现状为什么扛不住 |
|---|---|---|
| 格式漂移：列顺序、表头行数、新增交易类型 | #236 #217 #172 #135 #132 #84 #138 #33 #29 #28 | `skipLeadingRows` 写死；交易类型枚举写在 Go 代码里 |
| 文件格式：PDF、EML、XML、JSON、链上 API | #142 #147（中行 PDF）#186 | `ParseFile` 只有 csv/xls/xlsx 三个 case |
| 多腿交易：证券、交易所、手续费、换汇 | #14 #78 #186（gas） | 账户名写进模板的 `vars`，用户换账户要改模板 |
| 分类规则表达力 | #214 #129 #31 #37 #10 #4 | 逐个加字段，没有统一的条件模型 |
| 对账：重复、失败状态、退款 | #147 #123 #178 #82 #12 #71 | 没有去重层，退款逻辑散在各 provider |

每新增一家机构就加一个 Go 包，是这些问题的共同根源。v3 的方向已经定了：机构差异进模板仓库，DEG 本体只留引擎。本文回答的是**引擎该切成哪几层，每层的输入输出是什么**，使得上面五类问题各有一层去接。

## 一句话

```text
来源 ──读取──▶ 表格 ──整形──▶ 记录 ──映射──▶ 交易骨架 ──分类──▶ 完整交易 ──对账──▶ 账本
        Reader       Shaper        Mapper          Classifier        Reconciler
        模板          模板           模板             个人规则            个人规则+已有账本
```

五层，每层一个固定的数据形状衔接，每层只回答一个问题：

| 层 | 回答的问题 | 输入 | 输出 | 谁维护 |
|---|---|---|---|---|
| Reader 读取 | 字节怎么变成表格 | 文件 / URL / API | `Table` | 模板 |
| Shaper 整形 | 哪些行是一笔交易 | `Table` | `[]Record` | 模板 |
| Mapper 映射 | 这笔交易的 Beancount 字段是什么 | `Record` | `Transaction`（骨架） | 模板 |
| Classifier 分类 | 账户归谁 | `Transaction` | `Transaction`（完整） | 用户 |
| Reconciler 对账 | 该不该写进账本 | `[]Transaction` + 已有账本 | 账本 | 用户 |

业界三家都是这个分法，只是没把它说成层：hledger 的 `fields`（映射）和 `if`（分类）分开；beangulp 的 `identify/extract`（读取+映射）和用户钩子（分类）分开；Firefly 的 role（映射）和 rules（分类）分开。没有一家让格式模板写账户名。

下面逐层定契约。

## 第一层 Reader：一切皆表格

契约：无论什么格式，Reader 的输出都是

```go
type Table struct {
    Headers []string      // 可能为空，由 Shaper 定位
    Rows    [][]string    // 单元格全是字符串，不做任何类型解释
    Source  string        // 文件名或 URL，供报错和 provenance
}
```

这个形状故意贫瘠。Reader 不知道什么是金额、日期、交易，只负责把结构化内容铺平。好处是后面四层对所有格式一视同仁：PDF 账单和 CSV 账单用同一套模板语法，区别只在 `reader:` 块。

模板里的写法：

```yaml
reader:
  format: csv            # csv | xlsx | xls | json | xml | text | eml | api
  encoding: gb18030
  delimiter: ","
```

各格式需要的额外参数：

**表格类（csv / xlsx / xls）**：现有参数照搬。`xlsx` 加 `sheet`（名字或序号，缺省第一张）。

**树形类（json / xml）**：树怎么变成表格，用两个参数说清楚。两种格式都用 XPath。

```yaml
reader:
  format: xml
  records: "//Ntry"                 # 每个匹配节点是一行（XPath / JSONPath）
  columns:                          # 节点内相对路径 → 列名
    date: "BookgDt/Dt"
    amount: "Amt"
    currency: "Amt/@Ccy"
    direction: "CdtDbtInd"
    narration: "NtryDtls/TxDtls/RmtInf/Ustrd"
```

银行的 camt.053、OFX/QFX、交易所的 JSON 导出、链上 explorer 的 JSON 响应，全是这一种。

**文本类（text）**：面向 PDF 转出来的版面文本。#142 结论是 Go 里没有可靠的 PDF 库，所以**引擎不解析 PDF**，只解析带版面的纯文本。PDF 到文本的转换放在引擎外：

- CLI：调外部 `pdftotext -layout`，缺失时报错并提示安装。
- 网页（deg-online / site）：浏览器里跑 pdf.js，输出同样的版面文本。
- 用户手工转换后传 `.txt`。

三条路产出同一种文本，共用一个模板。Beancount 社区的 PDF 导入器几乎都是这个套路（`pdftotext -layout` + 正则），区别只是他们每家写一个 Python 文件，我们把正则放进模板。

```yaml
reader:
  format: text
  convert: pdftotext-layout         # 可选；输入是 pdf 时先转
  record: '^(?P<date>\d{4}-\d{2}-\d{2})\s+(?P<narration>.+?)\s+(?P<amount>-?[\d,]+\.\d{2})\s+(?P<balance>[\d,]+\.\d{2})$'
  continuation: '^\s{20,}(?P<narration>\S.*)$'   # 可选；缩进行并入上一条的 narration
```

命名分组即列名。`continuation` 处理 PDF 常见的摘要换行。版面文本正则脆弱是公认的，第五层的余额断言用来兜底：`balance` 列解析出来后逐行核对，对不上就报错而不是默默出错账。

**容器类（eml / zip）**：不是账单格式，是装账单的壳。参数只有「取哪一部分」加「里面用什么 Reader」。

```yaml
reader:
  format: eml
  part: "text/html"                 # 或 attachment:*.pdf
  inner:
    format: text
    record: ...
```

Reader 可以嵌套一层。交行 EML 里套 HTML 表格，招行邮件里套 PDF，都能写。

**接口类（api）**：#186 的链上地址。`import` 的第二个参数不再限定是文件，可以是 `eth:0x...`。Reader 按模板里的 endpoint 模板拉 JSON，然后走树形类的 `records/columns`。API key 从环境变量读，不进模板。

```yaml
reader:
  format: api
  url: "https://api.etherscan.io/v2/api?chainid={chain}&module=account&action=txlist&address={address}&apikey={env.ETHERSCAN_KEY}"
  records: "//result/*"
  columns:
    hash: hash
    time: timeStamp
    from: from
    to: to
    value: value
    gasUsed: gasUsed
    gasPrice: gasPrice
```

Reader 的 Go 接口：

```go
type Reader interface {
    Read(ctx context.Context, src Source, cfg ReaderConfig) (Table, error)
}
```

`Source` 是文件路径、`io.Reader` 或 URL。WASM 构建里 `api` 和 `pdftotext-layout` 不可用，模板的 `requiredCapabilities` 已经有了，就用它声明，网页端提前拒绝而不是中途崩。

## 第二层 Shaper：哪些行是一笔交易

契约：`Table → []Record`，`Record = map[string]string`（列名到值）。这一层处理所有「行」和「交易」不是一对一的情况。现状把这些事混在 `skipLeadingRows`、`skipInvalidRows`、`headerLocate` 和规则里的 `ignore: true` 四处。

整形是一列有序操作：

```yaml
shape:
  - locateHeader: { anchor: "交易时间", scanRows: 50 }   # 用锚点定位表头，不数行数 → 杀掉 #236 这类 bug
  - dropMatching: '^(导出|说明|共计|---)'                # 表头之前/之后的说明行，整行文本匹配
  - dropIf: '<当前状态> ~ "失败|撤销|已关闭"'            # #123 的失败还款、#71 的 0 元记录在这里过滤
  - merge:                                                # hxsec：一笔成交拆成「金额行」和「数量价格行」
      key: ["<合同号>", "<成交号>"]
      take: { 成交金额: first-nonzero, 成交数量: first-nonzero, 成交价格: first-nonzero }
  - split:                                                # 一行含两笔：本金+手续费各成一条记录
      when: '<手续费>.number > 0'
      into:
        - { }                                             # 原行
        - { 金额: "<手续费>", 交易类型: "手续费", _parent: "<交易单号>" }
```

`locateHeader` 用锚点而不是行数，是对格式漂移最有效的一招：微信、支付宝、建行历次改版，大多只是说明行多了少了一行。

`merge` 和 `split` 让多行一笔、一行多笔都在进入映射层之前变成「一条 Record 就是一笔交易」。后面的层永远不用再操心行的问题。

这一层不接触账户、金额含义，所以它和 Mapper 一样属于模板，不属于用户。

## 第三层 Mapper：Beancount 字段是什么

契约：`Record → Transaction`。`Transaction` 只有 Beancount 交易本身有的槽：

```go
type Transaction struct {
    Date      time.Time
    Payee     string
    Narration string
    Flag      string
    Tags, Links []string
    Metadata  []KV                 // 有序
    Legs      []Leg                // 见下
    Provenance []FieldSource       // PR #3，每个值从哪来
}
type Leg struct {
    Role     string                // "from" "to" | "cash" "position" "fee" "pnl" "gas" ...
    Account  string                // 模板不填，留空；由 Classifier 填
    Amount   Decimal               // 精确小数，ExactMoney 已经做完
    Currency string
    Cost, Price *Amount            // {} 和 @，证券/币用
}
```

这就是 PR #2 的槽位契约，加两处修正：

**1. `direction` 代替 `amountSign.metadata`。** 金额正负是输入，不是输出；用 metadata 的值反推方向，会让「删掉某个 metadata 键」这个用户动作影响金额方向。改成显式字段：

```yaml
map:
  date: "<交易时间>"
  payee: "<交易对方>"
  narration: "<商品>"
  amount: "<金额(元)>.number"
  currency: CNY
  direction:                       # 三选一
    column: "<收/支>"
    outflow: ["支出", "支"]
    inflow: ["收入", "收"]
  # 或 direction: { sign: amount }           金额自带正负
  # 或 direction: { outflow: "<支出金额>", inflow: "<收入金额>" }   两列各一边（银行流水常见）
  metadata:
    method: "<支付方式>"
    orderId: "<交易单号>"
    status: "<当前状态>"
```

三种 direction 形态覆盖了全部 26 个现有模板。

**2. 多腿交易用「角色」而不是账户。** 这是现状最大的问题：htsec/huobi/hxsec 把 `Assets:Hxsec:Cash` 写在模板 `vars` 里，用户换账户名必须复制整段模板规则。改为模板只声明腿的**角色**和金额，账户留空：

```yaml
map:
  date: "<成交日期> <成交时间>"
  narration: "<操作>-<证券名称>"
  legs:
    - { role: cash,     amount: "-<成交金额>", currency: CNY }
    - { role: position, amount: "<成交数量>", currency: "<var.security>", cost: "{<成交价格> CNY}", price: "@@ <成交金额> CNY" }
    - { role: cash,     amount: "-<手续费>", currency: CNY }
    - { role: fee,      amount: "<手续费>",  currency: CNY }
  when: '<操作> == "证券买入"'        # map 可以有多个分支，按 when 选第一个命中的
```

`role` 是封闭核心集 `from / to / cash / position / fee / pnl / gas / custody`，引擎校验拼写并据此推 FIXME 的账户类型；机构特有的腿用 `x-` 前缀（如 `x-margin`），只做绑定不做推断。普通收支就是两条腿 `from` 和 `to`，是 `legs` 的特例，所以 `direction` + `amount` 这种简写在引擎内部展开成两条腿。

映射层允许多个分支（证券买入、卖出、红利、转账各一个），`when` 用现有的条件语法。这就是现在 `templateRules` 在做的事，但产物里没有账户。

## 第四层 Classifier：账户归谁

契约：`Transaction → Transaction`，只允许改账户、metadata、tags、flag，以及 `ignore`。这是用户唯一要写的文件。

```yaml
template: wechat@2026-04-28
accounts:                          # 角色 → 账户的默认绑定，一次写完
  from: Assets:Digital:Wechat:Cash
  cash: Assets:Broker:Htsec:Cash
  position: Assets:Broker:Htsec:Positions
  fee: Expenses:Broker:Commission
  pnl: Income:Broker:PnL
rules:
  - id: 餐饮
    when: payee ~ "美团|饿了么"
    to: Expenses:Food
  - id: 一卡通
    when: payee ~ "一卡通"
    from: Assets:Current:零钱
  - id: 零钱通转入
    when: metadata.txType == "转入零钱通"
    to: Assets:Digital:Wechat:MiniFund
  - id: 内部转账不记
    when: metadata.method ~ "招商银行"       # #147：这笔会从银行账单那边进来
    ignore: true
  - id: 去掉订单号
    metadataDrop: [orderId]
```

`accounts` 是角色绑定：一行解决「用户换账户」。`rules` 用 `when` 覆盖个别交易。规则里写 `to:` 就是给 `role: to` 的腿绑账户，写 `cash:` 就是给 `role: cash` 的腿绑账户，所以普通收支和证券用同一套语法。

没绑到账户的腿由引擎补 `Assets:FIXME` / `Expenses:FIXME` / `Income:FIXME`（按角色和方向推），并在 provenance 里标 `engine`，网页端能高亮出来。这是 hledger 的 `unknown` 账户和 beangulp 的做法。

`when` 读的是映射后的槽（`payee`、`narration`、`amount`、`metadata.x`），也允许 `<列名>` 直接点名列。列名改了，只引用槽的规则不用动。

`ignore` 粘性（Mirato 语义，一旦为真不再清除）保留。`postingsMode`、`vars`、模板级 `postings` 不再对用户暴露；需要自定义腿的极端场景，用户规则里允许 `legs:` 整体替换，和模板用一样的写法。

## 第五层 Reconciler：该不该写进账本

契约：`[]Transaction + 已有账本 → 账本`。现状完全没有这一层，#147 的用户在手删重复。

三件事，都是可选开关：

```yaml
reconcile:
  dedupe:
    key: [metadata.orderId]                     # 强键：同 orderId 就是同一笔
    fuzzy: { window: 2d, amount: exact, narration: 0.8 }   # 弱键：跨来源（微信 vs 银行卡）
    against: ./main.bean                         # 已有账本；默认不去重
  balance:
    column: "<余额>"                              # 有余额列就逐行断言，对不上报错
    account: from
  flagUncertain: "!"                            # 模糊命中、FIXME 未补的交易标 ! 而不是 *
```

`dedupe.against` 读已有账本是 beangulp `extract(existing)` 的做法。跨来源去重（微信里付的京东订单，京东账单里又出现一次）靠模糊键加人工复核，所以标 `!` 而不是直接丢。

`balance` 断言是 PDF 文本正则的安全网：版面解析错一行，余额立刻对不上。

输出端（Beancount / Ledger 编译器）不动，仍然吃 `ir.IR`。`Transaction → ir.Order` 的转换是一个适配函数，`Legs` 渲染成 `Postings`，老编译器一行不改。

## 每类来源怎么走这五层

| 来源 | Reader | Shaper | Mapper | 特别之处 |
|---|---|---|---|---|
| 支付平台（微信、支付宝、京东、美团、拼多多 #234） | csv/xlsx | locateHeader + dropIf 失败状态 | direction.column | 退款：`when: metadata.status ~ "退款"` 分支把 from/to 对调，不再写进 Go |
| 银行借记/信用（15 家） | csv/xlsx/xls；PDF 走 text；EML 套 text | locateHeader | direction 两列或 sign | 余额列 → balance 断言 |
| 证券（htsec、hxsec、海外券商） | xlsx/csv | merge 拆分行 | legs 多分支（买/卖/红利/转账） | role: cash/position/fee/pnl |
| 交易所（huobi、币安 CSV/JSON） | csv/json | — | legs，手续费币种分支 | 精度：Decimal 已就位 |
| 链上（oklink CSV、explorer API #186） | csv 或 api | dropIf 失败交易 | legs：转账 + gas 腿 | role: gas，账户由用户绑 |
| 标准格式（OFX、camt.053） | xml | — | 一个通用模板即可 | 一个模板覆盖所有用这种格式的银行 |

## 和 dev-v3 现状的关系

不推倒。dev-v3 已经做对的：模板仓库 + 版本 pin、`<列>.方法` 表达式、条件语法、ExactMoney、provenance（PR #3）、Mirato 互通的规则文件交换。这些全保留。

改的是切分：

| 现状 | 去向 |
|---|---|
| `template.fileFormat/encoding/delimiter/...` | `reader:` 块，原样搬 |
| `skipLeadingRows / skipInvalidRows / headerLocate` | `shape:` 的 `locateHeader / dropWhile / dropIf` |
| 规则里 `ignore: true` 用来丢辅助行 | `shape: merge`，不再用规则做行处理 |
| `templateRules` 的 date/payee/narration/amount/metadata | `map:` 槽位 |
| `templateRules` 的 `vars` + `postings` | `map.legs` 带 role，账户去掉 |
| `templateRules` 里的 `from/to` 账户 | 删除；用户 `accounts:` 绑定 |
| `personalRules` | `rules:`，语法基本不变，`from/to` 变成角色绑定的特例 |
| `amountSign.metadata`（PR #2） | `map.direction` |
| 无 | `reconcile:` |

旧的 `templateRules` 模板继续能导入（PR #2 已经留了这条路），模板仓库逐家迁移，每迁一家跑一次 `expected.beancount` 回归。

## 实施顺序

每步独立可合并，每步结束模板仓库 26 家的 `expected.beancount` 不变。

1. **抽出 Reader**：`pkg/reader`，`Table` 类型，csv/xlsx/xls 搬进去，`ParseFile` 变成调度。顺手加 json/xml，因为树形 Reader 只有 100 行。
2. **Shaper**：`locateHeader` 锚点版替换 `skipLeadingRows`（修 #236 一类），`dropIf`、`merge`。hxsec 模板改用 `merge`，去掉用规则丢行的写法。
3. **Mapper 槽位**：合入 PR #2、#3，`amountSign` 改 `direction`，`legs` + `role` 替换模板里的 `vars/postings`。`Transaction` 类型落地，加 `Transaction → ir.Order` 适配。
4. **Classifier**：`accounts:` 角色绑定；htsec/huobi/hxsec 三家模板迁移，用户规则文件从 60 行缩到 10 行。
5. **text / eml Reader + pdftotext**：先做招行 PDF 一家验证版面正则 + 余额断言；网页端接 pdf.js。
6. **Reconciler**：`dedupe.against` 和 `balance`。
7. **api Reader**：以太坊 explorer 一条链验证，gas 腿。

1–4 是主线，决定模板格式，先做；5–7 各自独立，谁有需求谁先做。

## 不做的事

- 不在 Go 里解析 PDF 二进制。#142 已经验证过所有 Go 库，没有能用的；版面文本是稳定的中间格式。
- 不做机器学习分类。规则 + FIXME + `!` 标记的人工复核，对个人记账够用，且可解释。
- 不让模板写账户名。这是整个设计的硬约束。
- 不给 Reader 加语义参数（什么是金额、日期）。Reader 只认结构。

## 参考

- hledger CSV rules：`fields` 映射与 `if` 分类分离，`account3/amount3` 多腿，`balance` 断言。https://hledger.org/csv.html
- beangulp：`identify / account / extract(existing)` 接口，去重靠已有账本。https://github.com/beancount/beangulp
- Firefly III data importer：列 role 词表 + 独立 rules。https://docs.firefly-iii.org/references/data-importer/roles/
- Beancount 社区 PDF 导入：`pdftotext -layout` + 正则 + 余额核对。https://beancount.io/forum/t/bank-statement-pdfs-to-beancount-ocr-and-parsing-for-the-automation-averse/229
- 上游 issue #142（PDF/EML）、#147（去重）、#186（链上）、#236（表头行数）。
