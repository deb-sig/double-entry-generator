#!/usr/bin/env bash
# End-to-End test functions for provider testing

set -euo pipefail

DIFF_SUPPORTS_COLOR=0

init_test_env() {
    TEST_DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")
    export ROOT_DIR="$TEST_DIR/.."
    
    mkdir -p "$ROOT_DIR/test/output"

    # BSD diff (macOS) doesn't support `--color`; GNU diff does. Auto-detect to keep
    # `make test-providers` usable across environments.
    if diff -u --color /dev/null /dev/null >/dev/null 2>&1; then
        DIFF_SUPPORTS_COLOR=1
    fi
}

build_binary() {
    # Always build to avoid using a stale `bin/double-entry-generator`.
    if ! make -C "$ROOT_DIR" build >/dev/null 2>&1; then
        echo "Error: Failed to build double-entry-generator" >&2
        return 1
    fi
}

diff_expected_output() {
    local expected_file=$1
    local output_file=$2

    if [ "$DIFF_SUPPORTS_COLOR" -eq 1 ]; then
        diff -u --color "$expected_file" "$output_file"
        return $?
    fi

    diff -u "$expected_file" "$output_file"
}

INPUT_FILES=()
collect_input_files() {
    local example_dir=$1
    local case_name=$2

    INPUT_FILES=()

    if [ "$case_name" = "." ]; then
        # Root-level case: only consider input files directly under the provider
        # directory (e.g. `example/wechat/example-wechat-records.csv`). This keeps
        # provider root tests independent from `credit/`, `debit/` sub-cases.
        for ext in csv xls xlsx; do
            for file in "$example_dir"/*."$ext"; do
                [ -f "$file" ] && INPUT_FILES+=("$file")
            done
        done
        return 0
    fi

    for ext in csv xls xlsx; do
        while IFS= read -r file; do
            [ -f "$file" ] && INPUT_FILES+=("$file")
        done < <(find "$example_dir" -type f -name "*.$ext" 2>/dev/null || true)
    done
}

run_translate_and_diff() {
    local provider=$1
    local target=$2
    local config_file=$3
    local expected_file=$4
    local output_file=$5
    local input_file=$6

    local translate_args=(
        "--provider" "$provider"
        "--config" "$config_file"
        "--output" "$output_file"
    )

    case "$provider" in
        wechat)
            # WeChat CSV/XLSX exports may contain tx types not covered by the parser.
            # Existing examples include such lines; this flag keeps E2E stable.
            translate_args+=("--ignore-invalid-tx-types")
            ;;
    esac

    [ "$target" != "beancount" ] && translate_args+=("--target" "$target")
    translate_args+=("$input_file")

    "$ROOT_DIR/bin/double-entry-generator" translate "${translate_args[@]}" || return 1
    diff_expected_output "$expected_file" "$output_file" || return 1
}

run_test_case() {
    local provider=$1
    local target=$2
    local case_name=$3
    
    # Build directory and file paths
    # "." indicates root directory (no subdirectory), other values are subdirectory names
    local example_dir="$ROOT_DIR/example/$provider"
    local file_suffix="$provider"
    
    if [ "$case_name" != "." ]; then
        example_dir="$example_dir/$case_name"
        file_suffix="$provider-$case_name"
    fi
    
    # Unified file path construction
    local config_file="$example_dir/config.yaml"
    local expected_file="$example_dir/example-$file_suffix-output.$target"
    
    # Validate file existence
    [ ! -f "$config_file" ] && { echo "Error: Config not found: $config_file" >&2; return 1; }
    [ ! -f "$expected_file" ] && { echo "Error: Expected file not found: $expected_file" >&2; return 1; }

    collect_input_files "$example_dir" "$case_name"
    [ ${#INPUT_FILES[@]} -eq 0 ] && { echo "Error: No input files found in $example_dir" >&2; return 1; }

    if [ "$provider" = "wechat" ] && [ "$case_name" = "." ]; then
        # Special-case WeChat root test:
        # - Keep parity with the old scripts by validating BOTH CSV and XLSX inputs.
        # - Keep output filenames stable:
        #     `test-wechat-output.*` for CSV
        #     `test-wechat-xlsx-output.*` for XLSX
        # - Expected output stays the same for both inputs (`example-wechat-output.*`).
        local ran_any=0
        local input_file=""
        for input_file in "${INPUT_FILES[@]}"; do
            local output_suffix=""
            case "$input_file" in
                *.csv)
                    output_suffix=""
                    ;;
                *.xlsx)
                    output_suffix="-xlsx"
                    ;;
                *)
                    # Ignore other inputs (e.g. .xls) for now to avoid growing output
                    # naming conventions without a concrete fixture.
                    continue
                    ;;
            esac

            ran_any=1
            local output_file="$ROOT_DIR/test/output/test-$file_suffix${output_suffix}-output.$target"
            run_translate_and_diff "$provider" "$target" "$config_file" "$expected_file" "$output_file" "$input_file" || return 1
        done

        [ $ran_any -eq 1 ] || { echo "Error: No supported wechat input files found in $example_dir" >&2; return 1; }
        return 0
    fi

    # Default behavior: pick the first input file by extension priority
    # (`csv > xls > xlsx`) and run once.
    local input_file="${INPUT_FILES[0]}"
    local output_file="$ROOT_DIR/test/output/test-$file_suffix-output.$target"
    run_translate_and_diff "$provider" "$target" "$config_file" "$expected_file" "$output_file" "$input_file" || return 1
    
    return 0
}

test_provider() {
    local provider=$1
    local target=$2
    local cases=$3
    local skip_init=${4:-false}
    
    if [ "$skip_init" != "true" ]; then
        init_test_env
    fi
    
    local failed=0
    IFS=',' read -ra case_array <<< "$cases"
    
    for case_name in "${case_array[@]}"; do
        local mode_desc=""
        [ "$case_name" != "." ] && mode_desc=" ($case_name mode)"
        
        if ! run_test_case "$provider" "$target" "$case_name"; then
            echo "[FAIL] $provider provider$mode_desc output is different from expected output."
            failed=1
        fi
    done
    
    [ $failed -eq 1 ] && return 1
    
    echo "[PASS] All $provider provider for $target target tests!"
    return 0
}

# Find test config for specified provider and target from TEST_CONFIGS
find_test_config() {
    local search_provider=$1
    local search_target=$2
    
    for config in $TEST_CONFIGS; do
        IFS=':' read -r provider targets cases <<< "$config"
        if [ "$provider" = "$search_provider" ]; then
            if echo "$targets" | grep -qw "$search_target"; then
                echo "$cases"
                return 0
            fi
        fi
    done
    
    return 1
}

# Run test for a single provider
run_single_test() {
    local provider=$1
    local target=$2
    
    if [ -z "$provider" ] || [ -z "$target" ]; then
        echo "Usage: run_single_test <provider> <target>" >&2
        return 1
    fi
    
    local cases=$(find_test_config "$provider" "$target")
    if [ -z "$cases" ]; then
        echo "Error: No test config found for $provider with target $target" >&2
        echo "Check TEST_CONFIGS for available configurations" >&2
        return 1
    fi
    
    init_test_env
    build_binary || return 1
    
    echo ">> Testing $provider ($target)"
    test_provider "$provider" "$target" "$cases" true
}

# Run all tests
run_all_tests() {
    init_test_env
    
    echo "Building double-entry-generator..."
    if ! build_binary; then
        echo "Error: Build failed, cannot proceed with tests" >&2
        return 1
    fi
    echo "Build successful!"
    echo ""
    
    local failed=0
    local total=0
    local passed=0
    
    for config in $TEST_CONFIGS; do
        IFS=':' read -r provider targets cases <<< "$config"
        IFS=',' read -ra target_array <<< "$targets"
        
        for target in "${target_array[@]}"; do
            total=$((total + 1))
            echo ""
            echo ">> Testing $provider ($target)"
            
            if test_provider "$provider" "$target" "$cases" true; then
                passed=$((passed + 1))
            else
                failed=$((failed + 1))
            fi
        done
    done
    
    echo ""
    echo "========================================"
    echo "Test Summary: $passed/$total passed"
    echo "========================================"
    
    [ $failed -gt 0 ] && return 1 || return 0
}
