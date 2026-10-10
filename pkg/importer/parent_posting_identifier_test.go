package importer

import (
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"testing"
)

func TestParentPostingPreservesAccountIdentifiers(t *testing.T) {
	for _, account := range []string{"Assets:Bank:Card123-456", "Assets:Bank:2026", "Expenses:Food:餐饮123-456"} {
		input := account + " 1+2*3 CNY"
		got, err := renderPostingTextStrict(input, Row{}, ir.Order{})
		if err != nil {
			t.Errorf("%q: %v", input, err)
			continue
		}
		if want := account + " 7.00 CNY"; got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}
