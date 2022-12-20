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
	return file, nil
}
