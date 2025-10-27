#!/usr/bin/env bash
#
# E2E test for bocom debit provider (beancount output).

set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."
OUTPUT="$ROOT_DIR/test/output/test-bocom-debit-output.beancount"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider bocom_debit \
    --config "$ROOT_DIR/example/bocom_debit/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/bocom_debit/example-bocom-debit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/bocom_debit/example-bocom-debit-output.beancount" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] bocom_debit provider beancount output is different from expected output."
    exit 1
fi
echo "[PASS] bocom_debit provider beancount tests!"
