//go:build js && wasm
// +build js,wasm

package validator

import "fmt"

func TranslateArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("failed to translate: Require the bill content")
	}
	return nil
}
