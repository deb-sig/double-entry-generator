---
title: Home
layout: home
nav_order: 1
description: "Rule-based double-entry bookkeeping importer"
permalink: /en/
lang: en
---

# Rule-based Double-Entry Bookkeeping Importer

Intelligently convert various bill formats to Beancount or Ledger formats, making double-entry bookkeeping simple and efficient

[Getting Started]({{ site.baseurl }}{% link en/getting-started.md %}){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 } [中文]({{ site.baseurl }}/){: .btn .btn-secondary .fs-5 .mb-4 .mb-md-0 .mr-2 } [GitHub](https://github.com/deb-sig/double-entry-generator){: .btn .fs-5 .mb-4 .mb-md-0 }

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
- [China Construction Bank (CCB)]({{ site.baseurl }}{% link en/providers/banks/ccb.md %}) - Supports CSV, XLS, XLSX formats
- [Industrial and Commercial Bank of China (ICBC)]({{ site.baseurl }}{% link en/providers/banks/icbc.md %}) - Auto-detects debit/credit cards
- [China CITIC Bank (CITIC)]({{ site.baseurl }}{% link en/providers/banks/citic.md %}) - Credit card statements
- [HSBC Hong Kong]({{ site.baseurl }}{% link en/providers/banks/hsbchk.md %}) - HSBC Hong Kong
- [Bank of Montreal (BMO)]({{ site.baseurl }}{% link en/providers/banks/bmo.md %}) - Bank of Montreal
- [Toronto-Dominion Bank (TD)]({{ site.baseurl }}{% link en/providers/banks/td.md %}) - Toronto-Dominion Bank

### 💰 Payment Tools
- [Alipay]({{ site.baseurl }}{% link en/providers/payment/alipay.md %}) - Supports CSV format
- [WeChat]({{ site.baseurl }}{% link en/providers/payment/wechat.md %}) - Supports CSV and XLSX formats

### 📈 Securities Trading
- [Haitong Securities (HTSEC)]({{ site.baseurl }}{% link en/providers/securities/htsec.md %}) - Trading records
- [Huaxi Securities (HXSEC)]({{ site.baseurl }}{% link en/providers/securities/hxsec.md %}) - Trading records

### 🪙 Cryptocurrency
- [Huobi]({{ site.baseurl }}{% link en/providers/crypto/huobi.md %}) - Crypto trading records

### 🛒 Life Services
- [Meituan (MT)]({{ site.baseurl }}{% link en/providers/life/mt.md %}) - Meituan delivery/dine-in bills
- [JD.com (JD)]({{ site.baseurl }}{% link en/providers/life/jd.md %}) - JD.com shopping bills

## Configuration Guide

- [Configuration Overview]({{ site.baseurl }}{% link en/configuration/README.md %}) - Learn about configuration file structure
- [Rules Configuration]({{ site.baseurl }}{% link en/configuration/rules.md %}) - Learn how to write matching rules
- [Account Mapping]({{ site.baseurl }}{% link en/configuration/accounts.md %}) - Set up account correspondences

## Examples

- [Basic Usage Examples]({{ site.baseurl }}{% link en/examples/basic-usage.md %})
- [Advanced Rules Configuration]({{ site.baseurl }}{% link en/examples/advanced-rules.md %})

## Output Formats

Two double-entry bookkeeping formats are supported:

- **Beancount** - Python ecosystem double-entry bookkeeping system
- **Ledger** - Command-line double-entry bookkeeping system

## Contributing

Contributions to code and documentation are welcome! Please check our [GitHub repository](https://github.com/deb-sig/double-entry-generator).

## License

This project is licensed under the Apache 2.0 License.
