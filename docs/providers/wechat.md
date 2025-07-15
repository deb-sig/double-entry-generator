# 微信 (WeChat) Provider

微信 Provider 支持将微信账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式（默认）
- XLSX 格式（新增支持）

## 使用方法

### 基本命令

```bash
# 转换 CSV 格式的微信账单
double-entry-generator translate -p wechat -t beancount wechat_records.csv

# 转换 XLSX 格式的微信账单
double-entry-generator translate -p wechat -t beancount wechat_records.xlsx
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:WeChat
defaultCurrency: CNY
title: 微信账单转换

wechat:
  rules:
    - item: 三快
      targetAccount: Expenses:Food
    - item: 电子商务,天猫,京东,特约商户
      targetAccount: Expenses:Shopping
    - item: 电费,网上国网
      targetAccount: Expenses:Electricity
    - item: 滴滴出行,嘀嘀,中国石油
      targetAccount: Expenses:Transport
    - item: 现金奖励
      targetAccount: Income:Rewards
    - item: 财付通(银联云闪付)
      ignore: true
    - item: 财付通还款
      targetAccount: Assets:WeChat
```

## 配置说明

### 全局配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCashAccount`: 微信账户（等同于支付宝中的 `methodAccount`）
- `defaultCurrency`: 默认货币

### 规则配置

微信 Provider 提供基于规则的匹配，可以指定：

- `item`（交易描述）的完全/包含匹配
- `peer`（交易对方）的完全/包含匹配
- `type`（交易类型）的完全/包含匹配
- `time`（交易时间）的区间匹配
- `minPrice`（最小金额）和 `maxPrice`（最大金额）的区间匹配

### 规则选项

- `sep`: 分隔符，默认为 `,`
- `fullMatch`: 是否使用完全匹配，默认为 `false`
- `tag`: 设置流水的 Tag
- `ignore`: 是否忽略匹配的交易，默认为 `false`

## 账户关系

`targetAccount` 与 `defaultCashAccount` 的增减账户关系：

| 收/支 | minusAccount | plusAccount |
|-------|-------------|-------------|
| 收入 | targetAccount | defaultCashAccount |
| 支出 | defaultCashAccount | targetAccount |

## 示例文件

- [微信账单示例 (CSV)](../../example/wechat/example-wechat-records.csv)
- [微信账单示例 (XLSX)](../../example/wechat/example-wechat-records.xlsx)
- [配置示例](../../example/wechat/config.yaml) 