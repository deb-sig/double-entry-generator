//go:build !js
// +build !js

package writer

import (
	"fmt"
	"log"
	"os"
)

func GetWriter(outputFile string) (OutputWriter, error) {
	log.Printf("Writing to %s", outputFile)
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("create output file  %s error: %v", outputFile, err)
	}
	// 设置文件编码为UTF-8（在Windows上尤为重要）
	// Go的os.Create默认创建UTF-8文件，但在某些Windows环境下可能需要明确指定
	return file, nil
}
