#!/usr/bin/env bash
#
# E2E test for citic provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate citic bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider citic \
    --config "$ROOT_DIR/example/citic/credit/config.yaml" \
    --output "$ROOT_DIR/test/output/test-citic-credit-output.beancount" \
    "$ROOT_DIR/example/citic/credit/example-citic-records.xls"

diff -u --color \
    "$ROOT_DIR/example/citic/credit/example-citic-output.beancount" \
    "$ROOT_DIR/test/output/test-citic-credit-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] citic provider output is different from expected output."
    exit 1
fi

echo "[PASS] All citic provider tests!"
