package htsec

type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

type Rule struct {
	// Peer              *string `mapstructure:"peer,omitempty"`
	Item              *string `mapstructure:"item,omitempty"` // "513050-中概互联"
	Type              *string `mapstructure:"type,omitempty"` // "买"、"卖"
	Time              *string `mapstructure:"time,omitempty"`
	TimestampRange    *string `mapstructure:"timestamp_range,omitempty"`
	Separator         *string `mapstructure:"sep,omitempty"` // default: ,
	CashAccount       *string `mapstructure:"cashAccount,omitempty"`
	PositionAccount   *string `mapstructure:"positionAccount,omitempty"`
	CommissionAccount *string `mapstructure:"commissionAccount,omitempty"`
	PnlAccount        *string `mapstructure:"pnlAccount,omitempty"`
	FullMatch         bool    `mapstructure:"fullMatch,omitempty"`
	Ignore            bool    `mapstructure:"ignore,omitempty"` // default: false
}
