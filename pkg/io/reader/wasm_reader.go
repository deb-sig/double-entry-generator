//go:build js && wasm
// +build js,wasm

package reader

import (
	"bufio"
	"bytes"
	"io"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GetReader(fileContent string) (io.Reader, error) {
	return bufio.NewReader(bytes.NewBuffer([]byte(fileContent))), nil
}

// GetGBKReader 处理 GBK 编码的文件（如支付宝）
func GetGBKReader(fileContent string) (io.Reader, error) {
	// 将 GBK 编码的字节流转换为 UTF-8
	reader := bytes.NewReader([]byte(fileContent))
	gbkReader := transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
	return bufio.NewReader(gbkReader), nil
}
