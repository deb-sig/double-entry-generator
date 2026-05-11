//go:build !js
// +build !js

package validator

import (
	"fmt"
	"os"
)

func TranslateArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Failed to translate: Require the bill file")
	}

	for _, arg := range args {
		if _, err := os.Stat(arg); err != nil {
			return fmt.Errorf("Failed to translate: %v", err)
		}
	}
	return nil
}
