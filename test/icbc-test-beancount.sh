#!/usr/bin/env bash
#
# E2E test for icbc provider.

# set -x # debug
# set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate icbc credit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider icbc \
    --config "$ROOT_DIR/example/icbc/credit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-icbc-credit-output.beancount" \
    "$ROOT_DIR/example/icbc/credit/example-icbc-credit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/icbc/credit/example-icbc-credit-output.beancount" \
    "$ROOT_DIR/test/output/test-icbc-credit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] ICBC provider (credit mode) output is different from expected output."
    exit 1
fi

# generate icbc debit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider icbc \
    --config "$ROOT_DIR/example/icbc/debit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-icbc-debit-output.beancount" \
    "$ROOT_DIR/example/icbc/debit/example-icbc-debit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/icbc/debit/example-icbc-debit-output.beancount" \
    "$ROOT_DIR/test/output/test-icbc-debit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] ICBC provider (debit mode) output is different from expected output."
    exit 1
fi

echo "[PASS] All ICBC provider tests!"
