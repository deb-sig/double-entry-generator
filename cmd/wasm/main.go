//go:build js && wasm
// +build js,wasm

/*
Copyright © 2024 BeanBridge Team

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"syscall/js"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/wasm"
	"github.com/spf13/viper"
)

var (
	fileReader      *wasm.SimpleFileReader
	currentProvider string
	currentConfig   *config.Config
)

func main() {
	// 初始化默认配置
	defaultConfig := &config.Config{
		Title:               "BeanBridge 财务记录",
		DefaultMinusAccount: "Assets:FIXME",
		DefaultPlusAccount:  "Expenses:FIXME",
		DefaultCurrency:     "CNY",
	}

	// 创建文件读取器
	fileReader = wasm.NewSimpleFileReader(defaultConfig)
	currentProvider = "alipay"
	currentConfig = defaultConfig

	// 注册 JavaScript 函数
	js.Global().Set("processFileFromInput", js.FuncOf(processFileFromInput))
	js.Global().Set("processFileFromInputWithFormat", js.FuncOf(processFileFromInputWithFormat))
	js.Global().Set("processFileContent", js.FuncOf(processFileContent))
	js.Global().Set("parseYamlConfig", js.FuncOf(parseYamlConfig))
	js.Global().Set("setProvider", js.FuncOf(setProvider))
	js.Global().Set("getSupportedProviders", js.FuncOf(getSupportedProviders))
	js.Global().Set("getCurrentProvider", js.FuncOf(getCurrentProvider))
	js.Global().Set("testFunction", js.FuncOf(testFunction))
	js.Global().Set("keepAlive", js.FuncOf(keepAlive))

	// 保持程序运行
	select {}
}

// processFileFromInput 处理文件输入元素
func processFileFromInput(this js.Value, args []js.Value) interface{} {
	fileInputID := "fileInput"
	if len(args) > 0 && args[0].Type() == js.TypeString {
		fileInputID = args[0].String()
	}

	fileInput := js.Global().Get("document").Call("getElementById", fileInputID)
	if fileInput.IsNull() {
		return map[string]interface{}{
			"success": false,
			"error":   "文件输入元素未找到",
		}
	}

	files := fileInput.Get("files")
	if files.Length() == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "没有选择文件",
		}
	}

	file := files.Index(0)

	// 创建 FileReader
	reader := js.Global().Get("FileReader").New()

	// 创建 Promise 来处理异步读取
	promise := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result")
			
			// 获取 ArrayBuffer
			arrayBuffer := result
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
			length := uint8Array.Get("length").Int()
			
			// 转换为 Go 字节数组
			data := make([]byte, length)
			js.CopyBytesToGo(data, uint8Array)

			// 获取文件名
			fileName := file.Get("name").String()
			
			log.Printf("[WASM-Main] 文件读取完成：%s, 大小: %d bytes", fileName, len(data))

			// 处理文件（传递文件名和原始字节）
			processResult := fileReader.ProcessFile(fileName, data)

			resolve.Invoke(processResult)
			return nil
		}))

		reader.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			reject.Invoke(map[string]interface{}{
				"success": false,
				"error":   "文件读取失败",
			})
			return nil
		}))

		// 使用 readAsArrayBuffer 而不是 readAsText，避免编码问题
		reader.Call("readAsArrayBuffer", file)
		return nil
	}))

	return promise
}

// processFileFromInputWithFormat 处理文件输入元素，支持指定输出格式
func processFileFromInputWithFormat(this js.Value, args []js.Value) interface{} {
	// 第一个参数：fileInputID (string)
	// 第二个参数：format (string, "beancount" 或 "ledger")
	fileInputID := "fileInput"
	format := "beancount"
	
	if len(args) > 0 && args[0].Type() == js.TypeString {
		fileInputID = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeString {
		format = args[1].String()
	}

	fileInput := js.Global().Get("document").Call("getElementById", fileInputID)
	if fileInput.IsNull() {
		return map[string]interface{}{
			"success": false,
			"error":   "文件输入元素未找到",
		}
	}

	files := fileInput.Get("files")
	if files.Length() == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "没有选择文件",
		}
	}

	file := files.Index(0)

	// 创建 FileReader
	reader := js.Global().Get("FileReader").New()

	// 创建 Promise 来处理异步读取
	promise := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result")
			
			// 获取 ArrayBuffer
			arrayBuffer := result
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
			length := uint8Array.Get("length").Int()
			
			// 转换为 Go 字节数组
			data := make([]byte, length)
			js.CopyBytesToGo(data, uint8Array)

			// 获取文件名
			fileName := file.Get("name").String()
			
			log.Printf("[WASM-Main] 文件读取完成：%s, 格式: %s, 大小: %d bytes", fileName, format, len(data))

			// 处理文件（传递文件名、原始字节和格式）
			processResult := fileReader.ProcessFileWithFormat(fileName, data, format)

			resolve.Invoke(processResult)
			return nil
		}))

		reader.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			reject.Invoke(map[string]interface{}{
				"success": false,
				"error":   "文件读取失败",
			})
			return nil
		}))

		// 使用 readAsArrayBuffer 而不是 readAsText，避免编码问题
		reader.Call("readAsArrayBuffer", file)
		return nil
	}))

	return promise
}

// processFileContent 处理文件内容（暂时不支持直接内容处理）
func processFileContent(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   "请使用 processFileFromInput 处理文件",
	}
}

// parseYamlConfig 解析 YAML 配置
func parseYamlConfig(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error":   "需要配置参数",
		}
	}

	yamlStr := args[0].String()

	// 使用 Viper 解析 YAML 配置（与 CLI 保持一致）
	if err := config.InitConfig(yamlStr); err != nil {
		log.Printf("初始化配置失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("初始化配置失败: %v", err),
		}
	}
	
	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Printf("配置解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("配置解析失败: %v", err),
		}
	}

	log.Printf("配置解析成功: Title=%s, DefaultCurrency=%s", cfg.Title, cfg.DefaultCurrency)

	// 更新全局配置和文件读取器
	currentConfig = cfg
	fileReader = wasm.NewSimpleFileReader(currentConfig)
	
	// 重要：恢复之前选择的 provider
	if currentProvider != "" {
		fileReader.SetProvider(currentProvider)
		log.Printf("[parseYamlConfig] 恢复 provider: %s", currentProvider)
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("配置更新成功: %s", cfg.Title),
		"config": map[string]interface{}{
			"title":               cfg.Title,
			"defaultCurrency":     cfg.DefaultCurrency,
			"defaultMinusAccount": cfg.DefaultMinusAccount,
			"defaultPlusAccount":  cfg.DefaultPlusAccount,
		},
	}
}

// setProvider 设置当前 Provider
func setProvider(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error":   "需要 Provider 参数",
		}
	}

	provider := args[0].String()
	currentProvider = provider
	fileReader.SetProvider(provider)

	return map[string]interface{}{
		"success":  true,
		"provider": provider,
	}
}

// getCurrentProvider 获取当前 Provider
func getCurrentProvider(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"success":  true,
		"provider": currentProvider,
	}
}

// getSupportedProviders 获取支持的 Provider 列表
func getSupportedProviders(this js.Value, args []js.Value) interface{} {
	providers := []string{
		"alipay", "wechat", "icbc", "ccb", "citic",
		"hsbchk", "htsec", "huobi", "td", "bmo", "mt", "jd", "oklink",
	}

	// 转换为 JavaScript 数组
	jsProviders := js.Global().Get("Array").New(len(providers))
	for i, p := range providers {
		jsProviders.SetIndex(i, js.ValueOf(p))
	}

	return map[string]interface{}{
		"success":   true,
		"providers": jsProviders,
	}
}

// testFunction 测试函数
func testFunction(this js.Value, args []js.Value) interface{} {
	var a, b int = 1, 2
	if len(args) >= 2 {
		if args[0].Type() == js.TypeNumber {
			a = args[0].Int()
		}
		if args[1].Type() == js.TypeNumber {
			b = args[1].Int()
		}
	}

	return map[string]interface{}{
		"success": true,
		"args":    a + b,
		"message": fmt.Sprintf("测试成功: %d + %d = %d", a, b, a+b),
	}
}

// keepAlive 心跳函数
func keepAlive(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"success": true,
	}
}
