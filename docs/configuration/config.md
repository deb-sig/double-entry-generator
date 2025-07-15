# 配置文件说明

Double Entry Generator 使用 YAML 格式的配置文件来定义转换规则和账户映射。

## 配置文件结构

```yaml
# 全局配置
title: 账单转换配置
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank
defaultCurrency: CNY

# Provider 特定配置
alipay:
  rules:
    # 支付宝规则...

wechat:
  rules:
    # 微信规则...

ccb:
  rules:
    # 建设银行规则...
```

## 全局配置项

### 必需配置

- `defaultMinusAccount`: 默认金额减少的账户
- `defaultPlusAccount`: 默认金额增加的账户
- `defaultCurrency`: 默认货币

### 可选配置

- `title`: 配置标题
- `defaultCashAccount`: 默认现金账户（用于银行类 Provider）
- `defaultPositionAccount`: 默认持仓账户（用于证券类 Provider）
- `defaultCommissionAccount`: 默认手续费账户（用于证券类 Provider）
- `defaultPnlAccount`: 默认盈亏账户（用于证券类 Provider）

## Provider 配置

每个 Provider 都有自己的配置节，包含规则列表：

```yaml
provider_name:
  rules:
    - item: 关键词
      targetAccount: Expenses:Category
      sep: ","
      fullMatch: false
      tag: "tag1,tag2"
      ignore: false
```

## 规则配置详解

### 匹配字段

- `item`: 交易描述匹配
- `peer`: 交易对方匹配
- `type`: 交易类型匹配
- `status`: 交易状态匹配
- `time`: 交易时间匹配
- `minPrice`/`maxPrice`: 金额范围匹配

### 规则选项

- `targetAccount`: 目标账户
- `methodAccount`: 方法账户
- `commissionAccount`: 手续费账户
- `sep`: 分隔符（用于多个关键词）
- `fullMatch`: 是否完全匹配
- `tag`: 标签（多个标签用分隔符分隔）
- `ignore`: 是否忽略匹配的交易

## 账户命名规范

建议使用以下账户命名规范：

```
Assets:Bank:CCB          # 银行账户
Assets:WeChat            # 微信账户
Assets:Alipay            # 支付宝账户
Expenses:Food            # 餐饮支出
Expenses:Shopping        # 购物支出
Expenses:Transport       # 交通支出
Expenses:Electricity     # 电费支出
Income:Salary            # 工资收入
Income:Rewards           # 奖励收入
```

## 配置文件示例

### 基础配置

```yaml
title: 个人账单转换
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:CCB
defaultCurrency: CNY

wechat:
  rules:
    - item: 三快
      targetAccount: Expenses:Food
    - item: 滴滴出行
      targetAccount: Expenses:Transport
```

### 高级配置

```yaml
title: 完整账单转换
defaultMinusAccount: Assets:FIXME
defaultPlusAccount: Expenses:FIXME
defaultCashAccount: Assets:Bank:CCB
defaultCurrency: CNY

wechat:
  rules:
    - item: 三快,美团
      targetAccount: Expenses:Food
      sep: ","
      tag: "food,delivery"
    - item: 滴滴出行
      targetAccount: Expenses:Transport
      fullMatch: true
      tag: "transport"
    - item: 财付通还款
      targetAccount: Assets:WeChat
      ignore: false
```

## 配置文件位置

默认情况下，程序会在以下位置查找配置文件：

1. 当前目录的 `config.yaml`
2. 用户主目录的 `.double-entry-generator.yaml`
3. 通过 `-c` 参数指定的配置文件

## 验证配置

可以使用以下命令验证配置文件：

```bash
double-entry-generator translate -p wechat -t beancount --config config.yaml input.csv
``` 