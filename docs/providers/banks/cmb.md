---
title: 招商银行 (CMB)
layout: default
parent: 提供商支持
nav_order: 7
---

# 招商银行 (CMB) Provider

招商银行 Provider 支持将招商银行账单转换为 Beancount/Ledger 格式，支持储蓄卡和信用卡账单。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换招商银行储蓄卡账单
double-entry-generator translate -p cmb -t beancount -c config.yaml cmb_records.csv

# 转换招商银行信用卡账单
double-entry-generator translate -p cmb -t beancount -c config.yaml cmb_records.csv
```

### 配置文件

#### 储蓄卡配置示例

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:DebitCard:CMB
defaultCurrency: CNY
title: 招商银行储蓄卡账单转换

cmb:
  rules:
    # 消费
    - peer: 电费,网上国网,国网
      targetAccount: Expenses:Electricity
    - peer: 中国移动
      targetAccount: Expenses:Mobile
    # 保险赔付
    - peer: 太平洋健康保险股份有限公司
      item: 汇入汇款
      targetAccount: Income:Insurance
```

#### 信用卡配置示例

```yaml
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Liabilities:CreditCard:CMB
defaultCurrency: CNY
title: 招商银行信用卡账单转换

cmb:
  rules:
    - item: 掌上生活影票
      targetAccount: Expenses:Movie
    - item: 手机银行饭票
      targetAccount: Expenses:Food
    - item: 中国移动
      targetAccount: Expenses:Mobile
    - item: 财付通
      ignore: true
```

## 配置说明

### 全局配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCashAccount`: 招商银行账户
  - 储蓄卡：`Assets:DebitCard:CMB`
  - 信用卡：`Liabilities:CreditCard:CMB`
- `defaultCurrency`: 默认货币

### 规则配置

招商银行 Provider 提供基于规则的匹配，可以指定：

- `peer`（交易对手）的完全/包含匹配
- `item`（商品描述）的完全/包含匹配
- `type`（交易类型）的完全/包含匹配
- `txType`（交易类型）的完全/包含匹配

### 规则选项

- `sep`: 分隔符，默认为 `,`
- `fullMatch`: 是否使用完全匹配，默认为 `false`
- `tag`: 设置流水的 Tag
- `ignore`: 是否忽略匹配的交易，默认为 `false`
- `methodAccount`: 支付账户（可选）
- `targetAccount`: 目标账户

## 账户关系

`targetAccount` 与 `defaultCashAccount` 的增减账户关系：

| 收/支 | minusAccount       | plusAccount        |
|-------|-------------------|-------------------|
| 收入  | targetAccount     | defaultCashAccount |
| 支出  | defaultCashAccount | targetAccount      |

## 账单下载方式

### 储蓄卡账单

1. 打开招商银行 App
2. 搜索"流水打印"
3. 右下方切换"高级筛选"
4. 选择卡号、起始日期、结束日期
5. 设置账单格式
   - "展示摘要类型"选择"全部"
   - "展示交易对手信息"选择"开启"
   - "展示完整卡号"选择"开启"
   - "展示收入及支出汇总金额"选择"关闭"
   - "交易币种"选择"全部"
   - "金额区间"选择"关闭"
   - "交易类型"选择"全部"
   - "仅展示活期户流水"选择"关闭"
6. 填写接收邮箱地址，确认导出
7. 将导出的 PDF 文件使用 [bill-parser](https://github.com/deb-sig/bill-parser) 转换为 CSV 文件

### 信用卡账单

1. 打开掌上生活 App
2. 搜索"账单补寄"
3. 选择账单周期
4. 提交申请，确认导出
5. 将导出的 PDF 文件使用 [bill-parser](https://github.com/deb-sig/bill-parser) 转换为 CSV 文件

## 示例文件

- [招商银行储蓄卡示例](../../example/cmb/debit/example-cmb-records.csv)
- [招商银行信用卡示例](../../example/cmb/credit/example-cmb-records.csv)
- [储蓄卡配置示例](../../example/cmb/debit/config.yaml)
- [信用卡配置示例](../../example/cmb/credit/config.yaml)
- [输出示例](../../example/cmb/debit/example-cmb-output.beancount)

