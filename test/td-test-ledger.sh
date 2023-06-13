#!/usr/bin/env bash
#
# E2E test for td provider.

# set -x # debug
set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."
OUTPUT="$ROOT_DIR/test/output/test-td-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate td bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider td \
    --target ledger \
    --config "$ROOT_DIR/example/td/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/td/example-td-records.csv"

diff -u --color \
    "$ROOT_DIR/example/td/example-td-output.ledger" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] td provider output is different from expected output."
    exit 1
fi

echo "[PASS] All td provider for ledger target tests!"
