---
title: 首页
description: 基于规则的复式记账导入器
---


# 基于规则的复式记账导入器

将各种账单格式智能转换为 Beancount 或 Ledger 格式，让复式记账变得简单高效

[快速开始](getting-started.md){ .md-button .md-button--primary } [GitHub](https://github.com/deb-sig/double-entry-generator){ .md-button }

---

## ✨ 特性

<div class="features">
  <div class="feature">
    <h3>🏦 多银行支持</h3>
    <p>支持建设银行、工商银行、中信银行、汇丰银行等主流银行账单</p>
  </div>
  <div class="feature">
    <h3>💰 支付工具</h3>
    <p>支持支付宝、微信等主流支付平台的账单导入</p>
  </div>
  <div class="feature">
    <h3>📈 证券交易</h3>
    <p>支持海通证券、华西证券等券商的交易记录</p>
  </div>
  <div class="feature">
    <h3>🪙 加密货币</h3>
    <p>支持火币等交易所的币币交易记录</p>
  </div>
  <div class="feature">
    <h3>🛒 生活服务</h3>
    <p>支持美团、京东等生活服务平台的账单</p>
  </div>
  <div class="feature">
    <h3>⚙️ 智能规则</h3>
    <p>基于规则的智能分类，支持自定义账户映射</p>
  </div>
</div>

## 🚀 快速开始

### 安装

下面提供两种安装方式：

```bash
# 使用 Go 安装（推荐）
go install github.com/deb-sig/double-entry-generator/v2@latest

# 使用 Homebrew (macOS)
brew install deb-sig/deb-sig/double-entry-generator
```

### 基本用法

```bash
# 转换支付宝账单
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# 转换微信账单（支持CSV和XLSX）
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx

# 转换建设银行账单
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

## 支持的 Providers

### 🏦 银行
- [建设银行 (CCB)](providers/banks/ccb.md) - 支持 CSV、XLS、XLSX 格式
- [工商银行 (ICBC)](providers/banks/icbc.md) - 自动识别借记卡/信用卡
- [中信银行 (CITIC)](providers/banks/citic.md) - 信用卡账单
- [汇丰银行香港 (HSBC HK)](providers/banks/hsbchk.md) - 香港汇丰银行
- [加拿大银行 (BMO)](providers/banks/bmo.md) - Bank of Montreal
- [道明银行 (TD)](providers/banks/td.md) - Toronto-Dominion Bank

### 💰 支付工具
- [支付宝 (Alipay)](providers/payment/alipay.md) - 支持 CSV 格式
- [微信 (WeChat)](providers/payment/wechat.md) - 支持 CSV 和 XLSX 格式

### 📈 证券交易
- [海通证券 (HTSEC)](providers/securities/htsec.md) - 证券交易记录
- [华西证券 (HXSEC)](providers/securities/hxsec.md) - 证券交易记录

### 🪙 加密货币
- [火币 (Huobi)](providers/crypto/huobi.md) - 币币交易记录

### 🛒 生活服务
- [美团 (MT)](providers/life/mt.md) - 美团外卖/到店账单
- [京东 (JD)](providers/life/jd.md) - 京东购物账单

## 配置指南

- [配置总览](configuration/README.md) - 了解配置文件结构
- [规则配置](configuration/rules.md) - 学习如何编写匹配规则
- [账户映射](configuration/accounts.md) - 设置账户对应关系

## 示例

- [基本使用示例](examples/basic-usage.md)

## 输出格式

支持两种复式记账格式：

- **Beancount** - Python生态的复式记账系统
- **Ledger** - 命令行复式记账系统

## 贡献

欢迎贡献代码和文档！请查看我们的 [GitHub 仓库](https://github.com/deb-sig/double-entry-generator)。

## 许可证

本项目采用 Apache 2.0 许可证。