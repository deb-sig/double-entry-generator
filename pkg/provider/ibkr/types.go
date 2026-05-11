package ibkr

type Rule struct {
	Item              *string `mapstructure:"item,omitempty"`
	Type              *string `mapstructure:"type,omitempty"`
	Separator         *string `mapstructure:"sep,omitempty"`
	CashAccount       *string `mapstructure:"cashAccount,omitempty"`
	PositionAccount   *string `mapstructure:"positionAccount,omitempty"`
	CommissionAccount *string `mapstructure:"commissionAccount,omitempty"`
	PnlAccount        *string `mapstructure:"pnlAccount,omitempty"`
	FullMatch         bool    `mapstructure:"fullMatch,omitempty"`
	Ignore            bool    `mapstructure:"ignore,omitempty"`
}

type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}
