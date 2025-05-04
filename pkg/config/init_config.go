//go:build !js
// +build !js

package config

import (
	"log"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func InitConfig(cfgFile string) {
	log.SetPrefix("[Init-Config] ")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
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
}
