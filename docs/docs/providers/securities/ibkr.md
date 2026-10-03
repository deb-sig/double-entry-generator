---
title: Interactive Brokers (IBKR)
---

# Interactive Brokers (IBKR) Provider

IBKR Provider 支持将 Interactive Brokers Flex Query XML 导出的交易记录转换为 Beancount/Ledger 格式。

## 支持的文件格式

- Flex Query XML

## 支持的记录类型

Provider 会读取 Flex XML 中的以下节点：

- `Trade` + `assetCategory="STK"`: 股票、ETF、ADR 买卖，输出为证券交易。
- `Trade` + `assetCategory="CASH"`: 外汇交易，例如 `USD.HKD`，输出为换汇交易。
- `CashTransaction`: 入金/出金、股息、预扣税、利息和费用等现金流水。

`StatementOfFundsLine` 是资金流水视图，容易与 `Trade` / `CashTransaction` 重复，当前不会作为主数据源转换。

## 使用方法

### 基本命令

```bash
double-entry-generator translate -p ibkr -t beancount \
  --config example/ibkr/config.yaml \
  example/ibkr/example-ibkr-records.xml
```

Ledger 输出：

```bash
double-entry-generator translate -p ibkr -t ledger \
  --config example/ibkr/config.yaml \
  example/ibkr/example-ibkr-records.xml
```

### 配置文件

```yaml
defaultMinusAccount: Equity:Opening-Balances
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:IBKR:Cash
defaultPositionAccount: Assets:IBKR:Positions
defaultCommissionAccount: Expenses:IBKR:Fees
defaultPnlAccount: Income:IBKR:PnL
defaultCurrency: USD
bookingMethod: FIFO
title: IBKR

ibkr:
  rules:
    - item: "TSLA,NIO"
      sep: ","
      positionAccount: Assets:IBKR:Positions:Stocks
    - type: "Withholding Tax"
      commissionAccount: Expenses:IBKR:Tax
```

## 配置说明

### 全局账户

- `defaultCashAccount`: IBKR 现金账户。
- `defaultPositionAccount`: 默认证券持仓账户。
- `defaultCommissionAccount`: 默认手续费/税费账户。
- `defaultPnlAccount`: 默认收益账户，用于股息、利息和证券卖出损益补足行。
- `defaultMinusAccount`: 入金交易的默认来源账户。
- `bookingMethod`: 可选 Beancount 记账方式。示例使用 `FIFO`，用于避免多批次证券卖出时的 lot 匹配歧义。

### 规则字段

IBKR 规则支持：

- `item`: 证券名称或现金交易描述
- `type`: 原始类型，例如 `BUY`、`SELL`、`Withholding Tax`
- `sep`: 多值分隔符，默认为 `,`
- `fullMatch`: 是否完整匹配
- `ignore`: 匹配后忽略该交易
- `cashAccount`: 现金账户
- `positionAccount`: 持仓账户
- `commissionAccount`: 手续费/税费账户
- `pnlAccount`: 损益账户

## 输出细节

IBKR Provider 会尽量保留原始 metadata，包括 `account_id`、`trade_id`、`transaction_id`、`ib_order_id`、`ib_exec_id`、`exchange`、`isin`、`security_id`、`report_date` 和 `settle_date`。

证券买入会输出现金减少、持仓增加和手续费；证券卖出会输出持仓减少、现金增加、手续费和 PnL 补足行。现金类外汇交易会输出 `CurrencyExchange` 分录。

## 示例文件

- [配置示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/config.yaml)
- [Flex XML 输入示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-records.xml)
- [Beancount 输出示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-output.beancount)
- [Ledger 输出示例](https://github.com/deb-sig/double-entry-generator/blob/master/example/ibkr/example-ibkr-output.ledger)
