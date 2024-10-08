#!/usr/bin/env bash
#
# E2E test for icbc provider.

# set -x # debug
# set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."
CREDIT_OUTPUT="$ROOT_DIR/test/output/test-icbc-credit-output.ledger"
DEBIT_OUTPUT="$ROOT_DIR/test/output/test-icbc-debit-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate icbc credit bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider icbc \
    --target ledger \
    --config "$ROOT_DIR/example/icbc/credit/config.yaml" \
    --output "$CREDIT_OUTPUT" \
    "$ROOT_DIR/example/icbc/credit/example-icbc-credit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/icbc/credit/example-icbc-credit-output.ledger" \
    "$CREDIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] ICBC provider (credit mode) output is different from expected output."
    exit 1
fi

# generate icbc debit bills (v1) output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider icbc \
    --target ledger \
    --config "$ROOT_DIR/example/icbc/debit-v1/config.yaml" \
    --output "$DEBIT_OUTPUT" \
    "$ROOT_DIR/example/icbc/debit-v1/example-icbc-debit-v1-records.csv"

diff -u --color \
    "$ROOT_DIR/example/icbc/debit-v1/example-icbc-debit-v1-output.ledger" \
    "$DEBIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] ICBC provider (debit mode v1) output is different from expected output."
    exit 1
fi

# generate icbc debit bills (v1) output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider icbc \
    --target ledger \
    --config "$ROOT_DIR/example/icbc/debit-v2/config.yaml" \
    --output "$DEBIT_OUTPUT" \
    "$ROOT_DIR/example/icbc/debit-v2/example-icbc-debit-v2-records.csv"

diff -u --color \
    "$ROOT_DIR/example/icbc/debit-v2/example-icbc-debit-v2-output.ledger" \
    "$DEBIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] ICBC provider (debit mode v2) output is different from expected output."
    exit 1
fi

echo "[PASS] All ICBC provider tests!"
