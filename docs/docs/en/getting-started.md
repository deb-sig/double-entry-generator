---
title: Getting Started
description: Quick start with Double Entry Generator
---


# Getting Started

Double Entry Generator is a rule-based double-entry bookkeeping importer that converts various bill formats to Beancount or Ledger formats.

## Installation

### Using Go (Recommended)

```bash
# Install the latest version
go install github.com/deb-sig/double-entry-generator/v2@latest

# Verify installation
double-entry-generator --version
```

### Using Homebrew (macOS)

```bash
# Add tap
brew tap deb-sig/deb-sig

# Install
brew install double-entry-generator

# Verify installation
double-entry-generator --version
```

### Using Pre-built Releases
- Visit [releases](https://github.com/deb-sig/double-entry-generator/releases)
- Choose the appropriate version
- Place it in a suitable directory in your local ledger directory
- Remember to modify the corresponding commands

### Building from Source

```bash
# Clone the repository
git clone https://github.com/deb-sig/double-entry-generator.git
cd double-entry-generator

# Build
make build

# Install to system
make install
```

## Basic Usage

### Command Format

```bash
double-entry-generator translate [options] <input file>
```

### Common Options

- `-p, --provider`: Specify data provider (e.g., alipay, wechat, ccb, etc.)
- `-t, --target`: Specify output format (beancount or ledger)
- `-c, --config`: Specify configuration file path
- `-o, --output`: Specify output file path

### Examples

#### Converting Alipay Bills

```bash
# Convert to Beancount format
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# Convert to Ledger format
double-entry-generator translate -p alipay -t ledger alipay_records.csv
```

#### Converting WeChat Bills

```bash
# Supports both CSV and XLSX formats
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx
```

#### Converting Bank Statements

```bash
# China Construction Bank
double-entry-generator translate -p ccb -t beancount ccb_records.xls

# Industrial and Commercial Bank of China
double-entry-generator translate -p icbc -t beancount icbc_records.csv
```

## Configuration File

### Basic Configuration Structure

The configuration file is YAML. Top-level keys include default accounts, default currency, and **per-provider** sections (e.g. `alipay`, `wechat`). Rules are defined under the provider they apply to, not in a global `rules` or `conditions` block.

Example (Alipay): default accounts and currency, then `alipay.rules` with match fields and `targetAccount` / `methodAccount`:

```yaml
# config.yaml (example for Alipay)
defaultMinusAccount: Assets:Alipay:Cash
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: My Alipay Ledger

alipay:
  rules:
    - category: "餐饮美食"
      time: "07:00-11:00"
      targetAccount: Expenses:Food:Breakfast
    - category: "餐饮美食"
      time: "11:00-15:00"
      targetAccount: Expenses:Food:Lunch
    - peer: "DiDi,Amap"
      sep: ","
      targetAccount: Expenses:Transport:Taxi
    - method: "Yu'e Bao"
      methodAccount: Assets:Alipay:YuEBao
    - method: "Balance"
      methodAccount: Assets:Alipay:Cash
```

Match fields (`peer`, `category`, `type`, `item`, `method`, `time`, `minPrice`, `maxPrice`, etc.) are **top-level keys in each rule**; there is no nested `conditions` block. See [Rules Configuration](configuration/rules.md) for all supported fields per provider.

### Configuration File Locations

1. Specified via the `-c` / `--config` parameter (e.g. `-c config.yaml` in the current directory).
2. If not specified: the program looks for a file named `.double-entry-generator` (with extension such as `.yaml` or `.yml`) in your home directory, e.g. `~/.double-entry-generator.yaml`.

## Output Formats

### Beancount Format

```beancount
2024-01-15 * "Meituan Takeout" "Lunch"
  Assets:Alipay  -25.00 CNY
  Expenses:Food   25.00 CNY
  # imported
```

### Ledger Format

```ledger
2024-01-15 * Meituan Takeout Lunch
    Assets:Alipay  -25.00 CNY
    Expenses:Food   25.00 CNY
    ; imported
```

## Next Steps

- Check the [Configuration Guide](configuration/README.md) for detailed configuration
- Browse [Supported Providers](providers.md) to see all supported data sources
- View [Examples](examples/basic-usage.md) to learn advanced usage

## Common Questions

### Q: How to handle unsupported bill formats?

A: You can:
1. Check if there are similar providers you can reference
2. Submit an issue on GitHub requesting support
3. Contribute code to add a new provider

### Q: How to customize account mapping?

A: Use the **default account** keys (`defaultMinusAccount`, `defaultPlusAccount`, `defaultCashAccount`, etc.) for fallbacks, and use **rules** under each provider (e.g. `alipay.rules`) to set `targetAccount`, `methodAccount`, or `pnlAccount` per match. There is no separate `accounts` section; account assignment is done via these rule fields and defaults. See [Configuration Guide](configuration/README.md) and [Account Mapping](configuration/accounts.md).

### Q: Output file encoding issues?

A: Ensure the input file uses UTF-8 encoding, or specify the correct encoding format through configuration.
