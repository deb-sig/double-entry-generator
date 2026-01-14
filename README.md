# double-entry-generator

[![GitHub](https://img.shields.io/github/license/deb-sig/double-entry-generator)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deb-sig/double-entry-generator)](go.mod)
[![Documentation](https://img.shields.io/badge/docs-online-brightgreen)](https://deb-sig.github.io/double-entry-generator/)

基于规则的复式记账导入器，支持将各种账单格式智能转换为 [Beancount](https://beancount.github.io/) 或 [Ledger](https://www.ledger-cli.org/) 格式。

> 📖 **完整文档**: 访问 [在线文档站点](https://deb-sig.github.io/double-entry-generator/) 获取详细的使用指南和配置说明。

## ✨ 特性

- 🏦 **多银行支持** - 支持建设银行、工商银行、中信银行、汇丰银行等主流银行账单
- 💰 **支付工具** - 支持支付宝、微信等主流支付平台的账单导入
- 📈 **证券交易** - 支持海通证券、华西证券等券商的交易记录
- 🪙 **加密货币** - 支持火币等交易所的币币交易记录
- 🛒 **生活服务** - 支持美团、京东等生活服务平台的账单
- ⚙️ **智能规则** - 基于规则的智能分类，支持自定义账户映射
- 🔧 **可扩展架构** - 易于添加新的账单格式和记账语言支持

## 🚀 快速开始

### 安装

#### 使用 Go 安装（推荐）

```bash
go install github.com/deb-sig/double-entry-generator/v2@latest
```

#### 使用 Homebrew 安装（macOS）

```bash
brew install deb-sig/tap/double-entry-generator
```

#### 二进制安装

在 [GitHub Release](https://github.com/deb-sig/double-entry-generator/releases) 页面下载对应架构的二进制文件。

> [!TIP]
> 通过 Go 安装后，可使用 `go version -m $(which double-entry-generator)` 查看版本。

### 基本用法

```bash
# 转换支付宝账单为 Beancount 格式
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# 转换微信账单为 Ledger 格式
double-entry-generator translate -p wechat -t ledger wechat_records.xlsx

# 转换建设银行账单
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

更多使用说明请参考 [快速开始文档](https://deb-sig.github.io/double-entry-generator/getting-started/)。

## 📋 支持的 Providers

### 🏦 银行

- [建设银行 (CCB)](https://deb-sig.github.io/double-entry-generator/providers/banks/ccb.html) - 支持 CSV、XLS、XLSX 格式
- [工商银行 (ICBC)](https://deb-sig.github.io/double-entry-generator/providers/banks/icbc.html) - 自动识别借记卡/信用卡
- [中信银行 (CITIC)](https://deb-sig.github.io/double-entry-generator/providers/banks/citic.html) - 信用卡账单
- [汇丰银行香港 (HSBC HK)](https://deb-sig.github.io/double-entry-generator/providers/banks/hsbchk.html) - 香港汇丰银行
- [加拿大银行 (BMO)](https://deb-sig.github.io/double-entry-generator/providers/banks/bmo.html) - Bank of Montreal
- [道明银行 (TD)](https://deb-sig.github.io/double-entry-generator/providers/banks/td.html) - Toronto-Dominion Bank
- [招商银行 (CMB)](https://deb-sig.github.io/double-entry-generator/providers/banks/cmb.html) - 支持储蓄卡和信用卡
- [交通银行储蓄卡 (BOCOM Debit)](https://deb-sig.github.io/double-entry-generator/providers/banks/bocom-debit.html)
- [农业银行储蓄卡 (ABC Debit)](https://deb-sig.github.io/double-entry-generator/providers/banks/abc_debit.html) - 借记卡账单

### 💰 支付工具

- [支付宝 (Alipay)](https://deb-sig.github.io/double-entry-generator/providers/payment/alipay.html) - 支持 CSV 格式
- [微信 (WeChat)](https://deb-sig.github.io/double-entry-generator/providers/payment/wechat.html) - 支持 CSV 和 XLSX 格式

### 📈 证券交易

- [海通证券 (HTSEC)](https://deb-sig.github.io/double-entry-generator/providers/securities/htsec.html) - 证券交易记录
- [华西证券 (HXSEC)](https://deb-sig.github.io/double-entry-generator/providers/securities/hxsec.html) - 证券交易记录

### 🪙 加密货币

- [火币 (Huobi)](https://deb-sig.github.io/double-entry-generator/providers/crypto/huobi.html) - 币币交易记录
- [OKLink](https://deb-sig.github.io/double-entry-generator/providers/crypto/oklink.html) - 多链代币转账记录（ERC20、TRC20 等）

### 🛒 生活服务

- [美团 (MT)](https://deb-sig.github.io/double-entry-generator/providers/life/mt.html) - 美团外卖/到店账单
- [京东 (JD)](https://deb-sig.github.io/double-entry-generator/providers/life/jd.html) - 京东购物账单

完整列表请查看 [Providers 文档](https://deb-sig.github.io/double-entry-generator/providers.html)。

## ⚙️ 配置指南

Double Entry Generator 使用 YAML 格式的配置文件来定义转换规则和账户映射。

### 基本配置结构

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 我的账本配置

# Provider 特定配置
alipay:
  rules:
    - category: 餐饮美食
      time: 11:00-14:00
      targetAccount: Expenses:Food:Lunch
    - peer: 滴滴
      targetAccount: Expenses:Transport:Taxi
```

### 配置文档

- [配置总览](https://deb-sig.github.io/double-entry-generator/configuration/) - 了解配置文件结构
- [规则配置](https://deb-sig.github.io/double-entry-generator/configuration/rules.html) - 学习如何编写匹配规则
- [账户映射](https://deb-sig.github.io/double-entry-generator/configuration/accounts.html) - 设置账户对应关系

## 📖 示例

项目提供了丰富的示例配置和账单文件，位于 `example/` 目录下。

### 支付宝示例

```bash
double-entry-generator translate \
  --config ./example/alipay/config.yaml \
  --output ./example/alipay/example-alipay-output.beancount \
  ./example/alipay/example-alipay-records.csv
```

### 微信示例

```bash
double-entry-generator translate \
  --config ./example/wechat/config.yaml \
  --provider wechat \
  --output ./example/wechat/example-wechat-output.beancount \
  ./example/wechat/example-wechat-records.csv
```

更多示例请查看 [示例文档](https://deb-sig.github.io/double-entry-generator/examples/)。

## 🏗️ 架构

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

架构支持扩展：
- 添加新的账单格式：实现新的 [provider](pkg/provider)
- 添加新的记账语言：实现新的 [compiler](pkg/compiler)

## 📚 文档

完整的文档请访问：

- 🌐 [在线文档站点](https://deb-sig.github.io/double-entry-generator/) - 完整的在线文档
- 📖 [快速开始](https://deb-sig.github.io/double-entry-generator/getting-started/) - 安装和基本使用
- 📋 [Providers 列表](https://deb-sig.github.io/double-entry-generator/providers.html) - 所有支持的数据源
- ⚙️ [配置指南](https://deb-sig.github.io/double-entry-generator/configuration/) - 详细的配置说明
- 💡 [示例](https://deb-sig.github.io/double-entry-generator/examples/) - 使用示例和最佳实践

## 🐛 常见问题

### 如何处理不支持的交易类型？

如果遇到 `"Failed to get the tx type"` 错误：

1. 在 [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues) 上报问题
2. 若该交易类型为支出，且版本 >= `v2.10.0`，可使用 `--ignore-invalid-tx-types` 参数忽略该错误

### 如何获取账单文件？

各 Provider 的账单下载方式请参考：
- [支付宝账单下载](https://blog.triplez.cn/posts/bills-export-methods/#%e6%94%af%e4%bb%98%e5%ae%9d)
- [微信账单下载](https://blog.triplez.cn/posts/bills-export-methods/#%e5%be%ae%e4%bf%a1%e6%94%af%e4%bb%98)
- [工商银行账单下载](https://blog.triplez.cn/posts/bills-export-methods/#%e4%b8%ad%e5%9b%bd%e5%b7%a5%e5%95%86%e9%93%b6%e8%a1%8c)

更多问题请查看 [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues)。

## 🤝 贡献

欢迎贡献代码和文档！请查看 [贡献指南](https://deb-sig.github.io/double-entry-generator/contributing/)。

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 [Apache 2.0](LICENSE) 许可证。

## 🙏 致谢

- [dilfish/atb](https://github.com/dilfish/atb) - 支付宝账单转 Beancount 的早期版本

## 📞 联系方式

- GitHub: [deb-sig/double-entry-generator](https://github.com/deb-sig/double-entry-generator)
- Issues: [GitHub Issues](https://github.com/deb-sig/double-entry-generator/issues)
