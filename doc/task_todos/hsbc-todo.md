# HSBC HK 账单转换功能实现计划

> Mostly written by cursor.

## 总体任务

为 double-entry-generator 添加支持 HSBC HK 的借记卡和信用卡账单转换功能，参照现有的 ICBC 实现方式。

## 详细任务列表

### 1. 代码实现

- [x] 在 `pkg/consts/consts.go` 中添加 HSBC HK 提供商的常量定义
- [x] 创建 `pkg/provider/hsbchk` 目录结构
  - [x] 实现 `types.go` 定义 HSBC HK 账单相关的数据结构
  - [x] 实现 `config.go` 提供配置结构
  - [x] 实现 `parse.go` 解析 HSBC HK 账单
  - [x] 实现 `convert.go` 将 HSBC HK 账单转换为中间表示
  - [x] 实现 `hsbchk.go` 主要逻辑，识别借记卡和信用卡模式
- [x] 在 `pkg/provider/interface.go` 注册 HSBC HK 提供商
- [x] 创建 `pkg/analyser/hsbchk` 目录结构
  - [x] 实现 `hsbchk.go` 分析器逻辑
- [x] 在 `pkg/analyser/interface.go` 注册 HSBC HK 分析器

### 2. 示例文件

- [x] 创建 `example/hsbchk` 目录
  - [x] 在 `example/hsbchk/debit` 中添加借记卡示例
    - [x] 添加示例账单 CSV 文件
    - [x] 添加配置文件 `config.yaml`
    - [x] ~~添加期望输出的 Beancount 和 Ledger 文件~~
  - [x] 在 `example/hsbchk/credit` 中添加信用卡示例
    - [x] 添加示例账单 CSV 文件
    - [x] 添加配置文件 `config.yaml`
    - [x] ~~添加期望输出的 Beancount 和 Ledger 文件~~

### 3. 测试脚本

- [x] 创建 `test/hsbchk-test-beancount.sh` 测试脚本
- [x] 创建 `test/hsbchk-test-ledger.sh` 测试脚本

### 4. 文档更新

- [x] 在 README.md 中添加 HSBC HK 支持信息
  - [x] 在支持账单列表中添加 HSBC HK
  - [x] 添加 HSBC HK 示例命令
  - [x] 添加 HSBC HK 账单下载与格式说明
  - [x] 添加 HSBC HK 配置示例和说明
  
### 5. 测试

- [x] 执行测试脚本验证功能
  - [x] 借记卡账单转换测试
  - [x] 信用卡账单转换测试
  - [x] Beancount 格式输出测试
  - [x] Ledger 格式输出测试

### 6. Makefile更新

- [x] 在 Makefile 中添加 HSBC HK 测试 target
  - [x] 添加 `test-hsbchk-beancount` 和 `test-hsbchk-ledger` target
  - [x] 将 HSBC HK 测试 target 添加到 `test` target 依赖中
  - [x] 确保测试命令正确引用测试脚本
