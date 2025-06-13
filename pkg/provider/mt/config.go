package mt

// Config is the configuration for MT.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Item           *string  `mapstructure:"item,omitempty"`
	Type           *string  `mapstructure:"type,omitempty"`
	Method         *string  `mapstructure:"method,omitempty"`
	Status         *string  `mapstructure:"status,omitempty"`
	Separator      *string  `mapstructure:"sep,omitempty"` // default: ,
	Time           *string  `mapstructure:"time,omitempty"`
	TimestampRange *string  `mapstructure:"timestamp_range,omitempty"`
	MethodAccount  *string  `mapstructure:"methodAccount,omitempty"`
	TargetAccount  *string  `mapstructure:"targetAccount,omitempty"`
	FullMatch      bool     `mapstructure:"fullMatch,omitempty"`
	Tags           *string  `mapstructure:"tags,omitempty"`
	Ignore         bool     `mapstructure:"ignore,omitempty"` // default: false
	MinPrice       *float64 `mapstructure:"minPrice,omitempty"`
	MaxPrice       *float64 `mapstructure:"maxPrice,omitempty"`
}
