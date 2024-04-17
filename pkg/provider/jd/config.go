package jd

// Config is the configuration for JD.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer           *string `mapstructure:"peer,omitempty"`
	Item           *string `mapstructure:"item,omitempty"`
	Category       *string `mapstructure:"category,omitempty"`
	Type           *string `mapstructure:"type,omitempty"`
	Method         *string `mapstructure:"method,omitempty"`
	OrderStatus    *string `mapstructure:"orderStatus,omitempty"`
	Separator      *string `mapstructure:"sep,omitempty"` // default: ,
	Time           *string `mapstructure:"time,omitempty"`
	TimestampRange *string `mapstructure:"timestamp_range,omitempty"`
	MethodAccount  *string `mapstructure:"methodAccount,omitempty"`
	TargetAccount  *string `mapstructure:"targetAccount,omitempty"`
	PnlAccount     *string `mapstructure:"pnlAccount,omitempty"`
	FullMatch      bool    `mapstructure:"fullMatch,omitempty"`
	Tags           *string `mapstructure:"tags,omitempty"`
	Ignore         bool    `mapstructure:"ignore,omitempty"` // default: false
}
