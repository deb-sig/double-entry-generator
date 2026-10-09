package importer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// RulesFile is the user's side of an import: the file `config init` writes
// and `import --rules` reads. The CLI and the browser build both merge it
// onto a template profile through ApplyRulesFile.
type RulesFile struct {
	ProtocolVersion       string            `yaml:"protocolVersion"`
	RequiredCapabilities  []string          `yaml:"requiredCapabilities"`
	Template              string            `yaml:"template"`
	TemplateRules         []Rule            `yaml:"templateRules"`
	TemplateRuleOverrides []Rule            `yaml:"templateRuleOverrides"`
	PersonalRules         []Rule            `yaml:"personalRules"`
	Rules                 []Rule            `yaml:"rules"`
	Accounts              map[string]string `yaml:"accounts"`
	Reconcile             *Reconcile        `yaml:"reconcile"`
	Output                *OutputPrefs      `yaml:"output"`
	Options               RulesOptions      `yaml:"options"`
}

type RulesOptions struct {
	Title             string `yaml:"title"`
	OperatingCurrency string `yaml:"operatingCurrency"`
}

// AllPersonalRules returns personalRules followed by rules.
func (r RulesFile) AllPersonalRules() []Rule {
	return append(append([]Rule{}, r.PersonalRules...), r.Rules...)
}

// ParseRulesFile decodes a rules file and rejects unsupported protocols
// before anything is merged.
func ParseRulesFile(data []byte) (RulesFile, error) {
	var rf RulesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return RulesFile{}, fmt.Errorf("rules file: %w", err)
	}
	contract := Profile{ProtocolVersion: rf.ProtocolVersion, RequiredCapabilities: rf.RequiredCapabilities}
	if err := contract.ValidateCapabilities(); err != nil {
		return RulesFile{}, err
	}
	return rf, nil
}

// ApplyRulesFile merges a rules file onto the profile: rules append,
// account bindings override by role, options override title and currency.
func (p *Profile) ApplyRulesFile(rf RulesFile) {
	p.TemplateRules = append(p.TemplateRules, rf.TemplateRules...)
	p.TemplateRuleOverrides = append(p.TemplateRuleOverrides, rf.TemplateRuleOverrides...)
	p.PersonalRules = append(p.PersonalRules, rf.AllPersonalRules()...)
	if len(rf.Accounts) > 0 {
		if p.Accounts == nil {
			p.Accounts = map[string]string{}
		}
		for role, account := range rf.Accounts {
			p.Accounts[role] = account
		}
	}
	if rf.Reconcile != nil {
		p.Reconcile = rf.Reconcile
	}
	if !rf.Output.IsZero() {
		p.Output = rf.Output
	}
	if rf.ProtocolVersion != "" {
		p.ProtocolVersion = rf.ProtocolVersion
	}
	if len(rf.RequiredCapabilities) > 0 {
		p.RequiredCapabilities = append(p.RequiredCapabilities, rf.RequiredCapabilities...)
	}
	if rf.Options.Title != "" {
		p.Name = rf.Options.Title
	}
	if rf.Options.OperatingCurrency != "" {
		p.Template.DefaultCurrency = rf.Options.OperatingCurrency
	}
}

// LoadProfileBytes parses a template profile held in memory.
func LoadProfileBytes(data []byte, fallbackID string) (*Profile, error) {
	return loadProfileBytes(data, fallbackID)
}
