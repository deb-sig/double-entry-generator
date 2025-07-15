# Double Entry Generator 文档

欢迎使用 Double Entry Generator！这是一个基于规则的双重记账导入器，支持从支付宝、微信、建设银行等多种账单格式转换为 Beancount/Ledger 格式。

## 快速开始

### 安装

```bash
go install github.com/deb-sig/double-entry-generator/v2@latest
```

### 基本使用

```bash
# 转换支付宝账单
double-entry-generator translate -p alipay -t beancount alipay_records.csv

# 转换微信账单
double-entry-generator translate -p wechat -t beancount wechat_records.csv

# 转换建设银行账单
double-entry-generator translate -p ccb -t beancount ccb_records.xls
```

## 支持的 Provider

- [支付宝 (Alipay)](providers/alipay.md)
- [微信 (WeChat)](providers/wechat.md)
- [建设银行 (CCB)](providers/ccb.md)
- [中信银行信用卡 (CITIC)](providers/citic.md)
- [汇丰银行香港 (HSBC HK)](providers/hsbchk.md)
- [华西证券 (HXSEC)](providers/hxsec.md)
- [华泰证券 (HTSEC)](providers/htsec.md)
- [工商银行 (ICBC)](providers/icbc.md)
- [京东 (JD)](providers/jd.md)
- [美团 (MT)](providers/mt.md)
- [火币 (Huobi)](providers/huobi.md)
- [TD Ameritrade (TD)](providers/td.md)
- [BMO (BMO)](providers/bmo.md)

## 配置指南

- [配置文件说明](configuration/config.md)
- [规则配置详解](configuration/rules.md)
- [账户配置](configuration/accounts.md)

## 示例

- [支付宝配置示例](examples/alipay-config.yaml)
- [微信配置示例](examples/wechat-config.yaml)
- [建设银行配置示例](examples/ccb-config.yaml)

## 贡献指南

如果您想为项目贡献代码或文档，请查看我们的 [贡献指南](../CONTRIBUTING.md)。

## 许可证

本项目采用 Apache 2.0 许可证。详情请查看 [LICENSE](../LICENSE) 文件。 