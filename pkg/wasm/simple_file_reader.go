package wasm

import (
	"fmt"
	"log"

	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler/beancount"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/io/writer"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/alipay"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/bmo"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/ccb"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/citic"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/hsbchk"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/htsec"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/huobi"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/icbc"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/jd"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/mt"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/td"
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/wechat"
)

// SimpleFileReader 简化的文件读取器
type SimpleFileReader struct {
	config          *config.Config
	currentProvider string
}

// NewSimpleFileReader 创建简化的文件读取器
func NewSimpleFileReader(cfg *config.Config) *SimpleFileReader {
	return &SimpleFileReader{
		config:          cfg,
		currentProvider: "alipay",
	}
}

// SetProvider 设置当前 Provider
func (fr *SimpleFileReader) SetProvider(provider string) {
	fr.currentProvider = provider
}

// ProcessFile 处理文件（WASM 中 filename 实际上是文件内容）
func (fr *SimpleFileReader) ProcessFile(fileContent string) interface{} {

	var orders *ir.IR
	var err error

	// 在 WASM 中，provider 的 Translate 接收的是文件内容而不是文件路径
	// 根据 provider 调用不同的解析器
	switch fr.currentProvider {
	case "alipay":
		provider := alipay.New()
		orders, err = provider.Translate(fileContent)
	case "wechat":
		provider := wechat.New()
		orders, err = provider.Translate(fileContent)
	case "icbc":
		provider := icbc.New()
		orders, err = provider.Translate(fileContent)
	case "ccb":
		provider := ccb.New()
		orders, err = provider.Translate(fileContent)
	case "citic":
		provider := citic.New()
		orders, err = provider.Translate(fileContent)
	case "hsbchk":
		provider := hsbchk.New()
		orders, err = provider.Translate(fileContent)
	case "htsec":
		provider := htsec.New()
		orders, err = provider.Translate(fileContent)
	case "huobi":
		provider := huobi.New()
		orders, err = provider.Translate(fileContent)
	case "td":
		provider := td.New()
		orders, err = provider.Translate(fileContent)
	case "bmo":
		provider := bmo.New()
		orders, err = provider.Translate(fileContent)
	case "mt":
		provider := mt.New()
		orders, err = provider.Translate(fileContent)
	case "jd":
		provider := jd.New()
		orders, err = provider.Translate(fileContent)
	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("不支持的 Provider: %s", fr.currentProvider),
		}
	}

	if err != nil {
		log.Printf("[FileReader] 解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("解析失败: %v", err),
		}
	}

	// 编译为 Beancount
	beancount, err := fr.compileToBeancount(orders)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("编译失败: %v", err),
		}
	}

	return map[string]interface{}{
		"success":      true,
		"beancount":    beancount,
		"transactions": len(orders.Orders),
		"provider":     fr.currentProvider,
	}
}

// compileToBeancount 编译 IR 为 Beancount
func (fr *SimpleFileReader) compileToBeancount(ir *ir.IR) (string, error) {
	// 创建 analyser
	a, err := analyser.New(fr.currentProvider)
	if err != nil {
		return "", fmt.Errorf("创建 analyser 失败: %v", err)
	}

	// 创建 beancount compiler，使用 "memory:" 前缀让 writer 返回内存 writer
	c, err := beancount.New(
		fr.currentProvider,
		consts.CompilerBeanCount,
		"memory:output", // 使用 "memory:" 前缀触发内存 writer
		false,           // appendMode
		fr.config,
		ir,
		a,
	)
	if err != nil {
		return "", fmt.Errorf("创建 beancount compiler 失败: %v", err)
	}

	// 执行编译
	err = c.Compile()
	if err != nil {
		return "", fmt.Errorf("编译失败: %v", err)
	}

	// 从全局变量获取最后一个内存 writer
	memWriter := writer.GetLastMemoryWriter()
	if memWriter == nil {
		return "", fmt.Errorf("未找到内存 writer")
	}

	return memWriter.String(), nil
}
