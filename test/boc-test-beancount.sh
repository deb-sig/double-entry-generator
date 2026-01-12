
#!/usr/bin/env bash
#
# E2E test for boc provider.

# set -x # debug
set -eo errexit

TEST_DIR=`dirname "$(realpath $0)"`
ROOT_DIR="$TEST_DIR/.."

make -f "$ROOT_DIR/Makefile" build
mkdir -p "$ROOT_DIR/test/output"

# generate boc bills output in beancount format
"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider boc \
    --config "$ROOT_DIR/example/boc/config.yaml" \
    --output "$ROOT_DIR/test/output/test-boc-debit-output.bean" \
    "$ROOT_DIR/example/boc/example-boc-debit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/boc/example-boc-debit-output.bean" \
    "$ROOT_DIR/test/output/test-boc-debit-output.bean"

if [ $? -ne 0 ]; then
    echo "[FAIL] boc provider output for debit card is different from expected output."
    exit 1
fi

"$ROOT_DIR/bin/double-entry-generator" translate \
    --provider boc \
    --config "$ROOT_DIR/example/boc/config.yaml" \
    --output "$ROOT_DIR/test/output/test-boc-credit-output.bean" \
    "$ROOT_DIR/example/boc/example-boc-credit-records.csv"

diff -u --color \
    "$ROOT_DIR/example/boc/example-boc-credit-output.bean" \
    "$ROOT_DIR/test/output/test-boc-credit-output.bean"

if [ $? -ne 0 ]; then
    echo "[FAIL] boc provider output for credit card is different from expected output."
    exit 1
fi
echo "[PASS] All mt provider tests!"
