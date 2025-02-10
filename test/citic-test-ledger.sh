#!/usr/bin/env bash
#
# E2E test for citic provider.

# set -x # debug
set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."
OUTPUT="$ROOT_DIR/test/output/test-citic-credit-output.ledger"

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate citic bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider citic \
    --target ledger \
    --config "$ROOT_DIR/example/citic/credit/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/citic/credit/example-citic-records.xls"

diff -u --color \
    "$ROOT_DIR/example/citic/credit/example-citic-output.ledger" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] citic provider(credit mode) output is different from expected output."
    exit 1
fi

echo "[PASS] All citic provider for ledger target tests!"
