#!/usr/bin/env bash
#
# E2E test for bmo provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."
CREDIT_OUTPUT="$ROOT_DIR/test/output/test-bmo-credit-output.ledger"
DEBIT_OUTPUT="$ROOT_DIR/test/output/test-bmo-debit-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate bmo bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider bmo \
    --target ledger \
    --config "$ROOT_DIR/example/bmo/credit/config.yaml" \
    --output "$CREDIT_OUTPUT" \
    "$ROOT_DIR/example/bmo/credit/example-bmo-records.csv"

diff -u --color \
    "$ROOT_DIR/example/bmo/credit/example-bmo-output.ledger" \
    "$CREDIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] bmo provider(credit mode) output is different from expected output."
    exit 1
fi

"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider bmo \
    --target ledger \
    --config "$ROOT_DIR/example/bmo/debit/config.yaml" \
    --output "$CREDIT_OUTPUT" \
    "$ROOT_DIR/example/bmo/debit/example-bmo-records.csv"

diff -u --color \
    "$ROOT_DIR/example/bmo/debit/example-bmo-output.ledger" \
    "$CREDIT_OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] bmo provider(credit mode) output is different from expected output."
    exit 1
fi
echo "[PASS] All bmo provider tests!"
