#!/usr/bin/env bash
#
# E2E test for ABC debit provider.

set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate ABC debit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider abc_debit \
    --config "$ROOT_DIR/example/abc_debit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-abc_debit-output.beancount" \
    "$ROOT_DIR/example/abc_debit/example-abc_debit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/abc_debit/example-abc_debit-output.beancount" \
    "$ROOT_DIR/test/output/test-abc_debit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] ABC debit provider output is different from expected output."
    exit 1
fi

echo "[PASS] All ABC debit provider beancount tests!"
