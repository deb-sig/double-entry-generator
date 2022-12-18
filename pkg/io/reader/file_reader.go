//go:build !js
// +build !js

package reader

import (
	"io"
	"os"
)

func GetReader(filename string) (io.Reader, error) {
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return csvFile, nil
}
