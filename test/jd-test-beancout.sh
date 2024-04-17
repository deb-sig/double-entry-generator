#!/usr/bin/env bash
#
# E2E test for jd provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate jd bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider jd \
    --config "$ROOT_DIR/example/jd/config.yaml" \
    --output "$ROOT_DIR/test/output/test-jd-output.beancount" \
    "$ROOT_DIR/example/jd/example-jd-records.csv"

diff -u --color \
    "$ROOT_DIR/example/jd/example-jd-output.beancount" \
    "$ROOT_DIR/test/output/test-jd-output.beancount"

if [ $? -ne 0 ]; then
    echo "[FAIL] jd provider output is different from expected output."
    exit 1
fi

echo "[PASS] All jd provider tests!"
