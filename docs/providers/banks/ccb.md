# 建设银行 (CCB) Provider

建设银行 Provider 支持将建设银行账单转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式
- XLS 格式
- XLSX 格式

## 使用方法

### 基本命令

```bash
# 转换建设银行账单
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:CCB
defaultCurrency: CNY
title: 建设银行账单转换

ccb:
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
- `defaultCashAccount`: 建设银行账户
- `defaultCurrency`: 默认货币

### 规则配置

建设银行 Provider 提供基于规则的匹配，可以指定：

- `item`（交易描述）的完全/包含匹配
- `peer`（交易对方）的完全/包含匹配
- `type`（交易类型）的完全/包含匹配
- `status`（交易状态）的完全/包含匹配
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

## 文件格式说明

建设银行账单文件通常包含以下信息：

1. **文件头部**：包含银行信息和账户信息
2. **数据行**：包含交易记录
3. **文件尾部**：包含统计信息

### 数据字段

- 记账日：交易记账日期
- 交易日期：实际交易日期
- 交易金额：交易金额（正数为收入，负数为支出）
- 交易类型：交易类型描述
- 交易对方：交易对方信息
- 交易状态：交易状态（成功、失败等）

## 示例文件

- [建设银行账单示例](../../example/ccb/建设银行_xxxx_2025xxxx_2025xxxx.xls)
- [配置示例](../../example/ccb/config.yaml)
- [输出示例](../../example/ccb/example-ccb-output.beancount) 