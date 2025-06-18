#!/usr/bin/env bash
#
# E2E test for hxsec provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate hxsec bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider hxsec \
    --config "$ROOT_DIR/example/hxsec/config.yaml" \
    --target beancount \
    --output "$ROOT_DIR/test/output/test-hxsec-output.beancount" \
    "$ROOT_DIR/example/hxsec/example-hxsec-records.xls"

# Check if expected output file exists, create if not
EXPECTED_OUTPUT="$ROOT_DIR/example/hxsec/example-hxsec-output.beancount"
if [ ! -f "$EXPECTED_OUTPUT" ]; then
    echo "Expected output file $EXPECTED_OUTPUT not found. Creating it from generated output."
    cp "$ROOT_DIR/test/output/test-hxsec-output.beancount" "$EXPECTED_OUTPUT"
    echo "Please review the newly created $EXPECTED_OUTPUT file."
    exit 0 # Exit successfully after creating the file
fi

diff -u --color \
    "$EXPECTED_OUTPUT" \
    "$ROOT_DIR/test/output/test-hxsec-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] hxsec provider (beancount) output is different from expected output."
    exit 1
fi

echo "[PASS] hxsec provider (beancount) test passed!"
