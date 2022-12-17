package pkg

import (
	"syscall/js"
)

var document js.Value

func init() {
	document = js.Global().Get("document")
}

type WasmWriter js.Value

// Write implements io.Writer.
func (d WasmWriter) Write(p []byte) (n int, err error) {
	// outputArea := document.Call("getElementById", "output")
	// node := document.Call("createElement", "p")
	// text := strings.Replace(string(p), "\n", "<br>", -1)
	// node.Set("textContent", text)
	// js.Value(d).Call("appendChild", node)

	outputArea := js.Value(d)
	current := outputArea.Get("textContent").String()
	new := current + string(p)
	outputArea.Set("textContent", new)
	return len(p), nil
}
