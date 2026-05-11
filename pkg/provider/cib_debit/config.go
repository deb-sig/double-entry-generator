package cib_debit

// Config is the configuration for CIB debit.
type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

// Rule is the type for match rules.
type Rule struct {
	Peer          *string  `mapstructure:"peer,omitempty"`
	PeerBank      *string  `mapstructure:"peerBank,omitempty"`
	PeerAccount   *string  `mapstructure:"peerAccount,omitempty"`
	Item          *string  `mapstructure:"item,omitempty"`
	Type          *string  `mapstructure:"type,omitempty"`
	TxType        *string  `mapstructure:"txType,omitempty"`
	MinPrice      *float64 `mapstructure:"minPrice,omitempty"`
	MaxPrice      *float64 `mapstructure:"maxPrice,omitempty"`
	Separator     *string  `mapstructure:"sep,omitempty"`
	MethodAccount *string  `mapstructure:"methodAccount,omitempty"`
	TargetAccount *string  `mapstructure:"targetAccount,omitempty"`
	FullMatch     bool     `mapstructure:"fullMatch,omitempty"`
	Ignore        bool     `mapstructure:"ignore,omitempty"`
	Tag           *string  `mapstructure:"tag,omitempty"`
}
