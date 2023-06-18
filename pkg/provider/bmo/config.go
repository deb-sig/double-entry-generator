package bmo

// Config is the configuration for Bmo.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer          *string `mapstructure:"peer,omitempty"` // 交易对手
	Item          *string `mapstructure:"item,omitempty"` // 商品描述
	Type          *string `mapstructure:"type,omitempty"` // 类型
	Separator     *string `mapstructure:"sep,omitempty"`  // default: ,
	MethodAccount *string `mapstructure:"methodAccount,omitempty"`
	TargetAccount *string `mapstructure:"targetAccount,omitempty"`
	FullMatch     bool    `mapstructure:"fullMatch,omitempty"` // default: false
	Tag           *string `mapstructure:"tag,omitempty"`
	Ignore        bool    `mapstructure:"ignore,omitempty"` // default: false
}
