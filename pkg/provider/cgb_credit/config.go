package cgb_credit

// Config 是广发信用卡 provider 的规则配置。
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule 描述交易匹配条件和目标账户映射。
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
