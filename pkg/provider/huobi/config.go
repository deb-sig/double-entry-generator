package huobi

type Config struct {
	Rules []Rule `mapstructure:"rules,omitempty"`
}

type Rule struct {
	// Peer              *string `mapstructure:"peer,omitempty"`
	Item              *string `mapstructure:"item,omitempty"`   // "BTC/USDT"
	Type              *string `mapstructure:"type,omitempty"`   // "币币交易"
	TxType            *string `mapstructure:"txType,omitempty"` // "买入"、"卖出"
	Seperator         *string `mapstructure:"sep,omitempty"`    // default: ,
	CashAccount       *string `mapstructure:"cashAccount,omitempty"`
	PositionAccount   *string `mapstructure:"positionAccount,omitempty"`
	CommissionAccount *string `mapstructure:"commissionAccount,omitempty"`
	PnlAccount        *string `mapstructure:"pnlAccount,omitempty"`
}
