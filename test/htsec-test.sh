#!/usr/bin/env bash
#
# E2E test for htsec provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate htsec bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider htsec \
    --config "$ROOT_DIR/example/htsec/config.yaml" \
    --output "$ROOT_DIR/test/output/test-htsec-output.beancount" \
    "$ROOT_DIR/example/htsec/example-htsec-records.xlsx"

diff -u --color \
    "$ROOT_DIR/example/htsec/example-htsec-output.beancount" \
    "$ROOT_DIR/test/output/test-htsec-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] htsec provider output is different from expected output."
    exit 1
fi

echo "[PASS] All htsec provider tests!"
