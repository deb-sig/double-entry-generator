package config

import "github.com/gaocegege/double-entry-generator/pkg/provider/alipay"

// Config is the global configuration.
type Config struct {
	Title               string         `yaml:"title,omitempty"`
	DefaultMinusAccount string         `yaml:"defaultMinusAccount,omitempty"`
	DefaultPlusAccount  string         `yaml:"defaultPlusAccount,omitempty"`
	DefaultCurrency     string         `yaml:"defaultCurrency,omitempty"`
	Alipay              *alipay.Config `yaml:"alipay,omitempty"`
}
