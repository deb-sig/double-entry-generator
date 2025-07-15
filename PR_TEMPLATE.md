# Pull Request Template

## Description

Added xlsx format support to the WeChat provider to handle Excel files exported from WeChat Pay. This change allows users to directly process `.xlsx` files without converting them to CSV format first.

Fixes #172 - WeChat provider doesn't support xlsx format

## Motivation and Context

WeChat Pay now exports bill data in xlsx format instead of CSV, but the current WeChat provider only supports CSV files. Users need to manually convert xlsx files to CSV format before processing, which is inconvenient and error-prone. This change maintains backward compatibility with existing CSV files while adding support for the new xlsx format.

## Dependencies 

- `github.com/xuri/excelize/v2` - Already included in go.mod for other providers (htsec, ccb)

## Type of change

- [x] New feature (non-breaking change which adds functionality)
- [x] This change requires a documentation update

## How has this been tested?

Please describe the tests that you ran to verify your changes. Provide instructions so we can reproduce. 

Please also list any relevant details for your test configuration

- [x] Test A: Compiled successfully with `go build -o bin/double-entry-generator.exe .`
- [x] Test B: Verified file format detection logic works correctly
- [x] Test C: Confirmed backward compatibility with existing CSV files
- [x] Test D: Created test xlsx file from existing CSV example

**Test Instructions:**
1. Build the project: `go build -o bin/double-entry-generator.exe .`
2. Test with xlsx file: `./bin/double-entry-generator translate --provider wechat --config ./example/wechat/config.yaml --output ./test/output/test-wechat-xlsx-output.beancount ./example/wechat/example-wechat-records.xlsx`
3. Verify output matches expected format

## Is this change properly documented?

- [x] Code changes are self-documenting with clear function names and comments
- [x] Added Chinese error messages for better user experience
- [x] Maintained existing logging patterns for consistency

**Documentation Updates Needed:**
- Update README.md to mention xlsx support for WeChat provider
- Add example usage with xlsx files in documentation

## Technical Details

- Added file extension detection in `Translate()` method
- Split parsing logic into `translateCSV()` and `translateExcel()` functions
- Used `excelize` library to read xlsx files (Sheet1)
- Maintained same data processing logic for both formats
- Preserved existing 17-line header skip logic for xlsx files 