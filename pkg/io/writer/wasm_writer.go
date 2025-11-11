//go:build js && wasm
// +build js,wasm

package writer

import (
	"bytes"
	"fmt"
	"syscall/js"
)

var (
	document         js.Value
	lastMemoryWriter *MemoryWriter
)

func init() {
	document = js.Global().Get("document")
}

type WasmWriter js.Value

// Write implements io.Writer.
func (d WasmWriter) Write(p []byte) (n int, err error) {
	outputArea := js.Value(d)
	current := outputArea.Get("textContent").String()
	newContent := current + string(p)
	outputArea.Set("textContent", newContent)
	return len(p), nil
}

// Close implements io.Closer
func (d WasmWriter) Close() error {
	return nil
}

// MemoryWriter 是内存 writer
type MemoryWriter struct {
	buf bytes.Buffer
}

func (m *MemoryWriter) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *MemoryWriter) Close() error {
	return nil
}

func (m *MemoryWriter) String() string {
	return m.buf.String()
}

func GetWriter(fileName string) (OutputWriter, error) {
	// 如果文件名以 "memory:" 开头，返回内存 writer
	if len(fileName) > 7 && fileName[:7] == "memory:" {
		lastMemoryWriter = &MemoryWriter{}
		return lastMemoryWriter, nil
	}
	
	outputArea := document.Call("getElementById", fileName)
	if !outputArea.Truthy() {
		return nil, fmt.Errorf("can't get `%v` element from document object", fileName)
	}
	// flush the output
	outputArea.Set("textContent", "")
	return (*WasmWriter)(&outputArea), nil
}

// GetLastMemoryWriter 获取最后一个创建的内存 writer
func GetLastMemoryWriter() *MemoryWriter {
	return lastMemoryWriter
}
