package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultProviderName = "template"

type Profile struct {
	Schema                string            `json:"schema,omitempty" yaml:"schema,omitempty"`
	ID                    string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name                  string            `json:"name,omitempty" yaml:"name,omitempty"`
	Template              Template          `json:"template" yaml:"template"`
	TemplateRules         []Rule            `json:"templateRules,omitempty" yaml:"templateRules,omitempty"`
	TemplateRuleOverrides []Rule            `json:"templateRuleOverrides,omitempty" yaml:"templateRuleOverrides,omitempty"`
	PersonalRules         []Rule            `json:"personalRules,omitempty" yaml:"personalRules,omitempty"`
	UserRules             []Rule            `json:"userRules,omitempty" yaml:"userRules,omitempty"`
	Defaults              map[string]string `json:"defaults,omitempty" yaml:"defaults,omitempty"`
}

type Template struct {
	FileFormat      string            `json:"fileFormat,omitempty" yaml:"fileFormat,omitempty"`
	Encoding        string            `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Delimiter       string            `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
	StripTabs       bool              `json:"stripTabs,omitempty" yaml:"stripTabs,omitempty"`
	SkipLeadingRows int               `json:"skipLeadingRows,omitempty" yaml:"skipLeadingRows,omitempty"`
	SkipInvalidRows bool              `json:"skipInvalidRows,omitempty" yaml:"skipInvalidRows,omitempty"`
	DateFormat      string            `json:"dateFormat,omitempty" yaml:"dateFormat,omitempty"`
	AmountPrefix    string            `json:"amountPrefix,omitempty" yaml:"amountPrefix,omitempty"`
	SourceHeaders   []string          `json:"sourceHeaders,omitempty" yaml:"sourceHeaders,omitempty"`
	Columns         ColumnMapping     `json:"columns,omitempty" yaml:"columns,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	DefaultMinus    string            `json:"defaultMinusAccount,omitempty" yaml:"defaultMinusAccount,omitempty"`
	DefaultPlus     string            `json:"defaultPlusAccount,omitempty" yaml:"defaultPlusAccount,omitempty"`
	DefaultCurrency string            `json:"defaultCurrency,omitempty" yaml:"defaultCurrency,omitempty"`
}

type ColumnMapping struct {
	Date      string `json:"date,omitempty" yaml:"date,omitempty"`
	Time      string `json:"time,omitempty" yaml:"time,omitempty"`
	Amount    string `json:"amount,omitempty" yaml:"amount,omitempty"`
	AmountIn  string `json:"amountIn,omitempty" yaml:"amountIn,omitempty"`
	AmountOut string `json:"amountOut,omitempty" yaml:"amountOut,omitempty"`
	Payee     string `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration string `json:"narration,omitempty" yaml:"narration,omitempty"`
	Type      string `json:"type,omitempty" yaml:"type,omitempty"`
	Currency  string `json:"currency,omitempty" yaml:"currency,omitempty"`
}

type Rule struct {
	ID      string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name    string  `json:"name,omitempty" yaml:"name,omitempty"`
	Enabled *bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	When    string  `json:"when,omitempty" yaml:"when,omitempty"`
	Actions Actions `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type Actions struct {
	Type              string            `json:"type,omitempty" yaml:"type,omitempty"`
	From              TransferSide      `json:"from,omitempty" yaml:"from,omitempty"`
	To                TransferSide      `json:"to,omitempty" yaml:"to,omitempty"`
	Payee             string            `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration         string            `json:"narration,omitempty" yaml:"narration,omitempty"`
	Amount            string            `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency          string            `json:"currency,omitempty" yaml:"currency,omitempty"`
	Tag               string            `json:"tag,omitempty" yaml:"tag,omitempty"`
	Tags              []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Ignore            bool              `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Commission        string            `json:"commission,omitempty" yaml:"commission,omitempty"`
	CommissionAccount string            `json:"commissionAccount,omitempty" yaml:"commissionAccount,omitempty"`
	PnlAccount        string            `json:"pnlAccount,omitempty" yaml:"pnlAccount,omitempty"`
	Postings          []string          `json:"postings,omitempty" yaml:"postings,omitempty"`
}

type TransferSide struct {
	Account  string `json:"account,omitempty" yaml:"account,omitempty"`
	Amount   string `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
}

func (s *TransferSide) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		s.Account = strings.TrimSpace(value.Value)
		return nil
	case yaml.MappingNode:
		type side TransferSide
		var out side
		if err := value.Decode(&out); err != nil {
			return err
		}
		*s = TransferSide(out)
		return nil
	case 0:
		return nil
	default:
		return fmt.Errorf("from/to must be an account string or mapping")
	}
}

func (s TransferSide) IsZero() bool {
	return s.Account == "" && s.Amount == "" && s.Currency == ""
}

func isZeroActions(actions Actions) bool {
	return actions.Type == "" &&
		actions.From.IsZero() &&
		actions.To.IsZero() &&
		actions.Payee == "" &&
		actions.Narration == "" &&
		actions.Amount == "" &&
		actions.Currency == "" &&
		actions.Tag == "" &&
		len(actions.Tags) == 0 &&
		!actions.Ignore &&
		len(actions.Metadata) == 0 &&
		actions.Commission == "" &&
		actions.CommissionAccount == "" &&
		actions.PnlAccount == "" &&
		len(actions.Postings) == 0
}

func LoadProfile(path string) (*Profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
	default:
		return nil, fmt.Errorf("template profile must be a YAML file")
	}
	return loadProfileBytes(b, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
}

func loadProfileBytes(b []byte, fallbackID string) (*Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.ID == "" {
		p.ID = fallbackID
	}
	normalizeTemplate(&p.Template, p.Defaults)
	if err := validateTemplate(p.Template); err != nil {
		return nil, err
	}
	return &p, nil
}

func normalizeTemplate(t *Template, defaults map[string]string) {
	if t.Delimiter == "" {
		t.Delimiter = ","
	}
	if t.FileFormat == "" {
		t.FileFormat = "csv"
	}
	if t.DefaultMinus == "" {
		t.DefaultMinus = defaults["minusAccount"]
	}
	if t.DefaultPlus == "" {
		t.DefaultPlus = defaults["plusAccount"]
	}
	if t.DefaultCurrency == "" {
		t.DefaultCurrency = defaults["currency"]
	}
}

func validateTemplate(t Template) error {
	if t.Columns.Date == "" {
		return fmt.Errorf("template columns.date is required")
	}
	if t.Columns.Amount == "" && (t.Columns.AmountIn == "" || t.Columns.AmountOut == "") {
		return fmt.Errorf("template columns.amount or both columns.amountIn/amountOut are required")
	}
	return nil
}

func (p *Profile) IsV2() bool {
	return strings.Contains(strings.ToLower(p.Schema), "/v2")
}

func (p *Profile) Rules() []Rule {
	rules := make([]Rule, 0, len(p.TemplateRules)+len(p.PersonalRules)+len(p.UserRules))
	rules = append(rules, p.TemplateRules...)
	rules = applyTemplateRuleOverrides(rules, p.TemplateRuleOverrides)
	rules = append(rules, p.PersonalRules...)
	rules = append(rules, p.UserRules...)
	return enabledRules(rules)
}

func applyTemplateRuleOverrides(base []Rule, overrides []Rule) []Rule {
	if len(overrides) == 0 {
		return base
	}
	out := append([]Rule(nil), base...)
	for _, override := range overrides {
		replaced := false
		for i := range out {
			if out[i].ID == override.ID {
				out[i] = mergeRuleOverride(out[i], override)
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, override)
		}
	}
	return out
}

func mergeRuleOverride(base, override Rule) Rule {
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Enabled != nil {
		base.Enabled = override.Enabled
	}
	if override.When != "" {
		base.When = override.When
	}
	if !isZeroActions(override.Actions) {
		base.Actions = override.Actions
	}
	return base
}

func enabledRules(rules []Rule) []Rule {
	out := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if rule.Enabled != nil && !*rule.Enabled {
			continue
		}
		out = append(out, rule)
	}
	return out
}
