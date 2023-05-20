#!/usr/bin/env bash
#
# E2E test for huobi provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate huobi bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider huobi \
    --config "$ROOT_DIR/example/huobi/config.yaml" \
    --output "$ROOT_DIR/test/output/test-huobi-output.beancount" \
    "$ROOT_DIR/example/huobi/example-huobi-records.csv"

diff -u --color \
    "$ROOT_DIR/example/huobi/example-huobi-output.beancount" \
    "$ROOT_DIR/test/output/test-huobi-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] Huobi provider output is different from expected output."
    exit 1
fi

echo "[PASS] All Huobi provider tests!"
