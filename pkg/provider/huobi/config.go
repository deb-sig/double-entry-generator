package huobi

type Config struct {
	Rules []Rule `yaml:"rules,omitempty"`
}

type Rule struct {
	// Peer              *string `yaml:"peer,omitempty"`
	Item              *string `yaml:"item,omitempty"`   // "BTC/USDT"
	Type              *string `yaml:"type,omitempty"`   // "币币交易"
	TxType            *string `yaml:"txType,omitempty"` // "买入"、"卖出"
	Seperator         *string `yaml:"sep,omitempty"`    // default: ,
	CashAccount       *string `yaml:"cashAccount,omitempty"`
	PositionAccount   *string `yaml:"positionAccount,omitempty"`
	CommissionAccount *string `yaml:"commissionAccount,omitempty"`
	PnlAccount        *string `yaml:"pnlAccount,omitempty"`
}
