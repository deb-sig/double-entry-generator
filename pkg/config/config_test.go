package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestUnmarshalAlipayKeepRefundRecords(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader("alipay:\n  keepRefundRecords: true\n")); err != nil {
		t.Fatalf("read config: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg.Alipay == nil || !cfg.Alipay.KeepRefundRecords {
		t.Fatal("expected alipay.keepRefundRecords to be true")
	}
}
