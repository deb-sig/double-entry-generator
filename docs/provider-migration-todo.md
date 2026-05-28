# Provider Migration TODO

Goal: migrate every legacy provider/example to the registry-backed runtime template flow with strict zero diff against the legacy `example/*/*output.beancount` golden files.

Completed:

- [x] `alipay`
- [x] `wechat`
- [x] `oklink`
- [x] `td`
- [x] `bmo-debit`
- [x] `bmo-credit`
- [x] `cmb-credit`
- [x] `cmb-debit`
- [x] `bocom_credit`
- [x] `bocom_debit`
- [x] `abc_debit`
- [x] `citic-credit`
- [x] `hsbchk-debit`
- [x] `hsbchk-credit`
- [x] `icbc-credit`
- [x] `icbc-debit-v1`
- [x] `icbc-debit-v2`
- [x] `boc-credit`
- [x] `boc-debit`
- [x] `jd`

Acceptance for each remaining item:

- `template.yaml` is runtime v2 style: source headers in the template, base parsing in `templateRules`, user-editable decisions in `personalRules`.
- No `columns` mapping is required for user-authored rules.
- Field references use `<源表头>` style.
- Legacy config candidate accounts are represented as static accounts in matching rules, not as a separate open-account list.
- `config init <provider>` writes only the personal rule skeleton.
- `import <provider> --rules <personal>.yaml <bill> --output <out>` matches the old example golden with strict byte-for-byte diff.
- Latest and dated pin produce identical output for the same migrated example.
- `pkg/regression/runtime_golden_test.go` uses the new registry/import path to compare against the old golden.
- `go test ./...` passes.

Migration recipe:

1. Read the old provider parser, analyser, `example/<provider>/config.yaml`, `template.yaml`, `rules.yaml`, bill file, and golden output.
2. Convert stable parsing fields into `templateRules`: date, amount, currency, payee, narration, metadata, tags, and provider-specific scalar actions.
3. Convert old rule conditions to source-header expressions using `<...>`.
4. Translate old candidate account behavior from `GetAllCandidateAccounts` and config rules into static rule accounts, even when the sample bill does not hit that rule.
5. Keep provider ids stable and English. Use variant ids when the old provider has multiple bill formats.
6. Add latest and dated pin entries to the template registry.
7. Add or extend regression coverage so the new runtime import compares to the old golden.
8. Run the command-path check: `config init`, `import`, `diff`, latest-vs-pin, then `go test ./...`.

Recommended order:

- [x] `td`
  - Example: `example/td`
  - Why first: simple CSV, English headers, small ruleset, no Excel reader behavior.
  - Watch: amount is split across `amountIn` / `amountOut`; keep CAD default currency and old candidate open accounts.

- [x] `bmo-debit`
  - Example: `example/bmo/debit`
  - Why early: simple CSV and similar to `td`.
  - Watch: provider id should be a registry variant such as `bmo-debit`; preserve `defaultCashAccount` open behavior.

- [x] `bmo-credit`
  - Example: `example/bmo/credit`
  - Why early: same provider family as debit, simple CSV.
  - Watch: credit-card cash/liability account and candidate accounts from unmatched rules.

- [x] `cmb-credit`
  - Example: `example/cmb/credit`
  - Why early: small Chinese CSV ruleset.
  - Watch: ignored `财付通` rows and credit-card default cash account.

- [x] `cmb-debit`
  - Example: `example/cmb/debit`
  - Why early: small Chinese CSV ruleset.
  - Watch: income/expense direction and `Income:Insurance` candidate open.

- [x] `bocom_credit`
  - Example: `example/bocom_credit`
  - Why early: small CSV rule set.
  - Watch: amount/currency parsing from combined amount fields.

- [x] `bocom_debit`
  - Example: `example/bocom_debit`
  - Why early: CSV, ordinary bank statement rules.
  - Watch: debit/credit type mapping and candidate accounts from all config rules.

- [x] `abc_debit`
  - Example: `example/abc_debit`
  - Why early: CSV and ordinary debit-card mapping.
  - Watch: date + time merge and interest/tax transfer rules.

- [x] `citic-credit`
  - Example: `example/citic/credit`
  - Why mid: rules are straightforward but the bill is `.xls`.
  - Watch: xls parsing compatibility, transaction amount sign, and ignored payment-channel rows.

- [x] `hsbchk-debit`
  - Example: `example/hsbchk/debit`
  - Why mid: CSV with English headers, moderate rules.
  - Watch: `stripTabs`, HKD currency, and ignored rows.

- [x] `hsbchk-credit`
  - Example: `example/hsbchk/credit`
  - Why mid: similar parser to debit but more credit-card semantics.
  - Watch: `Credit / Debit` direction, billing currency, and liability cash account.

- [x] `icbc-credit`
  - Example: `example/icbc/credit`
  - Why mid: CSV, but ICBC has several variants.
  - Watch: skipped invalid rows, split income/outcome columns, and ignored payment-channel rows.

- [x] `icbc-debit-v1`
  - Example: `example/icbc/debit-v1`
  - Why mid: same family as ICBC credit.
  - Watch: `skipInvalidRows`, typo-compatible old config behavior such as `txTpe`, and debit-card cash account.

- [x] `icbc-debit-v2`
  - Example: `example/icbc/debit-v2`
  - Why mid: close to v1, but extra source columns.
  - Watch: keep it as a separate dated/variant template so old exports remain pinned.

- [x] `boc-credit`
  - Example: `example/boc/credit`
  - Why later: more account candidates and method-account rules.
  - Watch: all `methodAccount` candidates must become static rule accounts for zero diff.

- [x] `boc-debit`
  - Example: `example/boc/debit`
  - Why later: similar to credit, but different source columns.
  - Watch: shared `boc` rules should not hide variant-specific parsing differences.

- [x] `jd`
  - Example: `example/jd`
  - Why later: CSV but many e-commerce category rules and candidate accounts.
  - Watch: skip-leading rows, closed/ignored transactions, Baitiao/liability account behavior.

- [x] `mt`
  - Example: `example/mt`
  - Why later: CSV but long rule list and time-window conditions.
  - Watch: date-time methods such as `date.time`, typo-preserving account names in old golden, and candidate accounts.

- [x] `spdb_debit`
  - Example: `example/spdb_debit`
  - Why later: relatively small rules, but `.xls` statement.
  - Watch: split income/outcome amount columns and category metadata.

- [x] `ccb`
  - Example: `example/ccb`
  - Why late: xls statement and very large candidate account surface.
  - Watch: many unmatched config accounts must still be represented as static rule accounts for strict zero diff.

- [x] `huobi`
  - Example: `example/huobi`
  - Why late: crypto/trading behavior rather than ordinary two-posting bank import.
  - Watch: buy/sell direction, trading pair metadata, PnL account, fees, and crypto order type formatting.

- [x] `htsec`
  - Example: `example/htsec`
  - Why late: securities statement with xlsx input and trade-specific postings.
  - Watch: commission, stamp tax, transfer fee, position/cash/PnL accounts, and security quantity/price formatting.

- [x] `hxsec`
  - Example: `example/hxsec`
  - Why last: securities xls input with the broadest fee/position surface.
  - Watch: gb18030/tab parsing, security position accounts, PnL account, all fee fields, and exact securities output formatting.

Notes:

- Do not add a generic `openAccounts` user-facing field unless a legacy behavior cannot be expressed as rules. The preferred migration form is rule-owned static accounts.
- If a legacy golden contains unused opens, first look for the old config/analyser candidate rule that produced them and migrate that rule.
- If a provider needs new expression/runtime features, add focused importer tests before migrating that provider.
