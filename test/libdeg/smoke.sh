#!/usr/bin/env bash
# Builds libdeg as a C archive, links a tiny C program against it and runs
# one import through the C API.
set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$root"
go build -buildmode=c-archive -o "$work/libdeg.a" ./cmd/libdeg
cc -o "$work/smoke" test/libdeg/smoke.c -I"$work" "$work/libdeg.a" -lpthread -lm -ldl
cat > "$work/template.yaml" <<'YAML'
schema: https://deg.dev/template-profile/v2
id: bank
template:
  fileFormat: csv
  dateFormat: yyyy-MM-dd
  defaultCurrency: CNY
  sourceHeaders: [日期, 商户, 金额]
  slots:
    date: <日期>
    payee: <商户>
    amount: <金额>.number
YAML
cat > "$work/rules.yaml" <<'YAML'
accounts:
  self: Assets:Bank:Card
  other: Expenses:Misc
YAML
printf '日期,商户,金额\n2026-01-02,咖啡馆,-30.50\n' > "$work/bill.csv"
out="$("$work/smoke" "$work/template.yaml" "$work/rules.yaml" "$work/bill.csv")"
echo "$out"
grep -q '"from":"Assets:Bank:Card","to":"Expenses:Misc","amount":"30.50"' <<<"$out"
echo "libdeg smoke: ok ($(du -h "$work/libdeg.a" | cut -f1) archive)"
