#!/bin/bash

# 构建支持文件读取的 WASM 模块
# 用于 BeanBridge 项目

set -e

echo "开始构建 WASM 文件读取模块..."

# 设置环境变量
export GOOS=js
export GOARCH=wasm

# 构建目录
BUILD_DIR="/home/fatsheep/windows/documents/programs/beancount/BeanBridge/public/wasm"
SOURCE_DIR="cmd/wasm"

# 确保构建目录存在
mkdir -p "$BUILD_DIR"

# 构建 WASM 文件
echo "编译 WASM 模块..."
go build -o "$BUILD_DIR/double-entry-generator.wasm" "$SOURCE_DIR/main.go"

# 复制 wasm_exec.js
echo "复制 wasm_exec.js..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$BUILD_DIR/"

echo "WASM 构建完成！"
echo "输出文件："
echo "  - $BUILD_DIR/double-entry-generator.wasm"
echo "  - $BUILD_DIR/wasm_exec.js"

# 检查文件大小
echo ""
echo "文件大小："
ls -lh "$BUILD_DIR/double-entry-generator.wasm"
ls -lh "$BUILD_DIR/wasm_exec.js"

