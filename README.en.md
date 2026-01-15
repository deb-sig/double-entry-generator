# double-entry-generator

[![GitHub](https://img.shields.io/github/license/deb-sig/double-entry-generator)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deb-sig/double-entry-generator)](go.mod)
[![Documentation](https://img.shields.io/badge/docs-online-brightgreen)](https://deb-sig.github.io/double-entry-generator/)

A rule-based double-entry bookkeeping importer that intelligently converts various bill formats to [Beancount](https://beancount.github.io/) or [Ledger](https://www.ledger-cli.org/) formats.

> 📖 **Full Documentation**: Visit the [online documentation site](https://deb-sig.github.io/double-entry-generator/) for detailed usage guides and configuration instructions.

## ✨ Features

- 🏦 **Multi-Bank Support** - Supports major banks including CCB, ICBC, CITIC, HSBC, etc.
- 💰 **Payment Tools** - Supports Alipay, WeChat, and other mainstream payment platforms
- 📈 **Securities Trading** - Supports trading records from HTSEC, HXSEC, and other securities firms
- 🪙 **Cryptocurrency** - Supports crypto trading records from Huobi and other exchanges
- 🛒 **Life Services** - Supports bills from Meituan, JD.com, and other lifestyle platforms
- ⚙️ **Smart Rules** - Rule-based intelligent categorization with custom account mapping
- 🔧 **Extensible Architecture** - Easy to add new bill formats and accounting language support

## 🚀 Quick Start

### Installation

#### Using Go (Recommended)

```bash
go install github.com/deb-sig/double-entry-generator/v2@latest
```

#### Using Homebrew (macOS)

```bash
brew install deb-sig/tap/double-entry-generator
```

#### Binary Installation

Download the binary file for your architecture from the [GitHub Release](https://github.com/deb-sig/double-entry-generator/releases) page.

> [!TIP]
> After installing via Go, you can check the version using `go version -m $(which double-entry-generator)`.

### Basic Usage

```bash
# Convert Alipay bills to Beancount format
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# Convert WeChat bills to Ledger format
double-entry-generator translate -p wechat -t ledger wechat_records.xlsx

# Convert bank statements
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

For more usage instructions, please refer to the [Getting Started Guide](https://deb-sig.github.io/double-entry-generator/getting-started/).

## 📋 Supported Providers

### 🏦 Banks

- [China Construction Bank (CCB)](https://deb-sig.github.io/double-entry-generator/providers/banks/ccb.html) - Supports CSV, XLS, XLSX formats
- [Industrial and Commercial Bank of China (ICBC)](https://deb-sig.github.io/double-entry-generator/providers/banks/icbc.html) - Auto-detects debit/credit cards
- [China CITIC Bank (CITIC)](https://deb-sig.github.io/double-entry-generator/providers/banks/citic.html) - Credit card statements
- [HSBC Hong Kong](https://deb-sig.github.io/double-entry-generator/providers/banks/hsbchk.html) - HSBC Hong Kong
- [Bank of Montreal (BMO)](https://deb-sig.github.io/double-entry-generator/providers/banks/bmo.html)
- [Toronto-Dominion Bank (TD)](https://deb-sig.github.io/double-entry-generator/providers/banks/td.html)
- [China Merchants Bank (CMB)](https://deb-sig.github.io/double-entry-generator/providers/banks/cmb.html) - Supports savings and credit cards
- [Bank of Communications Debit Card (BOCOM Debit)](https://deb-sig.github.io/double-entry-generator/providers/banks/bocom-debit.html)
- [Agricultural Bank of China Debit Card (ABC Debit)](https://deb-sig.github.io/double-entry-generator/providers/banks/abc_debit.html)

### 💰 Payment Tools

- [Alipay](https://deb-sig.github.io/double-entry-generator/providers/payment/alipay.html) - Supports CSV format
- [WeChat](https://deb-sig.github.io/double-entry-generator/providers/payment/wechat.html) - Supports CSV and XLSX formats

### 📈 Securities Trading

- [Haitong Securities (HTSEC)](https://deb-sig.github.io/double-entry-generator/providers/securities/htsec.html) - Trading records
- [Huaxi Securities (HXSEC)](https://deb-sig.github.io/double-entry-generator/providers/securities/hxsec.html) - Trading records

### 🪙 Cryptocurrency

- [Huobi](https://deb-sig.github.io/double-entry-generator/providers/crypto/huobi.html) - Crypto trading records

### 🛒 Life Services

- [Meituan (MT)](https://deb-sig.github.io/double-entry-generator/providers/life/mt.html) - Meituan delivery/dine-in bills
- [JD.com (JD)](https://deb-sig.github.io/double-entry-generator/providers/life/jd.html) - JD.com shopping bills

For the complete list, please check the [Providers Documentation](https://deb-sig.github.io/double-entry-generator/providers.html).

## ⚙️ Configuration Guide

Double Entry Generator uses YAML format configuration files to define conversion rules and account mappings.

### Basic Configuration Structure

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: My Ledger Configuration

# Provider-specific configuration
alipay:
  rules:
    - category: 餐饮美食
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
    - peer: 滴滴
      targetAccount: Expenses:Transport:Taxi
```

### Configuration Documentation

- [Configuration Overview](https://deb-sig.github.io/double-entry-generator/configuration/) - Learn about configuration file structure
- [Rules Configuration](https://deb-sig.github.io/double-entry-generator/configuration/rules.html) - Learn how to write matching rules
- [Account Mapping](https://deb-sig.github.io/double-entry-generator/configuration/accounts.html) - Set up account correspondences

## 📖 Examples

The project provides rich example configurations and bill files in the `example/` directory.

### Alipay Example

```bash
double-entry-generator translate \
  --config ./example/alipay/config.yaml \
  --output ./example/alipay/example-alipay-output.beancount \
  ./example/alipay/example-alipay-records.csv
```

### WeChat Example

```bash
double-entry-generator translate \
  --config ./example/wechat/config.yaml \
  --provider wechat \
  --output ./example/wechat/example-wechat-output.beancount \
  ./example/wechat/example-wechat-records.csv
```

For more examples, please check the [Examples Documentation](https://deb-sig.github.io/double-entry-generator/examples/).

## 🏗️ Architecture

```
┌───────────┐  ┌──────────┐  ┌────┐  ┌──────────┐  ┌──────────┐
│ translate │->│ provider │->│ IR │->│ compiler │->│ analyser │
└───────────┘  └──────────┘  └────┘  └──────────┘  └──────────┘
                  alipay               beancount      alipay
                  wechat               ledger         wechat
                  huobi                               huobi
                  htsec                               htsec
                  icbc                                icbc
                  ccb                                 ccb
                  td                                  td
                  bmo                                 bmo
                  hsbchk                              hsbchk
```

The architecture supports extension:
- Add new bill formats: Implement new [provider](pkg/provider)
- Add new accounting languages: Implement new [compiler](pkg/compiler)

## 📚 Documentation

Complete documentation is available at:

- 🌐 [Online Documentation Site](https://deb-sig.github.io/double-entry-generator/) - Complete online documentation
- 📖 [Getting Started](https://deb-sig.github.io/double-entry-generator/getting-started/) - Installation and basic usage
- 📋 [Providers List](https://deb-sig.github.io/double-entry-generator/providers.html) - All supported data sources
- ⚙️ [Configuration Guide](https://deb-sig.github.io/double-entry-generator/configuration/) - Detailed configuration instructions
- 💡 [Examples](https://deb-sig.github.io/double-entry-generator/examples/) - Usage examples and best practices

## 🐛 Common Issues

### How to handle unsupported transaction types?

If you encounter `"Failed to get the tx type"` error:

1. Report the issue on [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues)
2. If the transaction type is an expense and version >= `v2.10.0`, you can use the `--ignore-invalid-tx-types` parameter to ignore this error

### How to obtain bill files?

For bill download methods for each provider, please refer to:
- [Alipay Bill Download](https://blog.triplez.cn/posts/bills-export-methods/#%e6%94%af%e4%bb%98%e5%ae%9d)
- [WeChat Bill Download](https://blog.triplez.cn/posts/bills-export-methods/#%e5%be%ae%e4%bf%a1%e6%94%af%e4%bb%98)
- [ICBC Bill Download](https://blog.triplez.cn/posts/bills-export-methods/#%e4%b8%ad%e5%9b%bd%e5%b7%a5%e5%95%86%e9%93%b6%e8%a1%8c)

For more issues, please check [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues).

## 🤝 Contributing

Contributions to code and documentation are welcome! Please check the [Contributing Guide](https://deb-sig.github.io/double-entry-generator/contributing/).

### How to Contribute

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the [Apache 2.0](LICENSE) License.

## 🙏 Acknowledgments

- [dilfish/atb](https://github.com/dilfish/atb) - Early version of Alipay bill to Beancount converter

## 📞 Contact

- GitHub: [deb-sig/double-entry-generator](https://github.com/deb-sig/double-entry-generator)
- Issues: [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues)
