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
