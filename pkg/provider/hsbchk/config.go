package hsbchk

type Rule struct {
	Item              *string  `yaml:"item,omitempty"`     // 交易描述
	Merchant          *string  `yaml:"merchant,omitempty"` // 商户名称
	Country           *string  `yaml:"country,omitempty"`  // 国家/地区
	Type              *string  `yaml:"type,omitempty"`     // 收/支
	Time              *string  `yaml:"time,omitempty"`     // 时间 (HH:mm-HH:mm or HH:mm:ss-HH:mm:ss)
	MinPrice          *float64 `yaml:"minPrice,omitempty"` // 最小金额 (包含)
	MaxPrice          *float64 `yaml:"maxPrice,omitempty"` // 最大金额 (包含)
	TimestampRange    *string  `yaml:"timestampRange,omitempty"`
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
