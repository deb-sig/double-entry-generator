package spdb_debit

// Config is the configuration for SPDB debit.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer          *string `mapstructure:"peer,omitempty"`
	Item          *string `mapstructure:"item,omitempty"`
	Category      *string `mapstructure:"category,omitempty"`
	Separator     *string `mapstructure:"sep,omitempty"` // default: ,
	TargetAccount *string `mapstructure:"targetAccount,omitempty"`
	FullMatch     bool    `mapstructure:"fullMatch,omitempty"`
	Ignore        bool    `mapstructure:"ignore,omitempty"` // default: false
}
