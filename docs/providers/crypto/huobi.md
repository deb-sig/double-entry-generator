# 火币 (Huobi) Provider

火币 Provider 支持将火币网币币交易记录转换为 Beancount/Ledger 格式。

## 支持的文件格式

- CSV 格式

## 使用方法

### 基本命令

```bash
# 转换火币交易记录
double-entry-generator translate -p huobi -t beancount huobi_records.csv
```

### 配置文件

创建配置文件 `config.yaml`：

```yaml
defaultMinusAccount: Assets:Crypto:Huobi
defaultPlusAccount: Assets:Crypto:Huobi
defaultCashAccount: Assets:Crypto:Huobi
defaultCurrency: USDT
title: 火币交易转换

huobi:
  rules:
    - type: 买入
      symbol: BTC
      targetAccount: Assets:Crypto:BTC
    - type: 卖出
      symbol: BTC
      targetAccount: Assets:Crypto:BTC
      pnlAccount: Income:Crypto:PnL
    - type: 买入
      symbol: ETH
      targetAccount: Assets:Crypto:ETH
```

## 配置说明

### 交易类型

火币支持多种币币交易：
- 现货买入/卖出
- 不同币种之间的兑换
- 手续费记录

### 账户设置

- `Assets:Crypto:Huobi`: 火币账户（通常以USDT计价）
- `Assets:Crypto:BTC`: BTC持仓账户
- `Assets:Crypto:ETH`: ETH持仓账户
- `Income:Crypto:PnL`: 交易损益账户

### 币种支持

支持主流加密货币：
- BTC (Bitcoin)
- ETH (Ethereum)  
- USDT (Tether)
- 以及火币支持的其他币种

## 示例文件

- [火币交易示例](../../example/huobi/example-huobi-records.csv)
- [配置示例](../../example/huobi/config.yaml)
- [输出示例](../../example/huobi/example-huobi-output.beancount)