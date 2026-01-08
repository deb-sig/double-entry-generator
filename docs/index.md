---
title: 首页
layout: home
nav_order: 1
description: "基于规则的复式记账导入器"
permalink: /
---

# 基于规则的复式记账导入器

将各种账单格式智能转换为 Beancount 或 Ledger 格式，让复式记账变得简单高效

[快速开始](/double-entry-generator/getting-started/){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 } [GitHub](https://github.com/deb-sig/double-entry-generator){: .btn .fs-5 .mb-4 .mb-md-0 }

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
- [建设银行 (CCB)](/double-entry-generator/providers/ccb/) - 支持 CSV、XLS、XLSX 格式
- [工商银行 (ICBC)](/double-entry-generator/providers/icbc/) - 自动识别借记卡/信用卡
- [中信银行 (CITIC)](/double-entry-generator/providers/citic/) - 信用卡账单
- [汇丰银行香港 (HSBC HK)](/double-entry-generator/providers/hsbchk/) - 香港汇丰银行
- [加拿大银行 (BMO)](/double-entry-generator/providers/bmo/) - Bank of Montreal
- [道明银行 (TD)](/double-entry-generator/providers/td/) - Toronto-Dominion Bank

### 💰 支付工具
- [支付宝 (Alipay)](/double-entry-generator/providers/alipay/) - 支持 CSV 格式
- [微信 (WeChat)](/double-entry-generator/providers/wechat/) - 支持 CSV 和 XLSX 格式

### 📈 证券交易
- [海通证券 (HTSEC)](/double-entry-generator/providers/htsec/) - 证券交易记录
- [华西证券 (HXSEC)](/double-entry-generator/providers/hxsec/) - 证券交易记录

### 🪙 加密货币
- [火币 (Huobi)](/double-entry-generator/providers/huobi/) - 币币交易记录

### 🛒 生活服务
- [美团 (MT)](/double-entry-generator/providers/mt/) - 美团外卖/到店账单
- [京东 (JD)](/double-entry-generator/providers/jd/) - 京东购物账单

## 配置指南

- [配置总览](/double-entry-generator/configuration/readme/) - 了解配置文件结构
- [规则配置](/double-entry-generator/configuration/rules/) - 学习如何编写匹配规则
- [账户映射](/double-entry-generator/configuration/accounts/) - 设置账户对应关系

## 示例

- [基本使用示例](/double-entry-generator/examples/basic-usage/)
- [高级规则配置](/double-entry-generator/examples/advanced-rules/)

## 输出格式

支持两种复式记账格式：

- **Beancount** - Python生态的复式记账系统
- **Ledger** - 命令行复式记账系统

## 贡献

欢迎贡献代码和文档！请查看我们的 [GitHub 仓库](https://github.com/deb-sig/double-entry-generator)。

## 许可证

本项目采用 Apache 2.0 许可证。