package bocomcredit

// Config is the configuration for BocomCredit.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule describes how to match bills and map to accounts.
type Rule struct {
	Peer          *string `mapstructure:"peer,omitempty"`
	Item          *string `mapstructure:"item,omitempty"`
	Type          *string `mapstructure:"type,omitempty"`
	Separator     *string `mapstructure:"sep,omitempty"`
	MethodAccount *string `mapstructure:"methodAccount,omitempty"`
	TargetAccount *string `mapstructure:"targetAccount,omitempty"`
	FullMatch     bool    `mapstructure:"fullMatch,omitempty"`
	Tag           *string `mapstructure:"tag,omitempty"`
	Ignore        bool    `mapstructure:"ignore,omitempty"`
}
