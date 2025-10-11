package wasm

import (
	"fmt"
	"log"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler/beancount"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler/ledger"
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
	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/hxsec"
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

// isExcelFile 判断文件是否为 Excel 格式（.xls 或 .xlsx）
func (fr *SimpleFileReader) isExcelFile(fileName string) bool {
	lowerName := strings.ToLower(fileName)
	return strings.HasSuffix(lowerName, ".xls") || strings.HasSuffix(lowerName, ".xlsx")
}

// ProcessFile 处理文件（WASM 中传递文件名和原始字节），默认输出 Beancount 格式
func (fr *SimpleFileReader) ProcessFile(fileName string, fileData []byte) interface{} {
	return fr.ProcessFileWithFormat(fileName, fileData, "beancount")
}

// ProcessFileWithFormat 处理文件并指定输出格式
func (fr *SimpleFileReader) ProcessFileWithFormat(fileName string, fileData []byte, format string) interface{} {
	log.Printf("[FileReader] ProcessFileWithFormat called with provider: %s, file: %s, format: %s, size: %d bytes",
		fr.currentProvider, fileName, format, len(fileData))

	var orders *ir.IR
	var err error

	// 根据 provider 和文件类型选择处理方式
	switch fr.currentProvider {
	case "alipay":
		provider := alipay.New()
		// Alipay 只支持 CSV
		orders, err = provider.Translate(string(fileData))
		
	case "wechat":
		provider := wechat.New()
		provider.IgnoreInvalidTxTypes = true
		if fr.isExcelFile(fileName) {
			orders, err = provider.TranslateFromExcelBytes(fileData)
		} else {
			orders, err = provider.Translate(string(fileData))
		}
		
	case "ccb":
		provider := ccb.New()
		if fr.isExcelFile(fileName) {
			orders, err = provider.TranslateFromExcelBytes(fileData)
		} else {
			orders, err = provider.Translate(string(fileData))
		}
		
	case "citic":
		provider := citic.New()
		// CITIC 只支持 XLS 格式
		orders, err = provider.TranslateFromExcelBytes(fileData)
		
	case "htsec":
		provider := htsec.New()
		if fr.isExcelFile(fileName) {
			orders, err = provider.TranslateFromExcelBytes(fileData)
		} else {
			orders, err = provider.Translate(string(fileData))
		}
		
	case "icbc":
		provider := icbc.New()
		orders, err = provider.Translate(string(fileData))
		
	case "hsbchk":
		provider := hsbchk.New()
		orders, err = provider.Translate(string(fileData))
		
	case "huobi":
		provider := huobi.New()
		orders, err = provider.Translate(string(fileData))
		
	case "td":
		provider := td.New()
		orders, err = provider.Translate(string(fileData))
		
	case "bmo":
		provider := bmo.New()
		orders, err = provider.Translate(string(fileData))
		
	case "mt":
		provider := mt.New()
		orders, err = provider.Translate(string(fileData))
		
	case "jd":
		provider := jd.New()
		orders, err = provider.Translate(string(fileData))
		
	case "hxsec":
		provider := hxsec.New()
		// hxsec 的 .xls 实际上是 TSV 文本文件，不是真正的 Excel
		orders, err = provider.Translate(string(fileData))
		
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

	// 根据格式编译
	var output string
	switch format {
	case "ledger":
		output, err = fr.compileToLedger(orders)
	case "beancount":
		fallthrough
	default:
		output, err = fr.compileToBeancount(orders)
	}
	
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("编译失败: %v", err),
		}
	}

	// 返回结果，字段名统一为 output，但保留 beancount 字段以兼容旧代码
	return map[string]interface{}{
		"success":      true,
		"output":       output,
		"beancount":    output, // 兼容字段
		"ledger":       output, // 兼容字段
		"format":       format,
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

// compileToLedger 编译 IR 为 Ledger
func (fr *SimpleFileReader) compileToLedger(ir *ir.IR) (string, error) {
	// 创建 analyser
	a, err := analyser.New(fr.currentProvider)
	if err != nil {
		return "", fmt.Errorf("创建 analyser 失败: %v", err)
	}

	// 创建 ledger compiler，使用 "memory:" 前缀让 writer 返回内存 writer
	c, err := ledger.New(
		fr.currentProvider,
		consts.CompilerLedger,
		"memory:output", // 使用 "memory:" 前缀触发内存 writer
		false,           // appendMode
		fr.config,
		ir,
		a,
	)
	if err != nil {
		return "", fmt.Errorf("创建 ledger compiler 失败: %v", err)
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
