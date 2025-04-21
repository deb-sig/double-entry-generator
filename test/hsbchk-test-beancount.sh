#!/usr/bin/env bash
#
# E2E test for HSBC HK provider.

# set -x # debug
# set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate HSBC HK debit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider hsbchk \
    --config "$ROOT_DIR/example/hsbchk/debit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-hsbchk-debit-output.beancount" \
    "$ROOT_DIR/example/hsbchk/debit/example-hsbchk-debit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/hsbchk/debit/example-hsbchk-debit-output.beancount" \
    "$ROOT_DIR/test/output/test-hsbchk-debit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] HSBC HK provider (debit mode) output is different from expected output."
    exit 1
fi

# generate HSBC HK credit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider hsbchk \
    --config "$ROOT_DIR/example/hsbchk/credit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-hsbchk-credit-output.beancount" \
    "$ROOT_DIR/example/hsbchk/credit/example-hsbchk-credit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/hsbchk/credit/example-hsbchk-credit-output.beancount" \
    "$ROOT_DIR/test/output/test-hsbchk-credit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] HSBC HK provider (credit mode) output is different from expected output."
    exit 1
fi

echo "[PASS] All HSBC HK provider tests!" 