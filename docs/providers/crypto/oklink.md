---
title: OKLink 多链代币
layout: default
parent: 提供商支持
nav_order: 2
---

# OKLink 多链代币 Provider

OKLink Provider 支持从 OKLink 区块链浏览器导出的代币转账 CSV 文件，支持多链、多地址配置。

## 支持的文件格式

- CSV 格式（从 OKLink 区块链浏览器导出）

## 支持的链

**当前支持**：
- ✅ **Ethereum (ERC20)** - 已测试可用
- ✅ **TRON (TRC20)** - 已测试可用

> ⚠️ **注意**：目前仅支持 Ethereum 和 TRON 链的代币转账记录。其他链（如 BSC、Polygon、Arbitrum 等）虽然 OKLink 可能支持导出，但由于 CSV 字段格式可能不一致，暂不支持。如需支持其他链，请提交 Issue 并提供示例 CSV 文件。

## 使用方法

### 基本命令

```bash
# 转换 OKLink 代币转账记录
deg translate \
    -p oklink \
    -c config.yaml \
    -o output.bean \
    your-export.csv

# 或使用完整命令
double-entry-generator translate -p oklink -t beancount -c config.yaml your-export.csv
```

### 导出 CSV 文件

从 OKLink 区块链浏览器导出代币转账记录：

1. 访问 https://www.oklink.com/
2. 搜索你的钱包地址（Ethereum 或 TRON）
3. 选择 "Token Transfer" 标签
4. 点击 "导出 CSV" 按钮
5. 选择时间范围并导出

**优势**：OKLink 无需注册账户，支持 Ethereum 和 TRON 链的代币转账导出。

### 配置文件

创建配置文件 `config.yaml`：

#### 单地址配置

```yaml
title: "OKLink 代币账本"
defaultCurrency: "CNY"
defaultMinusAccount: "Assets:FIXME"
defaultPlusAccount: "Assets:FIXME"

oklink:
  # 地址作为 key（单地址时也使用相同格式）
  "0x1429****7f855c":  # Ethereum 地址示例
    rules:
      # USDT 配置
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        tags: "Stablecoin,USDT"
      
      # USDC 配置
      - tokenSymbol: "USDC"
        methodAccount: "Assets:Crypto:Ethereum:USDC"
        tags: "Stablecoin,USDC"
```

#### 多地址配置（推荐）

```yaml
title: "OKLink 多链代币账本"
defaultCurrency: "CNY"
defaultMinusAccount: "Assets:FIXME"
defaultPlusAccount: "Assets:FIXME"

oklink:
  # Ethereum 地址
  "0x1429****7f855c":  
    rules:
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        tags: "Stablecoin,USDT"
  
  # TRON 地址
  "TExam****7890":  
    rules:
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:TRON:USDT"
        tags: "Stablecoin,USDT,TRON"
```


> 为什么需要多地址配置?适用于一个链上拥有多个地址的情况,不需要频繁修改配置文件就可以导出多个地址的账单
> 
> 在这里写明地址是因为需要根据这里的地址计算交易方向, 每次导出账单的时候只会应用对应地址下的规则配置,如果涉及到两个hd地址都配置的情况,则两个地址的规则配置会同时应用

## 配置说明

### 全局配置

```yaml
title: "OKLink 代币账本"
defaultCurrency: "CNY"  # 不影响 OKLink，OKLink 自动使用代币符号
defaultMinusAccount: "Assets:FIXME"  # 默认借方账户
defaultPlusAccount: "Expenses:FIXME"   # 默认贷方账户
```

### OKLink 配置

#### 配置格式

**地址作为 key**：每个地址可以有独立的配置

系统会**自动识别** CSV 文件中的 `from` 和 `to` 地址，并匹配配置中的规则：
- 如果 `from` 地址在配置中，按 **send**（发送）处理，使用该地址的规则
- 如果 `from` 不在配置中，但 `to` 地址在配置中，按 **recv**（接收）处理，使用该地址的规则
- 如果两个地址都在配置中，优先使用 `from` 地址的规则（send 视角）
- 如果两个地址都不在配置中，跳过该交易

```yaml
oklink:
  "0x...":  # Ethereum 地址（0x 开头）
    rules: [...]
  
  "T...":   # TRON 地址（T 开头）
    rules: [...]
```

> 💡 **提示**：一般一个账单文件只包含一个地址的交易记录。你只需要在配置文件中配置你拥有的地址即可，系统会自动匹配。

#### 地址配置字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `rules` | array | ❌ | 规则列表 |

#### 规则匹配

规则按顺序匹配，第一个匹配成功的规则将被应用。

##### 匹配条件

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `tokenSymbol` | string | 代币符号 | `"USDT"`, `"UNI"` |
| `tokenName` | string | 代币名称 | `"Tether USD"` |
| `contractAddress` | string | 合约地址 | `"0xdac17f958..."` |
| `from` | string | 发送地址 | `"0x123..."` |
| `to` | string | 接收地址 | `"0x456..."` |
| `peer` | string | 对方地址 | `"0x789..."` |
| `minAmount` | float | 最小金额 | `1.0` |
| `maxAmount` | float | 最大金额 | `1000.0` |
| `txHash` | string | 交易哈希 | `"0xabc..."` |
| `time` | string | 时间范围 | `"2024-01-01~2024-12-31"` |

##### 应用配置

| 字段 | 类型 | 说明 |
|------|------|------|
| `methodAccount` | string | 资产账户（代币账户） |
| `targetAccount` | string | 目标账户（收入/支出） |
| `currency` | string | 自定义货币单位（默认使用 tokenSymbol） |
| `tags` | string | 标签，逗号分隔 |
| `ignore` | bool | 是否忽略此交易 |
| `note` | string | 自定义备注 |

### 账户分配逻辑

#### 收款（recv）
```
+ methodAccount  (资产增加)
- targetAccount  (收入减少) 或 defaultMinusAccount
```

#### 付款（send）
```
- methodAccount  (资产减少)
+ targetAccount  (支出增加) 或 defaultPlusAccount
```

## 配置示例

### 示例 1: 单地址基础配置

```yaml
title: "OKLink 代币账本"
defaultCurrency: "CNY"
defaultMinusAccount: "Assets:FIXME"
defaultPlusAccount: "Assets:FIXME"

oklink:
  # 地址作为 key（单地址时也使用相同格式）
  "0x1429****7f855c":  # Ethereum 地址示例
    rules:
      # USDT 配置
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        tags: "Stablecoin,USDT"
      
      # USDC 配置
      - tokenSymbol: "USDC"
        methodAccount: "Assets:Crypto:Ethereum:USDC"
        tags: "Stablecoin,USDC"
```

### 示例 2: 多地址多链配置

```yaml
title: "OKLink 多链代币账本"
defaultCurrency: "CNY"
defaultMinusAccount: "Assets:FIXME"
defaultPlusAccount: "Assets:FIXME"

oklink:
  # Ethereum 地址
  "0x1429****7f855c":
    rules:
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        tags: "Stablecoin,USDT"
      
      - tokenSymbol: "USDC"
        methodAccount: "Assets:Crypto:Ethereum:USDC"
        tags: "Stablecoin,USDC"
  
  # TRON 地址
  "TExample123456789012345678901234567890":
    rules:
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:TRON:USDT"
        tags: "Stablecoin,USDT,TRON"
```

### 示例 3: 按交易方向匹配

```yaml
oklink:
  "0x1429****7f855c":  
    rules:
      # USDT 收款（从指定地址收款）
      - tokenSymbol: "USDT"
        peer: "0xd854****c172b3"
        direction: "recv"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        targetAccount: "Income:Crypto:Transfer"
        tags: "USDT,Receive"
      
      # USDT 付款（向指定地址付款,修改支出账户为Expenses:Food:Meal,并修改货币单位为CNY,但是并不能自动帮你换算汇率）
      - tokenSymbol: "USDT"
        peer: "0x3ef4****f2768a"
        direction: "send"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        targetAccount: "Expenses:Food:Meal"
        currency: "CNY"
        tags: "USDT,Send"
```

### 示例 4: 按金额范围区分

```yaml
oklink:
  "0x1429****7f855c":  
    rules:
      # USDT 大额转账（>= 1000）
      - tokenSymbol: "USDT"
        minAmount: 1000.0
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        targetAccount: "Income:Crypto:Business"
        tags: "LargeTransfer"
      
      # USDT 小额空投（1-10）
      - tokenSymbol: "USDT"
        minAmount: 1.0
        maxAmount: 10.0
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        targetAccount: "Income:Crypto:Airdrop"
        tags: "SmallAirdrop"
      
      # USDT 普通转账（其他金额）
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Ethereum:USDT"
        targetAccount: "Income:Crypto:Transfer"
        tags: "USDT"
```

### 示例 5: 使用软件钱包地址作为账户

```yaml
oklink:
  "0x1429****7f855c":  
    rules:
      # 使用完整地址作为账户的一部分
      - tokenSymbol: "USDT"
        methodAccount: "Assets:Crypto:Software:Ethereum:0x1429****7f855c"
        tags: "SoftwareWallet,USDT"
      
      # 或使用简化的地址标识
      - tokenSymbol: "USDC"
        methodAccount: "Assets:Crypto:Software:Ethereum:0x1429:USDC"
        tags: "SoftwareWallet,USDC"
```

## 输出示例

```beancount
option "title" "OKLink 代币账本"
option "operating_currency" "CNY"

1970-01-01 open Assets:Crypto:Ethereum:USDT
1970-01-01 open Assets:Crypto:Ethereum:USDC
1970-01-01 open Assets:Crypto:Software:Ethereum:0x1429****7f855c
1970-01-01 open Assets:FIXME
1970-01-01 open Expenses:Food:Meal
1970-01-01 open Income:Crypto:Airdrop
1970-01-01 open Income:Crypto:Transfer

2025-11-09 * "0xd854****c172b3" "USDT Receive"
	blockNo: "23759705"
	contractAddress: "0xff7b****1b9f44"
	direction: "recv"
	from: "0xd854****c172b3"
	to: "0x1429****7f855c"
	tokenName: "USDT"
	tokenSymbol: "USDT"
	txHash: "0xdedce3****0177adc9"
	Assets:Crypto:Ethereum:USDT 353.75150000 USDT
	Income:Crypto:Transfer -353.75150000 USDT

2025-11-09 * "0x3ef4****f2768a" "USDT Send"
	blockNo: "23759694"
	contractAddress: "0xdaee****71b2e1"
	direction: "send"
	from: "0x1429****7f855c"
	to: "0x3ef4****f2768a"
	tokenName: "USDT"
	tokenSymbol: "USDT"
	txHash: "0xb92a4b****867bcde5"
	Assets:Crypto:Ethereum:USDT -408.50000000 USDT
	Expenses:Food:Meal 408.50 CNY
```

## 特性

### ✅ 多链支持
- 一个站点（OKLink）导出所有链的数据
- 当前支持 Ethereum (ERC20) 和 TRON (TRC20)
- 自动识别链类型（通过地址格式：0x 开头为 Ethereum，T 开头为 TRON）

### ✅ 多地址配置
- 一个配置文件管理多个地址
- 每个地址独立配置规则
- 自动匹配交易对应的地址配置

### ✅ 高精度支持
- 自动保留 8 位小数，适合加密货币
- 正确处理科学计数法（如 `1.0E-6`）
- 正确处理千位分隔符（如 `5,586`）

### ✅ 灵活的账户结构
- 支持任意账户层级
- 完全自定义账户路径
- 支持一个账户记录多个代币单位（如 `Assets:Crypto:Software:Ethereum:0x... CNY,USDT,ETH`）

### ✅ 智能方向判断
- 根据配置的地址自动判断收发方向
- 支持多地址同时处理
- 自动跳过未配置地址的交易

### ✅ 货币单位
- 默认使用代币符号（USDT、UNI、ETH）
- 可在规则中自定义 `currency` 字段

## CSV 格式支持

OKLink 支持两种 CSV 格式：

### Ethereum 格式（中文表头）
```
交易哈希, 区块高度, UTC时间, 发送方, 接收方, 数量, 代币符号, 代币地址
```

### TRON 格式（英文表头）
```
Tx Hash, blockHeight, blockTime(UTC), from, to, value, symbol, tokenAddress
```

Provider 会自动识别并解析两种格式。

## 常见问题

### Q: 为什么有些交易显示为 "address not in config"？
A: 这是因为该交易的地址（from 或 to）没有在配置文件中配置。请为每个需要处理的地址添加配置。

### Q: 如何处理同一笔交易的多个代币？
A: OKLink 导出的 CSV 会将每个代币分开记录，每行都会生成一条 Beancount 记录。

### Q: 支持 NFT 转账吗？
A: 目前只支持代币转账（ERC20/TRC20/BEP20），不支持 NFT（ERC721/ERC1155）。

### Q: 可以处理 Gas 费用吗？
A: Gas 费用需要单独从 "Internal Transactions" 导出处理，当前 provider 只处理代币转账。

### Q: 如何处理跨链桥交易？
A: 建议为每条链单独导出和处理，然后在配置中使用不同的 `defaultCashAccount` 区分。

### Q: 单地址和多地址配置有什么区别？
A: 没有区别！单地址和多地址使用相同的配置格式，只是地址数量不同。单地址时只配置一个地址作为 key 即可。


### Q: 为什么导出账单的时候有的货币单位是地址？
A: 因为 OKLink 导出 Ethereum 账单时，未认证/危险代币会在账单的"代币符号"一栏填充发送方地址（而不是代币符号）。系统会继续处理这些交易，不会跳过。你可以通过以下方式自定义处理：

1. **通过合约地址匹配**：使用 `contractAddress` 字段匹配规则,
   ```yaml
   - contractAddress: "0x..."
     methodAccount: "Assets:Crypto:Ethereum:UnknownToken"
     targetAccount: "Income:Crypto:Transfer"
     tags: "UnknownToken"
   ```

2. **通过代币符号（地址）匹配**：直接匹配地址格式的代币符号
   ```yaml
   - tokenSymbol: "0x..."  # 完整的地址
     methodAccount: "Assets:Crypto:Ethereum:UnknownToken"
     targetAccount: "Income:Crypto:Transfer"
   ```

3. **忽略特定交易**：如果确实不需要记录，可以使用 `ignore: true`
   ```yaml
   - contractAddress: "0x..."
     ignore: true
   ```

> 💡 **提示**：系统会记录警告日志，格式为：`[OKLink] Warning: Token symbol is an address (likely unverified token): tokenSymbol=0x..., contractAddress=0x..., txHash=0x...`。日志中包含合约地址，方便你直接复制到配置文件中进行匹配。

## 示例文件

- [示例 CSV 文件](../../example/oklink/example-oklink-token-transfer.csv)
- [示例配置文件](../../example/oklink/config.yaml)
- [输出示例](../../example/oklink/example-oklink-output.beancount)

## 参考

- [OKLink 官网](https://www.oklink.com/)
- [Beancount 文档](https://beancount.github.io/docs/)
