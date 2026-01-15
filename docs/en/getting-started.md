---
title: Getting Started
layout: default
nav_order: 2
description: "Quick start with Double Entry Generator"
permalink: /en/getting-started/
lang: en
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

```yaml
# config.yaml
default:
  # Default account settings
  default_minus_account: "Assets:Bank:Checking"
  default_plus_account: "Expenses:Unknown"
  
  # Default currency
  default_currency: "CNY"
  
  # Default tags
  default_tags: ["imported"]

# Rule configuration
rules:
  - name: "Food & Dining"
    conditions:
      - field: "description"
        contains: ["Meituan", "Ele.me", "Restaurant"]
    target_account: "Expenses:Food"
    tags: ["food", "dining"]

# Account mapping
accounts:
  "Alipay": "Assets:Alipay"
  "WeChat": "Assets:WeChat"
```

### Configuration File Locations

1. `config.yaml` in the current directory
2. `~/.double-entry-generator/config.yaml` in the user's home directory
3. Specified via the `-c` parameter

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

- Check the [Configuration Guide]({{ '/en/configuration/' | relative_url }}) for detailed configuration
- Browse [Supported Providers]({{ '/en/providers/' | relative_url }}) to see all supported data sources
- View [Examples]({{ '/en/examples/' | relative_url }}) to learn advanced usage

## Common Questions

### Q: How to handle unsupported bill formats?

A: You can:
1. Check if there are similar providers you can reference
2. Submit an issue on GitHub requesting support
3. Contribute code to add a new provider

### Q: How to customize account mapping?

A: Add mapping relationships in the `accounts` section of the configuration file, supporting regular expression matching.

### Q: Output file encoding issues?

A: Ensure the input file uses UTF-8 encoding, or specify the correct encoding format through configuration.
