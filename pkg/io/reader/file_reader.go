//go:build !js
// +build !js

package reader

import (
	"io"
	"os"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GetReader(filename string) (io.Reader, error) {
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return csvFile, nil
}

func GetGBKReader(filename string) (io.Reader, error) {
	csvFile, err := GetReader(filename)
	if err != nil {
		return nil, err
	}

	gbkReader := transform.NewReader(csvFile, simplifiedchinese.GBK.NewDecoder())
	return gbkReader, nil
}
