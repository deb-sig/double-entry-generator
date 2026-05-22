package cgb_credit

import (
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// 测试未配置规则时，广发信用卡仍使用 defaultCashAccount 作为负债账户。
func TestGetAccountsAndTagsWithoutRulesUsesDefaultCashAccount(t *testing.T) {
	analyser := CgbCredit{}
	cfg := &config.Config{
		DefaultMinusAccount: "Assets:FIXME",
		DefaultPlusAccount:  "Expenses:FIXME",
		DefaultCashAccount:  "Liabilities:CGB:CreditCard",
	}

	_, minus, plus, _, _ := analyser.GetAccountsAndTags(&ir.Order{Type: ir.TypeSend}, cfg, "beancount", "cgb_credit")
	if minus != "Liabilities:CGB:CreditCard" || plus != "Expenses:FIXME" {
		t.Fatalf("send accounts = (%s, %s), want (Liabilities:CGB:CreditCard, Expenses:FIXME)", minus, plus)
	}

	_, minus, plus, _, _ = analyser.GetAccountsAndTags(&ir.Order{Type: ir.TypeRecv}, cfg, "beancount", "cgb_credit")
	if minus != "Assets:FIXME" || plus != "Liabilities:CGB:CreditCard" {
		t.Fatalf("recv accounts = (%s, %s), want (Assets:FIXME, Liabilities:CGB:CreditCard)", minus, plus)
	}
}
