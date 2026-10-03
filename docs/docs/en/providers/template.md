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

## Template Files

Template files describe how a bill is read:

```yaml
schema: https://deg.dev/template-profile/v2
id: htsec
name: htsec
template:
    fileFormat: xlsx          # csv / xlsx / xls
    encoding: gb18030         # optional text encoding
    delimiter: ','            # use "\t" for tab-delimited text
    skipLeadingRows: 1
    skipInvalidRows: true
    sourceHeaders:
        - date
        - time
        - amount
        - description
    defaultCurrency: CNY
```

Headers become rule fields and can be referenced as `<amount>` or `<description>`.

## Rule Files

Rules usually contain:

```yaml
options:
    title: My Ledger
    operatingCurrency: CNY

templateRules:
    - id: Base transaction
      actions:
          date: <date> <time>
          payee: <peer>
          narration: <description>

personalRules:
    - id: Food expense
      when: <peer> ~ "Restaurant"
      actions:
          to:
              account: Expenses:Food
```

`templateRules` describe the bill format. `personalRules` describe user-specific account choices. Rules run in order, and later matched rules can override earlier fields or variables.

## Conditions

```yaml
when: <type> == "expense"
when: <amount>.number >= 10
when: <peer> ~ "Restaurant"
when: <peer> !~ "Refund"
when: (<side> == "buy" || <side> == "sell") && <market> == "spot"
```

Supported operators: `==`, `!=`, `>`, `>=`, `<`, `<=`, `~`, `!~`, `&&`, `||`.

## Field Methods

```yaml
<amount>.number
<amount>.+
<amount>.-
<amount>.!
<code>.format("%06.0f")
<fee>.extract("^([.0-9]+)")
raw[created_at].time
```

- `.number`: normalize amount text.
- `.+`: force positive.
- `.-`: force negative.
- `.!`: invert sign.
- `.extract("regex")`: extract text with a regular expression.
- `.format("...")`: format a number or string.
- `.date`, `.time`, `.timestamp`: extract date/time values.

Simple arithmetic is supported:

```yaml
<quantity>.number * <price>.number
<fee>.number + <tax>.number
```

## Actions

Core actions:

```yaml
actions:
    date: <created_at>
    payee: <peer>
    narration: <description>
    note: <note>
    amount: <amount>.number
    currency: CNY
    from:
        account: Assets:Bank
    to:
        account: Expenses:Food
    metadata:
        orderId: <order_id>
    tags:
        - Food
    vars:
        cash: Assets:Broker:Cash
    postings:
        - <var.cash> -<amount>.format("%.2f") CNY
    ignore: true
```

Use `from` / `to` for ordinary two-posting transactions:

```yaml
- id: Default expense
  when: <type> == "expense"
  actions:
      from:
          account: Assets:FIXME
      to:
          account: Expenses:FIXME
      amount: <amount>.number
      currency: CNY
```

Use explicit `postings` for complex transactions:

```yaml
- id: Securities buy
  when: <businessType> == "buy"
  actions:
      narration: Buy-<code>.format("%06.0f")-<name>
      postings:
          - <var.cash> -<tradeAmount>.format("%.2f") CNY
          - <var.position> <quantity>.format("%.2f") <var.security> {<price>.format("%.3f") CNY} @@ <tradeAmount>.format("%.2f") CNY
          - <var.cash> -<fee>.format("%.2f") CNY
          - <var.feeExpense> <fee>.format("%.2f") CNY
```

## Vars

`vars` are rule-local variables for accounts, units, and intermediate values. They are not IR fields.

```yaml
templateRules:
    - id: Default broker vars
      actions:
          vars:
              cash: Assets:Broker:Cash
              position: Assets:Broker:Positions
              feeExpense: Expenses:Broker:Commission
              security: SH<code>.format("%06.0f")

personalRules:
    - id: HS300ETF account override
      when: <name> ~ "HS300ETF"
      actions:
          vars:
              cash: Assets:Rule1:Cash
              position: Assets:Broker:Positions:HS300
```

Later postings such as `<var.cash>` use the overridden value.

## Ignore Helper Rows

Some bills split one business event across helper rows. If the final postings can be derived from the main row, ignore the helper row and compute the missing value with rules:

```yaml
- id: Split trade target
  when: <quantity> != "0" && <price> != "0" && <amount> == "0"
  actions:
      vars:
          amount: <quantity>.number * <price>.number

- id: Split trade amount helper
  when: <amount> != "0" && <quantity> == "0" && <price> == "0"
  actions:
      ignore: true
```

## Recommendations

- Keep bill-format behavior in `templateRules`.
- Keep user account choices in `personalRules`.
- Prefer descriptive rule ids.
- Use `from` / `to` for simple cash flows.
- Use `postings` for investments, fees, FX, and crypto trades.
- Express provider-specific behavior in rules instead of adding runtime-specific actions.
