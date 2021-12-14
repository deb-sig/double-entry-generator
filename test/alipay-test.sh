#!/usr/bin/env bash
#
# E2E test for alipay provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate alipay bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider alipay \
    --config "$ROOT_DIR/example/alipay/config.yaml" \
    --output "$ROOT_DIR/test/output/test-alipay-output.beancount" \
    "$ROOT_DIR/example/alipay/example-alipay-records.csv"

diff -u --color \
    "$ROOT_DIR/example/alipay/example-alipay-output.beancount" \
    "$ROOT_DIR/test/output/test-alipay-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] alipay provider output is different from expected output."
    exit 1
fi

echo "[PASS] All alipay provider tests!"
