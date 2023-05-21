#!/usr/bin/env bash
#
# E2E test for wechat provider.

# set -x # debug
set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."
OUTPUT="$ROOT_DIR/test/output/test-wechat-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate wechat bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider wechat \
    --target ledger \
    --config "$ROOT_DIR/example/wechat/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/wechat/example-wechat-records.csv"

diff -u --color \
    "$ROOT_DIR/example/wechat/example-wechat-output.ledger" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] WeChat provider output is different from expected output."
    exit 1
fi

echo "[PASS] All WeChat provider for ledger target tests!"
