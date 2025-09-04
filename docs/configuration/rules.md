# 规则配置详解

规则是 Double Entry Generator 的核心功能，用于将交易自动分类到不同的账户。

## 规则匹配字段

### 通用字段

所有 provider 都支持的字段：

- `type`: 交易类型（如：支出、收入、其他）
- `peer`: 交易对方（商家名称、个人姓名等）
- `item`: 商品/服务描述
- `time`: 时间区间匹配
- `minPrice`: 最小金额
- `maxPrice`: 最大金额

### Provider 特定字段

不同 provider 支持的特殊字段：

#### 支付宝 (alipay)
- `category`: 消费分类（餐饮美食、日用百货等）
- `method`: 支付方式（余额、余额宝、信用卡等）

#### 微信 (wechat)
- `status`: 交易状态

#### 建设银行 (ccb)
- `txType`: 摘要信息
- `status`: 交易状态

## 匹配选项

### 基本选项

```yaml
- peer: "美团"
  fullMatch: true        # 完全匹配，默认为 false
  targetAccount: Expenses:Food
```

### 多关键字匹配

```yaml
- peer: "美团,饿了么,肯德基"
  sep: ","              # 分隔符，默认为 ","
  targetAccount: Expenses:Food
```

### 时间区间匹配

```yaml
- category: 餐饮美食
  time: "11:00-14:00"   # 11点到14点
  targetAccount: Expenses:Food:Lunch
```

### 金额区间匹配

```yaml
- category: 日用百货
  minPrice: 0           # 最小金额
  maxPrice: 10          # 最大金额
  targetAccount: Expenses:Food:Snacks
```

## 账户设置

### 基本账户设置

```yaml
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi    # 目标账户
  methodAccount: Assets:Alipay             # 支付账户（可选）
```

### 投资类交易

```yaml
- peer: "基金"
  type: "其他"
  item: "买入"
  targetAccount: Assets:Alipay:Invest:Fund
  pnlAccount: Income:Alipay:Invest:PnL     # 损益账户
```

## 特殊功能

### 忽略交易

```yaml
- peer: "财付通"
  ignore: true          # 忽略匹配的交易
```

### 添加标签

```yaml
- peer: "滴滴出行"
  targetAccount: Expenses:Transport:Taxi
  tag: "transport,taxi"  # 添加标签
  sep: ","              # 标签分隔符
```

## 规则优先级和覆盖

### 优先级规则

1. 规则按配置文件中的顺序依次匹配
2. 后面的规则会覆盖前面的设置
3. 一笔交易可以匹配多个规则

### 示例

```yaml
rules:
  # 通用规则（优先级低）
  - peer: "美团"
    targetAccount: Expenses:Food
  
  # 特定时间规则（优先级高）
  - peer: "美团"
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch  # 覆盖上面的设置
```

## 最佳实践

### 1. 从宽泛到具体

```yaml
rules:
  # 先设置大类
  - category: 餐饮美食
    targetAccount: Expenses:Food
  
  # 再细化特定情况
  - category: 餐饮美食
    time: "11:00-14:00"
    targetAccount: Expenses:Food:Lunch
```

### 2. 使用多关键字

```yaml
- peer: "美团,饿了么,肯德基,麦当劳"
  sep: ","
  targetAccount: Expenses:Food:FastFood
```

### 3. 合理设置优先级

将特殊规则放在后面，通用规则放在前面。

## 调试技巧

### 1. 逐步添加规则
从简单规则开始，逐步添加复杂规则。

### 2. 使用标签追踪
给重要规则添加标签，便于后续分析。

### 3. 定期检查输出
定期检查生成的账本，调整规则设置。