//go:build !js
// +build !js

package validator

import (
	"fmt"
	"os"
)

func TranslateArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("failed to translate: Require the bill file")
	} else if len(args) > 1 {
		// TODO(gaocegege): support it.
		return fmt.Errorf("failed to translate: Do not support multi-file now")
	}

	_, err := os.Stat(args[0])
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to translate: %v", err)
}
