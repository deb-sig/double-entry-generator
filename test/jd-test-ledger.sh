#!/usr/bin/env bash
#
# E2E test for jd provider.

# set -x # debug
set -eo errexit

TEST_DIR=$(dirname "$(realpath $0)")
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"
OUTPUT="$ROOT_DIR/test/output/test-jd-output.ledger"

# generate jd bills output in ledger format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider jd \
    --target ledger \
    --config "$ROOT_DIR/example/jd/config.yaml" \
    --output "$OUTPUT" \
    "$ROOT_DIR/example/jd/example-jd-records.csv"

diff -u --color \
    "$ROOT_DIR/example/jd/example-jd-output.ledger" \
    "$OUTPUT"

if [ $? -ne 0 ]; then
    echo "[FAIL] jd provider output is different from expected output."
    exit 1
fi

echo "[PASS] All jd provider for ledger target tests!"
