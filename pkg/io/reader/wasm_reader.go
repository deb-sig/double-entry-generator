//go:build js && wasm
// +build js,wasm

package reader

import (
	"bufio"
	"bytes"
	"io"
)

func GetReader(fileContent string) (io.Reader, error) {
	return bufio.NewReader(bytes.NewBuffer([]byte(fileContent))), nil
}
