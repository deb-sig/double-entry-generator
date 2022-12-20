//go:build js && wasm
// +build js,wasm

package config

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"
)

func InitConfig(cfgContent string) {
	if cfgContent != "" {
		viper.AutomaticEnv()
		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer([]byte(cfgContent)))
	} else {
		fmt.Errorf("[ERROR] Can't get config from args!")
	}
	return
}
