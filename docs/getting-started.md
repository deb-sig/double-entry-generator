---
title: 快速开始
layout: default
nav_order: 2
description: "快速上手 Double Entry Generator"
permalink: /getting-started/
---

# 快速开始

Double Entry Generator 是一个基于规则的双重记账导入器，支持将各种账单格式转换为 Beancount 或 Ledger 格式。

## 安装

### 使用 Go 安装（推荐）

```bash
# 安装最新版本
go install github.com/deb-sig/double-entry-generator/v2@latest

# 验证安装
double-entry-generator --version
```

### 使用 Homebrew 安装（macOS）

```bash
# 添加 tap
brew tap deb-sig/deb-sig

# 安装
brew install double-entry-generator

# 验证安装
double-entry-generator --version
```

### 使用 releases 构建好的版本
- 点击 [releases](https://github.com/deb-sig/double-entry-generator/releases)
- 选择合适的版本
- 放在本地账本目录合适的目录下
- 记得修改对应的命令哦

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/deb-sig/double-entry-generator.git
cd double-entry-generator

# 构建
make build

# 安装到系统
make install
```

## 基本用法

### 命令行格式

```bash
double-entry-generator translate [选项] <输入文件>
```

### 常用选项

- `-p, --provider`: 指定数据提供者（如 alipay, wechat, ccb 等）
- `-t, --target`: 指定输出格式（beancount 或 ledger）
- `-c, --config`: 指定配置文件路径
- `-o, --output`: 指定输出文件路径

### 示例

#### 转换支付宝账单

```bash
# 转换为 Beancount 格式
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# 转换为 Ledger 格式
double-entry-generator translate -p alipay -t ledger alipay_records.csv
```

#### 转换微信账单

```bash
# 支持 CSV 和 XLSX 格式
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx
```

#### 转换银行账单

```bash
# 建设银行
double-entry-generator translate -p ccb -t beancount ccb_records.xls

# 工商银行
double-entry-generator translate -p icbc -t beancount icbc_records.csv
```

## 配置文件

### 基本配置结构

```yaml
# config.yaml
default:
  # 默认账户设置
  default_minus_account: "Assets:Bank:Checking"
  default_plus_account: "Expenses:Unknown"
  
  # 默认货币
  default_currency: "CNY"
  
  # 默认标签
  default_tags: ["imported"]

# 规则配置
rules:
  - name: "餐饮消费"
    conditions:
      - field: "description"
        contains: ["美团", "饿了么", "餐厅"]
    target_account: "Expenses:Food"
    tags: ["food", "dining"]

# 账户映射
accounts:
  "支付宝": "Assets:Alipay"
  "微信": "Assets:WeChat"
```

### 配置文件位置

1. 当前目录的 `config.yaml`
2. 用户主目录的 `~/.double-entry-generator/config.yaml`
3. 通过 `-c` 参数指定

## 输出格式

### Beancount 格式

```beancount
2024-01-15 * "美团外卖" "午餐"
  Assets:Alipay  -25.00 CNY
  Expenses:Food   25.00 CNY
  # imported
```

### Ledger 格式

```ledger
2024-01-15 * 美团外卖 午餐
    Assets:Alipay  -25.00 CNY
    Expenses:Food   25.00 CNY
    ; imported
```

## 下一步

- 查看 [配置指南]({{ '/configuration/' | relative_url }}) 了解详细配置
- 浏览 [支持的 Providers]({{ '/providers/' | relative_url }}) 查看所有支持的数据源
- 查看 [示例]({{ '/examples/' | relative_url }}) 学习高级用法

## 常见问题

### Q: 如何处理不支持的账单格式？

A: 可以：
1. 查看是否有类似的 provider 可以参考
2. 在 GitHub 上提交 issue 请求支持
3. 贡献代码添加新的 provider

### Q: 如何自定义账户映射？

A: 在配置文件的 `accounts` 部分添加映射关系，支持正则表达式匹配。

### Q: 输出文件编码问题？

A: 确保输入文件使用 UTF-8 编码，或者通过配置指定正确的编码格式。
