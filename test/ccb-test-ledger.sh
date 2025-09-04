#!/usr/bin/env bash
#
# E2E test for CCB provider.

# set -x # debug
set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."
OUTPUT="$ROOT_DIR/test/output/test-ccb-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate CCB bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider ccb \
    --target ledger \
    --config "$ROOT_DIR/example/ccb/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/ccb/交易明细_xxxx_2025xxxx_2025xxxx.xls"

diff -u --color \
    "$ROOT_DIR/example/ccb/example-ccb-output.ledger" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] CCB provider ledger output is different from expected output."
    exit 1
fi

echo "[PASS] All CCB provider tests for ledger target!"