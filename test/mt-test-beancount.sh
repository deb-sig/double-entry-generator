#!/usr/bin/env bash
#
# E2E test for mt provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate mt bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider mt \
    --config "$ROOT_DIR/example/mt/config.yaml" \
    --output "$ROOT_DIR/test/output/test-mt-output.bean" \
    "$ROOT_DIR/example/mt/example-mt-records.csv"

diff -u --color \
    "$ROOT_DIR/example/mt/example-mt-output.bean" \
    "$ROOT_DIR/test/output/test-mt-output.bean"

if [ $? -ne 0 ]; then
    echo "[FAIL] mt provider output is different from expected output."
    exit 1
fi

echo "[PASS] All mt provider tests!"
