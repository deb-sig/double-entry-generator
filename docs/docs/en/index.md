---
title: Home
description: Rule-based double-entry bookkeeping importer
---


# Rule-based Double-Entry Bookkeeping Importer

Intelligently convert various bill formats to Beancount or Ledger formats, making double-entry bookkeeping simple and efficient

[:material-rocket: Getting Started](getting-started.md){ .md-button .md-button--primary }  [:material-github: GitHub](https://github.com/deb-sig/double-entry-generator){ .md-button }

---

## ✨ Features

<div class="features">
  <div class="feature">
    <h3>🏦 Multi-Bank Support</h3>
    <p>Supports major banks including CCB, ICBC, CITIC, HSBC, and more</p>
  </div>
  <div class="feature">
    <h3>💰 Payment Tools</h3>
    <p>Supports Alipay, WeChat, and other mainstream payment platforms</p>
  </div>
  <div class="feature">
    <h3>📈 Securities Trading</h3>
    <p>Supports trading records from HTSEC, HXSEC, and other securities firms</p>
  </div>
  <div class="feature">
    <h3>🪙 Cryptocurrency</h3>
    <p>Supports crypto trading records from Huobi and other exchanges</p>
  </div>
  <div class="feature">
    <h3>🛒 Life Services</h3>
    <p>Supports bills from Meituan, JD.com, and other lifestyle platforms</p>
  </div>
  <div class="feature">
    <h3>⚙️ Smart Rules</h3>
    <p>Rule-based intelligent categorization with custom account mapping</p>
  </div>
</div>

## 🚀 Quick Start

### Installation

Two installation methods are provided below:

```bash
# Install using Go (Recommended)
go install github.com/deb-sig/double-entry-generator/v2@latest

# Install using Homebrew (macOS)
brew install deb-sig/deb-sig/double-entry-generator
```

### Basic Usage

```bash
# Convert Alipay bills
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# Convert WeChat bills (supports CSV and XLSX)
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx

# Convert bank statements
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

## Supported Providers

### 🏦 Banks
- [China Construction Bank (CCB)](providers/banks/ccb.md) - Supports CSV, XLS, XLSX formats
- [Industrial and Commercial Bank of China (ICBC)](providers/banks/icbc.md) - Auto-detects debit/credit cards
- [China CITIC Bank (CITIC)](providers/banks/citic.md) - Credit card statements
- [HSBC Hong Kong](providers/banks/hsbchk.md) - HSBC Hong Kong
- [Bank of Montreal (BMO)](providers/banks/bmo.md) - Bank of Montreal
- [Toronto-Dominion Bank (TD)](providers/banks/td.md) - Toronto-Dominion Bank

### 💰 Payment Tools
- [Alipay](providers/payment/alipay.md) - Supports CSV format
- [WeChat](providers/payment/wechat.md) - Supports CSV and XLSX formats

### 📈 Securities Trading
- [Haitong Securities (HTSEC)](providers/securities/htsec.md) - Trading records
- [Huaxi Securities (HXSEC)](providers/securities/hxsec.md) - Trading records

### 🪙 Cryptocurrency
- [Huobi](providers/crypto/huobi.md) - Crypto trading records

### 🛒 Life Services
- [Meituan (MT)](providers/life/mt.md) - Meituan delivery/dine-in bills
- [JD.com (JD)](providers/life/jd.md) - JD.com shopping bills

## Configuration Guide

- [Configuration Overview](configuration/README.md) - Learn about configuration file structure
- [Rules Configuration](configuration/rules.md) - Learn how to write matching rules
- [Account Mapping](configuration/accounts.md) - Set up account correspondences

## Examples

- [Basic Usage Examples](examples/basic-usage.md)
- [Advanced Rules Configuration](examples/advanced-rules.md)

## Output Formats

Two double-entry bookkeeping formats are supported:

- **Beancount** - Python ecosystem double-entry bookkeeping system
- **Ledger** - Command-line double-entry bookkeeping system

## Contributing

Contributions to code and documentation are welcome! Please check our [GitHub repository](https://github.com/deb-sig/double-entry-generator).

## License

This project is licensed under the Apache 2.0 License.
