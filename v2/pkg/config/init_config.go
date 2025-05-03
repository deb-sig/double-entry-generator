//go:build !js
// +build !js

package config

import (
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func InitConfig(cfgFile string) {
	log.SetPrefix("[Init-Config] ")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

		// 先读取文件内容，并尝试解析YAML格式
		content, err := os.ReadFile(cfgFile)
		if err != nil {
			log.Fatalf("Cannot read config file: %v", err)
		}

		// 尝试解析YAML，看是否格式正确
		var yamlCheck map[string]interface{}
		if err := yaml.Unmarshal(content, &yamlCheck); err != nil {
			log.Fatalf("Config YAML format error: %v", err)
		}
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("Failed to find home directory: %v", err)
		}

		// Search config in home directory with name ".double-entry-generator" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".double-entry-generator")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read config file error: %v", err)
	}

	log.Printf("Use config file: %s", viper.ConfigFileUsed())
}
