#!/usr/bin/env bash
#
# E2E test for SPDB debit provider.

set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate SPDB debit bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider spdb_debit \
    --config "$ROOT_DIR/example/spdb_debit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-spdb_debit-output.beancount" \
    "$ROOT_DIR/example/spdb_debit/example-spdb_debit-records.xls"

diff -u --color \
    "$ROOT_DIR/example/spdb_debit/example-spdb_debit-output.beancount" \
    "$ROOT_DIR/test/output/test-spdb_debit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] SPDB debit provider output is different from expected output."
    exit 1
fi

echo "[PASS] All SPDB debit provider beancount tests!"
