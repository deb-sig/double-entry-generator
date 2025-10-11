//go:build js && wasm
// +build js,wasm

package config

import (
	"bytes"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func InitConfig(cfgContent string) error {
	if cfgContent == "" {
		return fmt.Errorf("[ERROR] Can't get config from args!")
	}

	// 重置 Viper 实例，避免之前的配置污染
	viper.Reset()
	
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	
	err := viper.ReadConfig(bytes.NewBuffer([]byte(cfgContent)))
	if err != nil {
		log.Printf("[InitConfig] Failed to read config: %v", err)
		return fmt.Errorf("failed to read config: %v", err)
	}
	
	log.Printf("[InitConfig] Config loaded successfully")
	return nil
}
