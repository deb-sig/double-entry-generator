package wasm

import (
	"fmt"
	"log"
	"strings"

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

// ProcessFile 处理文件（WASM 中传递文件名和原始字节）
func (fr *SimpleFileReader) ProcessFile(fileName string, fileData []byte) interface{} {
	log.Printf("[FileReader] ProcessFile called with provider: %s, file: %s, size: %d bytes", 
		fr.currentProvider, fileName, len(fileData))

	var orders *ir.IR
	var err error

	// 在 WASM 中，provider 需要根据文件类型选择不同的处理方式
	// CSV/TXT: 转换为字符串传递
	// XLS/XLSX: 保持字节数组，使用特殊的 WASM 处理方法
	
	// 根据 provider 调用不同的解析器
	switch fr.currentProvider {
	case "alipay":
		provider := alipay.New()
		// Alipay 只支持 CSV，转换为字符串
		orders, err = provider.Translate(string(fileData))
	case "wechat":
		provider := wechat.New()
		orders, err = provider.Translate(string(fileData))
	case "icbc":
		provider := icbc.New()
		orders, err = provider.Translate(string(fileData))
	case "ccb":
		// CCB 支持 CSV 和 Excel，需要特殊处理
		orders, err = fr.processCCB(fileName, fileData)
	case "citic":
		provider := citic.New()
		orders, err = provider.Translate(string(fileData))
	case "hsbchk":
		provider := hsbchk.New()
		orders, err = provider.Translate(string(fileData))
	case "htsec":
		provider := htsec.New()
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

// processCCB 处理 CCB 文件（支持 CSV 和 Excel）
func (fr *SimpleFileReader) processCCB(fileName string, fileData []byte) (*ir.IR, error) {
	provider := ccb.New()
	
	// 判断文件类型
	lowerName := strings.ToLower(fileName)
	if strings.HasSuffix(lowerName, ".xls") {
		// XLS 文件：使用 xlsReader 从字节流读取
		return fr.processCCBExcel(provider, fileData)
	} else if strings.HasSuffix(lowerName, ".xlsx") {
		// XLSX 文件：暂不支持（需要 excelize.OpenReader）
		return nil, fmt.Errorf("暂不支持 XLSX 格式，请转换为 XLS 或 CSV")
	} else {
		// CSV 文件：直接传递字符串
		return provider.Translate(string(fileData))
	}
}

// processCCBExcel 使用 xlsReader 从字节流读取 XLS 文件
func (fr *SimpleFileReader) processCCBExcel(provider *ccb.CCB, fileData []byte) (*ir.IR, error) {
	log.Printf("[WASM-CCB] 开始解析 XLS 文件，大小: %d bytes", len(fileData))
	
	// 调用 CCB provider 的 TranslateFromExcelBytes 方法
	return provider.TranslateFromExcelBytes(fileData)
}
