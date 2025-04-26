package hsbchk

type Rule struct {
	// Debit, Credit
	Peer              *string  `yaml:"peer,omitempty"`           // Description
	Item              *string  `yaml:"item,omitempty"`           // ,
	Type              *string  `yaml:"type,omitempty"`           // 收/支
	Time              *string  `yaml:"time,omitempty"`           // 时间区间
	MinPrice          *float64 `yaml:"minPrice,omitempty"`       // 最小金额 (包含)
	MaxPrice          *float64 `yaml:"maxPrice,omitempty"`       // 最大金额 (包含)
	TimestampRange    *string  `yaml:"timestampRange,omitempty"` // WIP
	TargetAccount     *string  `yaml:"targetAccount,omitempty"`
	MethodAccount     *string  `yaml:"methodAccount,omitempty"`
	CommissionAccount *string  `yaml:"commissionAccount,omitempty"`
	Separator         *string  `yaml:"sep,omitempty"`
	FullMatch         bool     `yaml:"fullMatch,omitempty"`
	Tag               *string  `yaml:"tag,omitempty"`
	Ignore            bool     `yaml:"ignore,omitempty"`
}

// Config is hsbchk config from yaml.
type Config struct {
	Rules []Rule `yaml:"rules,omitempty"`
}
