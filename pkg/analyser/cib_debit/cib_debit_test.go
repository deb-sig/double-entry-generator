package cib_debit

import (
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	provider "github.com/deb-sig/double-entry-generator/v2/pkg/provider/cib_debit"
)

func TestGetAccountsAndTagsMatchesPeerBankAndPeerAccount(t *testing.T) {
	targetAccount := "Expenses:Matched"
	peerBank := "示例银行"
	peerAccount := "PEER-MOCK-CNY-CARD"
	cfg := &config.Config{
		DefaultCashAccount:  "Assets:DebitCard:CIB",
		DefaultMinusAccount: "Income:FIXME",
		DefaultPlusAccount:  "Expenses:FIXME",
		CibDebit: &provider.Config{Rules: []provider.Rule{{
			PeerBank:      &peerBank,
			PeerAccount:   &peerAccount,
			TargetAccount: &targetAccount,
		}}},
	}
	order := &ir.Order{
		Type:     ir.TypeSend,
		Metadata: map[string]string{"peerBank": "示例银行", "peerAccount": "PEER-MOCK-CNY-CARD"},
	}

	_, _, plus, _, _ := CibDebit{}.GetAccountsAndTags(order, cfg, "beancount", "cib_debit")
	if plus != targetAccount {
		t.Fatalf("plus account = %q, want %q", plus, targetAccount)
	}
}

func TestGetAccountsAndTagsDoesNotMatchDifferentPeerAccount(t *testing.T) {
	targetAccount := "Expenses:Matched"
	peerBank := "示例银行"
	peerAccount := "PEER-MOCK-CNY-CARD"
	cfg := &config.Config{
		DefaultCashAccount:  "Assets:DebitCard:CIB",
		DefaultMinusAccount: "Income:FIXME",
		DefaultPlusAccount:  "Expenses:FIXME",
		CibDebit: &provider.Config{Rules: []provider.Rule{{
			PeerBank:      &peerBank,
			PeerAccount:   &peerAccount,
			TargetAccount: &targetAccount,
		}}},
	}
	order := &ir.Order{
		Type:     ir.TypeSend,
		Metadata: map[string]string{"peerBank": "示例银行", "peerAccount": "PEER-MOCK-OTHER-CARD"},
	}

	_, _, plus, _, _ := CibDebit{}.GetAccountsAndTags(order, cfg, "beancount", "cib_debit")
	if plus != cfg.DefaultPlusAccount {
		t.Fatalf("plus account = %q, want default %q", plus, cfg.DefaultPlusAccount)
	}
}
