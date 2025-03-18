#!/usr/bin/env bash
#
# E2E test for cmb provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."
CREDIT_OUTPUT="$ROOT_DIR/test/output/test-cmb-credit-output.beancount"
DEBIT_OUTPUT="$ROOT_DIR/test/output/test-cmb-debit-output.beancount"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate cmb bills output in beancount format

# credit
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider cmb \
    --config "$ROOT_DIR/example/cmb/credit/config.yaml" \
    --output "$CREDIT_OUTPUT" \
    "$ROOT_DIR/example/cmb/credit/example-cmb-records.csv"

diff -u --color \
    "$ROOT_DIR/example/cmb/credit/example-cmb-output.beancount" \
    "$CREDIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] cmb provider(credit mode) output is different from expected output."
    exit 1
fi

# debit
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider cmb \
    --config "$ROOT_DIR/example/cmb/debit/config.yaml" \
    --output "$DEBIT_OUTPUT" \
    "$ROOT_DIR/example/cmb/debit/example-cmb-records.csv"

diff -u --color \
    "$ROOT_DIR/example/cmb/debit/example-cmb-output.beancount" \
    "$DEBIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] cmb provider(debit mode) output is different from expected output."
    exit 1
fi
echo "[PASS] All cmb provider for beancount target tests!"
