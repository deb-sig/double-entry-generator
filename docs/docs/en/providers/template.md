---
title: Generic Template Provider
description: Import bills with runtime templates and rules
---

# Generic Template Provider

Traditional providers require new Go packages, provider registration, analyser/config logic, provider-specific documentation, and tests inside DEG. As the number of providers grows, this increases both contribution cost and review cost: each provider PR may introduce a new private logic path that maintainers need to inspect and verify.

The generic template provider separates bill format support from DEG itself. DEG reads bills, interprets templates, runs rules, and writes beancount/ledger output. Bill-specific formats live in the template registry. This makes provider updates lighter and lets DEG development focus on the import engine, rule model, output behavior, and user experience.

## Flow

```text
bill file
  -> template field parsing
  -> normalized intermediate row
  -> template rules
  -> personal rules
  -> beancount / ledger
```

Template rules describe the bill format itself, such as income/expense direction, refunds, fees, status, and metadata. Personal rules describe the user's own account mapping habits. Personal rules run after template rules, so they can override template results.

## List Templates

```bash
double-entry-generator template list
```

Search templates by keyword:

```bash
double-entry-generator template search wechat
double-entry-generator template search payment
```

When the search result has a single match, DEG expands it with category, tags, and pinned versions.

## Generate Personal Rules

```bash
double-entry-generator config init wechat -o wechat-rules.yaml
```

Pin a template version:

```bash
double-entry-generator config init wechat@2026-04-28 -o wechat-rules.yaml
```

## Import Bills

Use template rules only:

```bash
double-entry-generator import wechat bill.csv -o output.bean
```

Use personal rules:

```bash
double-entry-generator import wechat bill.csv --rules wechat-rules.yaml -o output.bean
```

Pin a template version:

```bash
double-entry-generator import wechat@2026-04-28 bill.csv --rules wechat-rules.yaml -o output.bean
```

Use a local template:

```bash
double-entry-generator import ./wechat.yaml bill.csv --rules wechat-rules.yaml -o output.bean
```

## Completion

Generate shell completion scripts:

```bash
double-entry-generator completion zsh
double-entry-generator completion bash
double-entry-generator completion powershell
```

Completion supports:

- `double-entry-generator import <TAB>`: fetch template candidates from the online registry.
- `double-entry-generator import wechat@<TAB>`: fetch pinned template versions.
- `double-entry-generator import wechat <TAB>`: complete bill files.
- `double-entry-generator import wechat --rules <TAB>`: complete personal rule YAML files.
- `double-entry-generator template <TAB>`: complete template commands.

`--rules` is optional. It appears as a flag candidate when the user types `-` or `--` and triggers completion.
