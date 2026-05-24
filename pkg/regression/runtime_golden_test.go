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
	Rules                 []importer.Rule `yaml:"rules"`
	TemplateRuleOverrides []importer.Rule `yaml:"templateRuleOverrides"`
	PersonalRules         []importer.Rule `yaml:"personalRules"`
	UserRules             []importer.Rule `yaml:"userRules"`
}

func TestRuntimeTemplatesMatchLegacyGoldens(t *testing.T) {
	root := filepath.Join("..", "..")
	tests := []struct {
		name     string
		profile  string
		rules    string
		bill     string
		expected string
	}{
		{
			name:     "alipay latest",
			profile:  filepath.Join(root, "example", "alipay", "template.yaml"),
			rules:    filepath.Join(root, "example", "alipay", "rules.yaml"),
			bill:     filepath.Join(root, "example", "alipay", "example-alipay-records.csv"),
			expected: filepath.Join(root, "example", "alipay", "example-alipay-output.beancount"),
		},
		{
			name:     "wechat latest",
			profile:  filepath.Join(root, "example", "wechat", "latest.yaml"),
			rules:    filepath.Join(root, "example", "wechat", "rules.yaml"),
			bill:     filepath.Join(root, "example", "wechat", "example-wechat-records.xlsx"),
			expected: filepath.Join(root, "example", "wechat", "example-wechat-output.beancount"),
		},
		{
			name:     "wechat 2026-04-28",
			profile:  filepath.Join(root, "example", "wechat", "2026-04-28.yaml"),
			rules:    filepath.Join(root, "example", "wechat", "rules.yaml"),
			bill:     filepath.Join(root, "example", "wechat", "example-wechat-records.xlsx"),
			expected: filepath.Join(root, "example", "wechat", "example-wechat-output.beancount"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compileRuntimeTemplate(t, tt.profile, tt.rules, tt.bill)
			want, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Fatalf("runtime output differs from legacy golden\nprofile: %s\nbill: %s\nexpected: %s", tt.profile, tt.bill, tt.expected)
			}
		})
	}
}

func compileRuntimeTemplate(t *testing.T, profilePath, rulesPath, billPath string) []byte {
	t.Helper()
	profile, err := importer.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	rules := loadRules(t, rulesPath)
	profile.TemplateRuleOverrides = append(profile.TemplateRuleOverrides, rules.TemplateRuleOverrides...)
	profile.PersonalRules = append(profile.PersonalRules, rules.Rules...)
	profile.PersonalRules = append(profile.PersonalRules, rules.PersonalRules...)
	profile.PersonalRules = append(profile.PersonalRules, rules.UserRules...)

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

func loadRules(t *testing.T, path string) ruleFile {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var direct []importer.Rule
	if err := yaml.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
		return ruleFile{Rules: direct}
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
