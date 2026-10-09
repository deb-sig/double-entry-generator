package importer

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func loadProfileYAML(t *testing.T, src string) *Profile {
	t.Helper()
	var p Profile
	if err := yaml.Unmarshal([]byte(src), &p); err != nil {
		t.Fatalf("profile: %v", err)
	}
	normalizeTemplate(&p.Template, p.Defaults)
	return &p
}

// A `reader:` block makes json, xml and text bills first-class: the named
// columns feed the same rule engine as csv headers.
func TestReaderBlockJSONFeedsRules(t *testing.T) {
	profile := loadProfileYAML(t, `
schema: https://deg.dev/template-profile/v2
id: chain
reader:
  format: json
  records: "//result/*"
  columns:
    hash: hash
    time: timeStamp
    value: value
template:
  sourceHeaders: [hash, time, value]
  defaultCurrency: ETH
templateRules:
  - id: base
    actions:
      date: <time>
      payee: chain
      narration: <hash>
      amount: <value>.number
      from: Assets:Wallet
      to: Expenses:Gas
`)
	data := []byte(`{"result":[{"hash":"0xa","timeStamp":"2026-01-02","value":"0.5"}]}`)
	rows, err := ParseBytes(profile, "txlist.json", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Raw["hash"] != "0xa" || rows[0].Raw["value"] != "0.5" {
		t.Fatalf("rows = %+v", rows)
	}
	if ReaderConfig(profile).Format != "json" {
		t.Fatalf("format = %q", ReaderConfig(profile).Format)
	}
}

func TestReaderBlockFallsBackToTemplateFields(t *testing.T) {
	profile := loadProfileYAML(t, `
template:
  fileFormat: csv
  encoding: gb18030
  delimiter: ";"
  stripTabs: true
  headerLocate: true
`)
	cfg := ReaderConfig(profile)
	if cfg.Format != "csv" || cfg.Encoding != "gb18030" || cfg.Delimiter != ";" || !cfg.StripTabs || !cfg.KeepBlankLines {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestReaderBlockRejectsMismatchedBill(t *testing.T) {
	profile := loadProfileYAML(t, `
id: stmt
reader:
  format: text
  record: '(?P<date>\d+)'
`)
	_, err := ParseBytes(profile, "stmt.xlsx", nil)
	if err == nil || !strings.Contains(err.Error(), `fileFormat="text"`) {
		t.Fatalf("err = %v", err)
	}
}

func yamlUnmarshal(src string, out interface{}) error {
	return yaml.Unmarshal([]byte(src), out)
}
