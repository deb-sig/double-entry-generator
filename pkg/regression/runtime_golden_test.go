package regression_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"gopkg.in/yaml.v3"
)

type ruleFile struct {
	TemplateRules         []importer.Rule `yaml:"templateRules"`
	TemplateRuleOverrides []importer.Rule `yaml:"templateRuleOverrides"`
	PersonalRules         []importer.Rule `yaml:"personalRules"`
	Options               ruleOptions     `yaml:"options"`
}

type ruleOptions struct {
	Title             string `yaml:"title"`
	OperatingCurrency string `yaml:"operatingCurrency"`
}

func TestRuntimeTemplatesMatchLegacyGoldens(t *testing.T) {
	root := filepath.Join("..", "..")
	registryPath := writeLocalRegistry(t, root)
	t.Setenv(importer.RegistryURLEnv, registryPath)

	tests := []struct {
		name     string
		template string
		bill     string
		expected string
	}{
		{
			name:     "alipay latest",
			template: "alipay",
			bill:     filepath.Join(root, "example", "alipay", "example-alipay-records.csv"),
			expected: filepath.Join(root, "example", "alipay", "example-alipay-output.beancount"),
		},
		{
			name:     "wechat latest",
			template: "wechat",
			bill:     filepath.Join(root, "example", "wechat", "example-wechat-records.xlsx"),
			expected: filepath.Join(root, "example", "wechat", "example-wechat-output.beancount"),
		},
		{
			name:     "wechat 2026-04-28",
			template: "wechat@2026-04-28",
			bill:     filepath.Join(root, "example", "wechat", "example-wechat-records.xlsx"),
			expected: filepath.Join(root, "example", "wechat", "example-wechat-output.beancount"),
		},
		{
			name:     "oklink latest",
			template: "oklink",
			bill:     filepath.Join(root, "example", "oklink", "example-oklink-token-transfer.csv"),
			expected: filepath.Join(root, "example", "oklink", "example-oklink-output.beancount"),
		},
		{
			name:     "td latest",
			template: "td",
			bill:     filepath.Join(root, "example", "td", "example-td-records.csv"),
			expected: filepath.Join(root, "example", "td", "example-td-output.beancount"),
		},
		{
			name:     "td 2026-05-27",
			template: "td@2026-05-27",
			bill:     filepath.Join(root, "example", "td", "example-td-records.csv"),
			expected: filepath.Join(root, "example", "td", "example-td-output.beancount"),
		},
		{
			name:     "bmo-debit latest",
			template: "bmo-debit",
			bill:     filepath.Join(root, "example", "bmo", "debit", "example-bmo-records.csv"),
			expected: filepath.Join(root, "example", "bmo", "debit", "example-bmo-debit-output.beancount"),
		},
		{
			name:     "bmo-debit 2026-05-28",
			template: "bmo-debit@2026-05-28",
			bill:     filepath.Join(root, "example", "bmo", "debit", "example-bmo-records.csv"),
			expected: filepath.Join(root, "example", "bmo", "debit", "example-bmo-debit-output.beancount"),
		},
		{
			name:     "bmo-credit latest",
			template: "bmo-credit",
			bill:     filepath.Join(root, "example", "bmo", "credit", "example-bmo-records.csv"),
			expected: filepath.Join(root, "example", "bmo", "credit", "example-bmo-credit-output.beancount"),
		},
		{
			name:     "bmo-credit 2026-05-28",
			template: "bmo-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "bmo", "credit", "example-bmo-records.csv"),
			expected: filepath.Join(root, "example", "bmo", "credit", "example-bmo-credit-output.beancount"),
		},
		{
			name:     "cmb-credit latest",
			template: "cmb-credit",
			bill:     filepath.Join(root, "example", "cmb", "credit", "example-cmb-records.csv"),
			expected: filepath.Join(root, "example", "cmb", "credit", "example-cmb-credit-output.beancount"),
		},
		{
			name:     "cmb-credit 2026-05-28",
			template: "cmb-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "cmb", "credit", "example-cmb-records.csv"),
			expected: filepath.Join(root, "example", "cmb", "credit", "example-cmb-credit-output.beancount"),
		},
		{
			name:     "cmb-debit latest",
			template: "cmb-debit",
			bill:     filepath.Join(root, "example", "cmb", "debit", "example-cmb-records.csv"),
			expected: filepath.Join(root, "example", "cmb", "debit", "example-cmb-debit-output.beancount"),
		},
		{
			name:     "cmb-debit 2026-05-28",
			template: "cmb-debit@2026-05-28",
			bill:     filepath.Join(root, "example", "cmb", "debit", "example-cmb-records.csv"),
			expected: filepath.Join(root, "example", "cmb", "debit", "example-cmb-debit-output.beancount"),
		},
		{
			name:     "bocom_credit latest",
			template: "bocom_credit",
			bill:     filepath.Join(root, "example", "bocom_credit", "example-bocom_credit-records.csv"),
			expected: filepath.Join(root, "example", "bocom_credit", "example-bocom_credit-output.beancount"),
		},
		{
			name:     "bocom_credit 2026-05-28",
			template: "bocom_credit@2026-05-28",
			bill:     filepath.Join(root, "example", "bocom_credit", "example-bocom_credit-records.csv"),
			expected: filepath.Join(root, "example", "bocom_credit", "example-bocom_credit-output.beancount"),
		},
		{
			name:     "bocom_debit latest",
			template: "bocom_debit",
			bill:     filepath.Join(root, "example", "bocom_debit", "example-bocom_debit-records.csv"),
			expected: filepath.Join(root, "example", "bocom_debit", "example-bocom_debit-output.beancount"),
		},
		{
			name:     "bocom_debit 2026-05-28",
			template: "bocom_debit@2026-05-28",
			bill:     filepath.Join(root, "example", "bocom_debit", "example-bocom_debit-records.csv"),
			expected: filepath.Join(root, "example", "bocom_debit", "example-bocom_debit-output.beancount"),
		},
		{
			name:     "abc_debit latest",
			template: "abc_debit",
			bill:     filepath.Join(root, "example", "abc_debit", "example-abc_debit-records.csv"),
			expected: filepath.Join(root, "example", "abc_debit", "example-abc_debit-output.beancount"),
		},
		{
			name:     "abc_debit 2026-05-28",
			template: "abc_debit@2026-05-28",
			bill:     filepath.Join(root, "example", "abc_debit", "example-abc_debit-records.csv"),
			expected: filepath.Join(root, "example", "abc_debit", "example-abc_debit-output.beancount"),
		},
		{
			name:     "citic-credit latest",
			template: "citic-credit",
			bill:     filepath.Join(root, "example", "citic", "credit", "example-citic-records.xls"),
			expected: filepath.Join(root, "example", "citic", "credit", "example-citic-credit-output.beancount"),
		},
		{
			name:     "citic-credit 2026-05-28",
			template: "citic-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "citic", "credit", "example-citic-records.xls"),
			expected: filepath.Join(root, "example", "citic", "credit", "example-citic-credit-output.beancount"),
		},
		{
			name:     "hsbchk-debit latest",
			template: "hsbchk-debit",
			bill:     filepath.Join(root, "example", "hsbchk", "debit", "example-hsbchk-debit-records.csv"),
			expected: filepath.Join(root, "example", "hsbchk", "debit", "example-hsbchk-debit-output.beancount"),
		},
		{
			name:     "hsbchk-debit 2026-05-28",
			template: "hsbchk-debit@2026-05-28",
			bill:     filepath.Join(root, "example", "hsbchk", "debit", "example-hsbchk-debit-records.csv"),
			expected: filepath.Join(root, "example", "hsbchk", "debit", "example-hsbchk-debit-output.beancount"),
		},
		{
			name:     "hsbchk-credit latest",
			template: "hsbchk-credit",
			bill:     filepath.Join(root, "example", "hsbchk", "credit", "example-hsbchk-credit-records.csv"),
			expected: filepath.Join(root, "example", "hsbchk", "credit", "example-hsbchk-credit-output.beancount"),
		},
		{
			name:     "hsbchk-credit 2026-05-28",
			template: "hsbchk-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "hsbchk", "credit", "example-hsbchk-credit-records.csv"),
			expected: filepath.Join(root, "example", "hsbchk", "credit", "example-hsbchk-credit-output.beancount"),
		},
		{
			name:     "icbc-credit latest",
			template: "icbc-credit",
			bill:     filepath.Join(root, "example", "icbc", "credit", "example-icbc-credit-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "credit", "example-icbc-credit-output.beancount"),
		},
		{
			name:     "icbc-credit 2026-05-28",
			template: "icbc-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "icbc", "credit", "example-icbc-credit-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "credit", "example-icbc-credit-output.beancount"),
		},
		{
			name:     "icbc-debit-v1 latest",
			template: "icbc-debit-v1",
			bill:     filepath.Join(root, "example", "icbc", "debit-v1", "example-icbc-debit-v1-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "debit-v1", "example-icbc-debit-v1-output.beancount"),
		},
		{
			name:     "icbc-debit-v1 2026-05-28",
			template: "icbc-debit-v1@2026-05-28",
			bill:     filepath.Join(root, "example", "icbc", "debit-v1", "example-icbc-debit-v1-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "debit-v1", "example-icbc-debit-v1-output.beancount"),
		},
		{
			name:     "icbc-debit-v2 latest",
			template: "icbc-debit-v2",
			bill:     filepath.Join(root, "example", "icbc", "debit-v2", "example-icbc-debit-v2-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "debit-v2", "example-icbc-debit-v2-output.beancount"),
		},
		{
			name:     "icbc-debit-v2 2026-05-28",
			template: "icbc-debit-v2@2026-05-28",
			bill:     filepath.Join(root, "example", "icbc", "debit-v2", "example-icbc-debit-v2-records.csv"),
			expected: filepath.Join(root, "example", "icbc", "debit-v2", "example-icbc-debit-v2-output.beancount"),
		},
		{
			name:     "boc-credit latest",
			template: "boc-credit",
			bill:     filepath.Join(root, "example", "boc", "credit", "example-boc-records.csv"),
			expected: filepath.Join(root, "example", "boc", "credit", "example-boc-credit-output.beancount"),
		},
		{
			name:     "boc-credit 2026-05-28",
			template: "boc-credit@2026-05-28",
			bill:     filepath.Join(root, "example", "boc", "credit", "example-boc-records.csv"),
			expected: filepath.Join(root, "example", "boc", "credit", "example-boc-credit-output.beancount"),
		},
		{
			name:     "boc-debit latest",
			template: "boc-debit",
			bill:     filepath.Join(root, "example", "boc", "debit", "example-boc-records.csv"),
			expected: filepath.Join(root, "example", "boc", "debit", "example-boc-debit-output.beancount"),
		},
		{
			name:     "boc-debit 2026-05-28",
			template: "boc-debit@2026-05-28",
			bill:     filepath.Join(root, "example", "boc", "debit", "example-boc-records.csv"),
			expected: filepath.Join(root, "example", "boc", "debit", "example-boc-debit-output.beancount"),
		},
		{
			name:     "jd latest",
			template: "jd",
			bill:     filepath.Join(root, "example", "jd", "example-jd-records.csv"),
			expected: filepath.Join(root, "example", "jd", "example-jd-output.beancount"),
		},
		{
			name:     "jd 2026-05-28",
			template: "jd@2026-05-28",
			bill:     filepath.Join(root, "example", "jd", "example-jd-records.csv"),
			expected: filepath.Join(root, "example", "jd", "example-jd-output.beancount"),
		},
		{
			name:     "mt latest",
			template: "mt",
			bill:     filepath.Join(root, "example", "mt", "example-mt-records.csv"),
			expected: filepath.Join(root, "example", "mt", "example-mt-output.beancount"),
		},
		{
			name:     "mt 2026-05-28",
			template: "mt@2026-05-28",
			bill:     filepath.Join(root, "example", "mt", "example-mt-records.csv"),
			expected: filepath.Join(root, "example", "mt", "example-mt-output.beancount"),
		},
		{
			name:     "spdb_debit latest",
			template: "spdb_debit",
			bill:     filepath.Join(root, "example", "spdb_debit", "example-spdb_debit-records.xls"),
			expected: filepath.Join(root, "example", "spdb_debit", "example-spdb_debit-output.beancount"),
		},
		{
			name:     "spdb_debit 2026-05-28",
			template: "spdb_debit@2026-05-28",
			bill:     filepath.Join(root, "example", "spdb_debit", "example-spdb_debit-records.xls"),
			expected: filepath.Join(root, "example", "spdb_debit", "example-spdb_debit-output.beancount"),
		},
		{
			name:     "ccb latest",
			template: "ccb",
			bill:     filepath.Join(root, "example", "ccb", "交易明细_xxxx_2025xxxx_2025xxxx.xls"),
			expected: filepath.Join(root, "example", "ccb", "example-ccb-output.beancount"),
		},
		{
			name:     "ccb 2026-05-28",
			template: "ccb@2026-05-28",
			bill:     filepath.Join(root, "example", "ccb", "交易明细_xxxx_2025xxxx_2025xxxx.xls"),
			expected: filepath.Join(root, "example", "ccb", "example-ccb-output.beancount"),
		},
		{
			name:     "huobi latest",
			template: "huobi",
			bill:     filepath.Join(root, "example", "huobi", "example-huobi-records.csv"),
			expected: filepath.Join(root, "example", "huobi", "example-huobi-output.beancount"),
		},
		{
			name:     "huobi 2026-05-28",
			template: "huobi@2026-05-28",
			bill:     filepath.Join(root, "example", "huobi", "example-huobi-records.csv"),
			expected: filepath.Join(root, "example", "huobi", "example-huobi-output.beancount"),
		},
		{
			name:     "htsec latest",
			template: "htsec",
			bill:     filepath.Join(root, "example", "htsec", "example-htsec-records.xlsx"),
			expected: filepath.Join(root, "example", "htsec", "example-htsec-output.beancount"),
		},
		{
			name:     "htsec 2026-05-28",
			template: "htsec@2026-05-28",
			bill:     filepath.Join(root, "example", "htsec", "example-htsec-records.xlsx"),
			expected: filepath.Join(root, "example", "htsec", "example-htsec-output.beancount"),
		},
		{
			name:     "hxsec latest",
			template: "hxsec",
			bill:     filepath.Join(root, "example", "hxsec", "example-hxsec-records.xls"),
			expected: filepath.Join(root, "example", "hxsec", "example-hxsec-output.beancount"),
		},
		{
			name:     "hxsec 2026-05-28",
			template: "hxsec@2026-05-28",
			bill:     filepath.Join(root, "example", "hxsec", "example-hxsec-records.xls"),
			expected: filepath.Join(root, "example", "hxsec", "example-hxsec-output.beancount"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compileRegistryImport(t, tt.template, tt.bill)
			want, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				diffDir := filepath.Join(os.TempDir(), "deg-runtime-golden-diff")
				if err := os.MkdirAll(diffDir, 0o755); err != nil {
					t.Fatal(err)
				}
				gotPath := filepath.Join(diffDir, tt.name+"-got.beancount")
				wantPath := filepath.Join(diffDir, tt.name+"-want.beancount")
				if err := os.WriteFile(gotPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(wantPath, want, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Fatalf("registry import output differs from legacy golden\ntemplate: %s\nbill: %s\nexpected: %s\ngot: %s\nwant: %s", tt.template, tt.bill, tt.expected, gotPath, wantPath)
			}
		})
	}
}

func compileRegistryImport(t *testing.T, templateRef, billPath string) []byte {
	t.Helper()
	profile, err := importer.LoadProfileRef(templateRef)
	if err != nil {
		t.Fatal(err)
	}
	rulesRef, err := importer.StarterRulesURLFromRegistry(templateRef)
	if err != nil {
		t.Fatal(err)
	}
	rules := loadRules(t, rulesRef)
	profile.TemplateRules = append(profile.TemplateRules, rules.TemplateRules...)
	if rules.Options.Title != "" {
		profile.Name = rules.Options.Title
	}
	if rules.Options.OperatingCurrency != "" {
		profile.Template.DefaultCurrency = rules.Options.OperatingCurrency
	}
	profile.TemplateRuleOverrides = append(profile.TemplateRuleOverrides, rules.TemplateRuleOverrides...)
	profile.PersonalRules = append(profile.PersonalRules, rules.PersonalRules...)

	ir, err := importer.ImportFile(profile, billPath)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "out.beancount")
	cfg := &config.Config{
		Title:               firstNonEmpty(profile.Name, profile.ID, "DEG Import"),
		DefaultMinusAccount: profile.Template.DefaultMinus,
		DefaultPlusAccount:  profile.Template.DefaultPlus,
		DefaultCurrency:     firstNonEmpty(profile.Template.DefaultCurrency, "CNY"),
	}
	cpl, err := compiler.New(importer.DefaultProviderName, consts.CompilerBeanCount, output, false, cfg, ir)
	if err != nil {
		t.Fatal(err)
	}
	if err := cpl.Compile(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func writeLocalRegistry(t *testing.T, root string) string {
	t.Helper()
	dir := t.TempDir()
	stage := func(provider, version, templateSource, rulesSource string) {
		for _, pin := range []string{"latest", version} {
			pinDir := filepath.Join(dir, provider, pin)
			if err := os.MkdirAll(pinDir, 0o755); err != nil {
				t.Fatal(err)
			}
			copyFile(t, filepath.Join(root, templateSource), filepath.Join(pinDir, "template.yaml"))
			copyFile(t, filepath.Join(root, rulesSource), filepath.Join(pinDir, "rules.yaml"))
		}
	}
	stage("alipay", "2026-05-23", "example/alipay/template.yaml", "example/alipay/rules.yaml")
	stage("wechat", "2026-04-28", "example/wechat/latest.yaml", "example/wechat/rules.yaml")
	stage("oklink", "2026-05-26", "example/oklink/template.yaml", "example/oklink/rules.yaml")
	stage("td", "2026-05-27", "example/td/template.yaml", "example/td/rules.yaml")
	stage("bmo-debit", "2026-05-28", "example/bmo/debit/template.yaml", "example/bmo/debit/rules.yaml")
	stage("bmo-credit", "2026-05-28", "example/bmo/credit/template.yaml", "example/bmo/credit/rules.yaml")
	stage("cmb-credit", "2026-05-28", "example/cmb/credit/template.yaml", "example/cmb/credit/rules.yaml")
	stage("cmb-debit", "2026-05-28", "example/cmb/debit/template.yaml", "example/cmb/debit/rules.yaml")
	stage("bocom_credit", "2026-05-28", "example/bocom_credit/template.yaml", "example/bocom_credit/rules.yaml")
	stage("bocom_debit", "2026-05-28", "example/bocom_debit/template.yaml", "example/bocom_debit/rules.yaml")
	stage("abc_debit", "2026-05-28", "example/abc_debit/template.yaml", "example/abc_debit/rules.yaml")
	stage("citic-credit", "2026-05-28", "example/citic/credit/template.yaml", "example/citic/credit/rules.yaml")
	stage("hsbchk-debit", "2026-05-28", "example/hsbchk/debit/template.yaml", "example/hsbchk/debit/rules.yaml")
	stage("hsbchk-credit", "2026-05-28", "example/hsbchk/credit/template.yaml", "example/hsbchk/credit/rules.yaml")
	stage("icbc-credit", "2026-05-28", "example/icbc/credit/template.yaml", "example/icbc/credit/rules.yaml")
	stage("icbc-debit-v1", "2026-05-28", "example/icbc/debit-v1/template.yaml", "example/icbc/debit-v1/rules.yaml")
	stage("icbc-debit-v2", "2026-05-28", "example/icbc/debit-v2/template.yaml", "example/icbc/debit-v2/rules.yaml")
	stage("boc-credit", "2026-05-28", "example/boc/credit/template.yaml", "example/boc/credit/rules.yaml")
	stage("boc-debit", "2026-05-28", "example/boc/debit/template.yaml", "example/boc/debit/rules.yaml")
	stage("jd", "2026-05-28", "example/jd/template.yaml", "example/jd/rules.yaml")
	stage("mt", "2026-05-28", "example/mt/template.yaml", "example/mt/rules.yaml")
	stage("spdb_debit", "2026-05-28", "example/spdb_debit/template.yaml", "example/spdb_debit/rules.yaml")
	stage("ccb", "2026-05-28", "example/ccb/template.yaml", "example/ccb/rules.yaml")
	stage("huobi", "2026-05-28", "example/huobi/template.yaml", "example/huobi/rules.yaml")
	stage("htsec", "2026-05-28", "example/htsec/template.yaml", "example/htsec/rules.yaml")
	stage("hxsec", "2026-05-28", "example/hxsec/template.yaml", "example/hxsec/rules.yaml")

	registry := `version: 1
templates:
  - id: alipay
    name: 支付宝账单
    latest: "2026-05-23"
    versions:
      - "2026-05-23"
    path: alipay/latest/template.yaml
    starterRules: alipay/latest/rules.yaml
  - id: wechat
    name: 微信支付账单
    latest: "2026-04-28"
    versions:
      - "2026-04-28"
    path: wechat/latest/template.yaml
    starterRules: wechat/latest/rules.yaml
  - id: oklink
    name: OKLink 代币账单
    latest: "2026-05-26"
    versions:
      - "2026-05-26"
    path: oklink/latest/template.yaml
    starterRules: oklink/latest/rules.yaml
  - id: td
    name: TD Bank
    latest: "2026-05-27"
    versions:
      - "2026-05-27"
    path: td/latest/template.yaml
    starterRules: td/latest/rules.yaml
  - id: bmo-debit
    name: BMO Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: bmo-debit/latest/template.yaml
    starterRules: bmo-debit/latest/rules.yaml
  - id: bmo-credit
    name: BMO Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: bmo-credit/latest/template.yaml
    starterRules: bmo-credit/latest/rules.yaml
  - id: cmb-credit
    name: CMB Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: cmb-credit/latest/template.yaml
    starterRules: cmb-credit/latest/rules.yaml
  - id: cmb-debit
    name: CMB Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: cmb-debit/latest/template.yaml
    starterRules: cmb-debit/latest/rules.yaml
  - id: bocom_credit
    name: BOCOM Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: bocom_credit/latest/template.yaml
    starterRules: bocom_credit/latest/rules.yaml
  - id: bocom_debit
    name: BOCOM Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: bocom_debit/latest/template.yaml
    starterRules: bocom_debit/latest/rules.yaml
  - id: abc_debit
    name: ABC Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: abc_debit/latest/template.yaml
    starterRules: abc_debit/latest/rules.yaml
  - id: citic-credit
    name: CITIC Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: citic-credit/latest/template.yaml
    starterRules: citic-credit/latest/rules.yaml
  - id: hsbchk-debit
    name: HSBC HK Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: hsbchk-debit/latest/template.yaml
    starterRules: hsbchk-debit/latest/rules.yaml
  - id: hsbchk-credit
    name: HSBC HK Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: hsbchk-credit/latest/template.yaml
    starterRules: hsbchk-credit/latest/rules.yaml
  - id: icbc-credit
    name: ICBC Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: icbc-credit/latest/template.yaml
    starterRules: icbc-credit/latest/rules.yaml
  - id: icbc-debit-v1
    name: ICBC Debit V1
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: icbc-debit-v1/latest/template.yaml
    starterRules: icbc-debit-v1/latest/rules.yaml
  - id: icbc-debit-v2
    name: ICBC Debit V2
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: icbc-debit-v2/latest/template.yaml
    starterRules: icbc-debit-v2/latest/rules.yaml
  - id: boc-credit
    name: BOC Credit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: boc-credit/latest/template.yaml
    starterRules: boc-credit/latest/rules.yaml
  - id: boc-debit
    name: BOC Debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: boc-debit/latest/template.yaml
    starterRules: boc-debit/latest/rules.yaml
  - id: jd
    name: JD
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: jd/latest/template.yaml
    starterRules: jd/latest/rules.yaml
  - id: mt
    name: Meituan
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: mt/latest/template.yaml
    starterRules: mt/latest/rules.yaml
  - id: spdb_debit
    name: spdb_debit
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: spdb_debit/latest/template.yaml
    starterRules: spdb_debit/latest/rules.yaml
  - id: ccb
    name: ccb
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: ccb/latest/template.yaml
    starterRules: ccb/latest/rules.yaml
  - id: huobi
    name: huobi
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: huobi/latest/template.yaml
    starterRules: huobi/latest/rules.yaml
  - id: htsec
    name: htsec
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: htsec/latest/template.yaml
    starterRules: htsec/latest/rules.yaml
  - id: hxsec
    name: hxsec
    latest: "2026-05-28"
    versions:
      - "2026-05-28"
    path: hxsec/latest/template.yaml
    starterRules: hxsec/latest/rules.yaml
`
	path := filepath.Join(dir, "registry.yaml")
	if err := os.WriteFile(path, []byte(registry), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func loadRules(t *testing.T, path string) ruleFile {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var wrapped ruleFile
	if err := yaml.Unmarshal(b, &wrapped); err != nil {
		t.Fatal(err)
	}
	return wrapped
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
