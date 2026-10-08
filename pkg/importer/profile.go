package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/deb-sig/double-entry-generator/v2/pkg/reader"
)

const DefaultProviderName = "template"

type Profile struct {
	Schema                string            `json:"schema,omitempty" yaml:"schema,omitempty"`
	ID                    string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name                  string            `json:"name,omitempty" yaml:"name,omitempty"`
	ProtocolVersion       string            `json:"protocolVersion,omitempty" yaml:"protocolVersion,omitempty"`
	RequiredCapabilities  []string          `json:"requiredCapabilities,omitempty" yaml:"requiredCapabilities,omitempty"`
	Template              Template          `json:"template" yaml:"template"`
	// Reader is the `reader:` block: how bytes become a table. When absent the
	// legacy template.fileFormat/encoding/delimiter fields are used instead.
	Reader                *reader.Config    `json:"reader,omitempty" yaml:"reader,omitempty"`
	TemplateRules         []Rule            `json:"templateRules,omitempty" yaml:"templateRules,omitempty"`
	TemplateRuleOverrides []Rule            `json:"templateRuleOverrides,omitempty" yaml:"templateRuleOverrides,omitempty"`
	PersonalRules         []Rule            `json:"personalRules,omitempty" yaml:"personalRules,omitempty"`
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
	HeaderLocate bool `json:"headerLocate,omitempty" yaml:"headerLocate,omitempty"`
	HeaderScanMaxRows int `json:"headerScanMaxRows,omitempty" yaml:"headerScanMaxRows,omitempty"`
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
	ID         string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name       string  `json:"name,omitempty" yaml:"name,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	TemplateID string  `json:"templateId,omitempty" yaml:"templateId,omitempty"`
	When       string  `json:"when,omitempty" yaml:"when,omitempty"`
	Actions    Actions `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type Actions struct {
	Date      string            `json:"date,omitempty" yaml:"date,omitempty"`
	Type      string            `json:"type,omitempty" yaml:"type,omitempty"`
	Note      string            `json:"note,omitempty" yaml:"note,omitempty"`
	From      TransferSide      `json:"from,omitempty" yaml:"from,omitempty"`
	To        TransferSide      `json:"to,omitempty" yaml:"to,omitempty"`
	Payee     string            `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration string            `json:"narration,omitempty" yaml:"narration,omitempty"`
	Amount    string            `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency  string            `json:"currency,omitempty" yaml:"currency,omitempty"`
	Tag       string            `json:"tag,omitempty" yaml:"tag,omitempty"`
	Tags      []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Ignore    bool              `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	Flag      string            `json:"flag,omitempty" yaml:"flag,omitempty"`
	Link      string            `json:"link,omitempty" yaml:"link,omitempty"`
	Vars         map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Postings     []string          `json:"postings,omitempty" yaml:"postings,omitempty"`
	// PostingsMode controls how Postings combine with auto from/to legs.
	// "" or "append" (DEG legacy default): render from/to then append Postings.
	// "replace" (Mirato export): Postings are the complete leg set; do not also
	// synthesize from/to amounts. Across rules, a later replace overwrites the
	// prior complete set; later append adds onto it.
	PostingsMode string `json:"postingsMode,omitempty" yaml:"postingsMode,omitempty"`
}

type TransferSide struct {
	Account  string `json:"account,omitempty" yaml:"account,omitempty"`
	Amount   string `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
}

type flexibleString string

func (s *flexibleString) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*s = flexibleString(strings.TrimSpace(value.Value))
		return nil
	case 0:
		return nil
	default:
		return fmt.Errorf("value must be a string")
	}
}

func (r *Rule) UnmarshalYAML(value *yaml.Node) error {
	type rule struct {
		ID         string         `yaml:"id,omitempty"`
		Name       string         `yaml:"name,omitempty"`
		Enabled    *bool          `yaml:"enabled,omitempty"`
		TemplateID string         `yaml:"templateId,omitempty"`
		When       flexibleString `yaml:"when,omitempty"`
		Actions    Actions        `yaml:"actions,omitempty"`
	}
	var out rule
	if err := value.Decode(&out); err != nil {
		return err
	}
	*r = Rule{
		ID:         out.ID,
		Name:       out.Name,
		Enabled:    out.Enabled,
		TemplateID: out.TemplateID,
		When:       string(out.When),
		Actions:    out.Actions,
	}
	return nil
}

func (a *Actions) UnmarshalYAML(value *yaml.Node) error {
	type actions struct {
		Date      flexibleString            `yaml:"date,omitempty"`
		Type      flexibleString            `yaml:"type,omitempty"`
		Note      flexibleString            `yaml:"note,omitempty"`
		From      TransferSide              `yaml:"from,omitempty"`
		To        TransferSide              `yaml:"to,omitempty"`
		Payee     flexibleString            `yaml:"payee,omitempty"`
		Narration flexibleString            `yaml:"narration,omitempty"`
		Amount    flexibleString            `yaml:"amount,omitempty"`
		Currency  flexibleString            `yaml:"currency,omitempty"`
		Tag       flexibleString            `yaml:"tag,omitempty"`
		Tags      []string                  `yaml:"tags,omitempty"`
		Ignore    bool                      `yaml:"ignore,omitempty"`
		Flag      flexibleString            `yaml:"flag,omitempty"`
		Link      flexibleString            `yaml:"link,omitempty"`
		Vars         map[string]flexibleString `yaml:"vars,omitempty"`
		Metadata     map[string]flexibleString `yaml:"metadata,omitempty"`
		Postings     []flexibleString          `yaml:"postings,omitempty"`
		PostingsMode flexibleString            `yaml:"postingsMode,omitempty"`
	}
	var out actions
	if err := value.Decode(&out); err != nil {
		return err
	}
	*a = Actions{
		Date:         string(out.Date),
		Type:         string(out.Type),
		Note:         string(out.Note),
		From:         out.From,
		To:           out.To,
		Payee:        string(out.Payee),
		Narration:    string(out.Narration),
		Amount:       string(out.Amount),
		Currency:     string(out.Currency),
		Tag:          string(out.Tag),
		Tags:         out.Tags,
		Ignore:       out.Ignore,
		Flag:         string(out.Flag),
		Link:         string(out.Link),
		Vars:         flexibleStringMap(out.Vars),
		Metadata:     flexibleStringMap(out.Metadata),
		Postings:     flexibleStringSlice(out.Postings),
		PostingsMode: string(out.PostingsMode),
	}
	return nil
}

func (s *TransferSide) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		s.Account = strings.TrimSpace(value.Value)
		return nil
	case yaml.MappingNode:
		type side struct {
			Account  flexibleString `yaml:"account,omitempty"`
			Amount   flexibleString `yaml:"amount,omitempty"`
			Currency flexibleString `yaml:"currency,omitempty"`
		}
		var out side
		if err := value.Decode(&out); err != nil {
			return err
		}
		*s = TransferSide{
			Account:  string(out.Account),
			Amount:   string(out.Amount),
			Currency: string(out.Currency),
		}
		return nil
	case 0:
		return nil
	default:
		return fmt.Errorf("from/to must be an account string or mapping")
	}
}

func flexibleStringMap(values map[string]flexibleString) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = string(value)
	}
	return out
}

func flexibleStringSlice(values []flexibleString) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func (s flexibleString) MarshalYAML() (any, error) {
	return string(s), nil
}

func (s TransferSide) IsZero() bool {
	return s.Account == "" && s.Amount == "" && s.Currency == ""
}

func isZeroActions(actions Actions) bool {
	return actions.Type == "" &&
		actions.Date == "" &&
		actions.From.IsZero() &&
		actions.To.IsZero() &&
		actions.Payee == "" &&
		actions.Narration == "" &&
		actions.Amount == "" &&
		actions.Currency == "" &&
		actions.Tag == "" &&
		len(actions.Tags) == 0 &&
		!actions.Ignore &&
		actions.Flag == "" &&
		actions.Link == "" &&
		len(actions.Vars) == 0 &&
		len(actions.Metadata) == 0 &&
		len(actions.Postings) == 0 &&
		actions.PostingsMode == ""
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
	// Reject incompatible base profiles before later rule files can replace
	// their protocol annotation during CLI assembly.
	if err := p.ValidateCapabilities(); err != nil {
		return nil, err
	}
	if p.ID == "" {
		p.ID = fallbackID
	}
	normalizeTemplate(&p.Template, p.Defaults)
	if err := validateTemplate(p); err != nil {
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

func validateTemplate(p Profile) error {
	t := p.Template
	if p.IsV2() && t.hasNoColumns() {
		return nil
	}
	if t.Columns.Date == "" {
		return fmt.Errorf("template columns.date is required")
	}
	if t.Columns.Amount == "" && (t.Columns.AmountIn == "" || t.Columns.AmountOut == "") {
		return fmt.Errorf("template columns.amount or both columns.amountIn/amountOut are required")
	}
	return nil
}

func (t Template) hasNoColumns() bool {
	return t.Columns.Date == "" &&
		t.Columns.Time == "" &&
		t.Columns.Amount == "" &&
		t.Columns.AmountIn == "" &&
		t.Columns.AmountOut == "" &&
		t.Columns.Payee == "" &&
		t.Columns.Narration == "" &&
		t.Columns.Type == "" &&
		t.Columns.Currency == ""
}

func (p *Profile) IsV2() bool {
	return strings.Contains(strings.ToLower(p.Schema), "/v2")
}

// SupportedCapabilities lists runtime features this DEG build implements for
// Mirato personal-rules exports. Unknown requiredCapabilities must fail closed.
var SupportedCapabilities = map[string]struct{}{
	"when.starts_with":       {},
	"when.ends_with":         {},
	"when.regex":             {},
	"when.literalFieldLookup": {},
	"actions.flag":           {},
	"actions.link":           {},
	"actions.replace":        {},
	"actions.quoted_literal": {},
	"actions.rawColumnRef":           {},
	"actions.postingsMode":           {},
	"actions.postingPriceCost":       {},
	"actions.dynamicPostingPriceCost": {},
	"rule.templateId":                {},
	"template.headerLocate":          {},
}

func (p *Profile) ValidateCapabilities() error {
	if p == nil {
		return nil
	}
	if p.ProtocolVersion != "" && p.ProtocolVersion != "mirato-deg-rules/1" {
		return fmt.Errorf("unsupported protocolVersion %q; upgrade DEG before importing these rules", p.ProtocolVersion)
	}
	for _, cap := range p.RequiredCapabilities {
		cap = strings.TrimSpace(cap)
		if cap == "" {
			continue
		}
		if _, ok := SupportedCapabilities[cap]; !ok {
			return fmt.Errorf("unsupported required capability %q (protocolVersion=%q); upgrade DEG or remove the capability requirement", cap, p.ProtocolVersion)
		}
	}
	for _, rule := range p.Rules() {
		if _, err := normalizePostingsMode(rule.Actions.PostingsMode); err != nil {
			if rule.ID != "" {
				return fmt.Errorf("rule %q: %w", rule.ID, err)
			}
			return err
		}
	}
	return nil
}

// normalizePostingsMode returns "append" or "replace". Empty defaults to append
// for legacy DEG local templates that only listed extra fee legs.
func normalizePostingsMode(mode string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "append":
		return "append", nil
	case "replace":
		return "replace", nil
	default:
		return "", fmt.Errorf("unsupported postingsMode %q (want append|replace)", mode)
	}
}

func (p *Profile) Rules() []Rule {
	rules := make([]Rule, 0, len(p.TemplateRules)+len(p.PersonalRules))
	rules = append(rules, p.TemplateRules...)
	rules = applyTemplateRuleOverrides(rules, p.TemplateRuleOverrides)
	rules = append(rules, p.PersonalRules...)
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
