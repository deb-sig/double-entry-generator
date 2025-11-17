#!/usr/bin/env bash
#
# E2E test for Bocom Credit provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate Bocom Credit bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider bocom_credit \
    --target ledger \
    --config "$ROOT_DIR/example/bocomcredit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-bocom_credit-output.ledger" \
    "$ROOT_DIR/example/bocomcredit/example-bocom_credit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/bocomcredit/example-bocom_credit-output.ledger" \
    "$ROOT_DIR/test/output/test-bocom_credit-output.ledger"

if [ $? -ne 0 ]; then
    echo "[FAIL] Bocom Credit provider output is different from expected output."
    exit 1
fi

echo "[PASS] All Bocom Credit provider ledger tests!"
