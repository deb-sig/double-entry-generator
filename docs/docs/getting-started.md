---
title: 快速开始
description: 快速上手 Double Entry Generator
---


# 快速开始

Double Entry Generator 是一个基于规则的复式记账导入器，支持将各种账单格式转换为 Beancount 或 Ledger 格式。

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

## 完整示例：支付宝账单转换

下面通过一个完整的支付宝账单转换示例，展示如何使用 Double Entry Generator。

### 1. 准备账单文件

从支付宝下载 CSV 格式的账单文件。支付宝账单通常包含以下字段：
- 交易时间
- 交易分类
- 交易对方
- 商品说明
- 收/支
- 金额
- 账户余额
- 交易渠道
- 交易订单号

### 2. 创建配置文件

创建 `alipay_config.yaml` 配置文件：

```yaml
defaultMinusAccount: Assets:Alipay:Cash
defaultPlusAccount: Expenses:FIXME
defaultCurrency: CNY
title: 我的支付宝账单

alipay:
  rules:
    # 餐饮按时间分类
    - category: 餐饮美食
      time: "07:00-11:00"
      targetAccount: Expenses:Food:Breakfast
    - category: 餐饮美食
      time: "11:00-15:00"
      targetAccount: Expenses:Food:Lunch
    - category: 餐饮美食
      time: "17:00-22:00"
      targetAccount: Expenses:Food:Dinner
    
    # 交通出行
    - peer: "滴滴出行,高德打车"
      sep: ","
      targetAccount: Expenses:Transport:Taxi
    
    # 网购
    - peer: "天猫,京东"
      sep: ","
      targetAccount: Expenses:Shopping:Online
    
    # 支付方式账户映射
    - method: "余额宝"
      methodAccount: Assets:Alipay:YuEBao
    - method: "余额"
      methodAccount: Assets:Alipay:Cash
```

### 3. 执行转换

```bash
double-entry-generator translate \
  --provider alipay \
  --target beancount \
  --config alipay_config.yaml \
  --output my_alipay.beancount \
  alipay_202501.csv
```

### 4. 查看结果

生成的 `my_alipay.beancount` 文件示例：

```beancount
option "title" "我的支付宝账单"
option "operating_currency" "CNY"

1970-01-01 open Assets:Alipay:Cash
1970-01-01 open Assets:Alipay:YuEBao
1970-01-01 open Expenses:Food:Lunch
1970-01-01 open Expenses:Transport:Taxi

2025-01-15 * "滴滴出行" "快车" 
    Expenses:Transport:Taxi     23.50 CNY
    Assets:Alipay:Cash         -23.50 CNY

2025-01-15 * "某餐厅" "午餐" 
    Expenses:Food:Lunch         35.00 CNY
    Assets:Alipay:YuEBao       -35.00 CNY
```

### 配置文件说明

- **defaultMinusAccount**: 默认的资产账户（钱从哪里来）
- **defaultPlusAccount**: 默认的支出账户（钱花到哪里去）
- **defaultCurrency**: 默认货币单位
- **alipay.rules**: 匹配规则列表，按顺序匹配，后面的规则会覆盖前面的设置

### 配置文件位置

- 通过 `-c` / `--config` 参数指定路径时，使用该文件（例如当前目录的 `config.yaml`）。
- 未指定时，程序在用户主目录下查找名为 `.double-entry-generator` 的文件（扩展名可为 `.yaml` 或 `.yml`），例如 `~/.double-entry-generator.yaml`。

## 下一步

- 📖 查看 [基本使用示例](examples/basic-usage.md) - 了解更多实际使用场景（微信、银行账单等）
- ⚙️ 查看 [配置指南](configuration/README.md) - 了解详细的配置选项和规则编写
- 📋 浏览 [支持的 Providers](providers.md) - 查看所有支持的数据源
## 常见问题

### Q: 如何处理不支持的账单格式？

A: 可以：
1. 查看是否有类似的 provider 可以参考
2. 在 GitHub 上提交 issue 请求支持
3. 贡献代码添加新的 provider

### Q: 如何自定义账户映射？

A: 通过**默认账户**配置（`defaultMinusAccount`、`defaultPlusAccount`、`defaultCashAccount` 等）设置兜底账户，通过各 provider 下的**规则**（如 `alipay.rules`）中的 `targetAccount`、`methodAccount`、`pnlAccount` 等字段按匹配结果指定账户。没有单独的 `accounts` 配置段，账户由规则与默认值共同决定。详见 [配置指南](configuration/README.md) 与 [账户映射](configuration/accounts.md)。

### Q: 输出文件编码问题？

A: 确保输入文件使用 UTF-8 编码，或者通过配置指定正确的编码格式。
