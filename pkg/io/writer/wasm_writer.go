//go:build js && wasm
// +build js,wasm

package writer

import (
	"fmt"
	"syscall/js"
)

var document js.Value

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

func GetWriter(fileName string) (OutputWriter, error) {
	outputArea := document.Call("getElementById", fileName)
	if !outputArea.Truthy() {
		return nil, fmt.Errorf("can't get `%v` element from document object", fileName)
	}
	// flush the output
	outputArea.Set("textContent", "")
	return (*WasmWriter)(&outputArea), nil
}
