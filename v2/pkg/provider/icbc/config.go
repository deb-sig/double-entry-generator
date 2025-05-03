package icbc

// Config is the configuration for ICBC.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer              *string `mapstructure:"peer,omitempty"`
	Item              *string `mapstructure:"item,omitempty"`
	Type              *string `mapstructure:"type,omitempty"`
	TxType            *string `mapstructure:"txType,omitempty"`
	Separator         *string `mapstructure:"sep,omitempty"` // default: ,
	Method            *string `mapstructure:"method,omitempty"`
	Time              *string `mapstructure:"time,omitempty"`
	TimestampRange    *string `mapstructure:"timestamp_range,omitempty"`
	MethodAccount     *string `mapstructure:"methodAccount,omitempty"`
	TargetAccount     *string `mapstructure:"targetAccount,omitempty"`
	CommissionAccount *string `mapstructure:"commissionAccount,omitempty"`
	FullMatch         bool    `mapstructure:"fullMatch,omitempty"` // default: false
	Tag               *string `mapstructure:"tag,omitempty"`
	Ignore            bool    `mapstructure:"ignore,omitempty"` // default: false
}
