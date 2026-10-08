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
	// Shape is the ordered list of row-level steps run between the reader
	// and the rules: header location, dropping, merging and splitting rows.
	Shape                 []ShapeOp         `json:"shape,omitempty" yaml:"shape,omitempty"`
	TemplateRules         []Rule            `json:"templateRules,omitempty" yaml:"templateRules,omitempty"`
	TemplateRuleOverrides []Rule            `json:"templateRuleOverrides,omitempty" yaml:"templateRuleOverrides,omitempty"`
	PersonalRules         []Rule            `json:"personalRules,omitempty" yaml:"personalRules,omitempty"`
	// Accounts binds leg roles to the user's accounts, once for the whole
	// file. A rule's from/to or accounts action overrides it per transaction.
	Accounts              map[string]string `json:"accounts,omitempty" yaml:"accounts,omitempty"`
	Defaults              map[string]string `json:"defaults,omitempty" yaml:"defaults,omitempty"`
}

type Template struct {
	FileFormat        string            `json:"fileFormat,omitempty" yaml:"fileFormat,omitempty"`
	Encoding          string            `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Delimiter         string            `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
	StripTabs         bool              `json:"stripTabs,omitempty" yaml:"stripTabs,omitempty"`
	SkipLeadingRows   int               `json:"skipLeadingRows,omitempty" yaml:"skipLeadingRows,omitempty"`
	SkipInvalidRows   bool              `json:"skipInvalidRows,omitempty" yaml:"skipInvalidRows,omitempty"`
	DateFormat        string            `json:"dateFormat,omitempty" yaml:"dateFormat,omitempty"`
	AmountPrefix      string            `json:"amountPrefix,omitempty" yaml:"amountPrefix,omitempty"`
	SourceHeaders     []string          `json:"sourceHeaders,omitempty" yaml:"sourceHeaders,omitempty"`
	HeaderLocate      bool              `json:"headerLocate,omitempty" yaml:"headerLocate,omitempty"`
	HeaderScanMaxRows int               `json:"headerScanMaxRows,omitempty" yaml:"headerScanMaxRows,omitempty"`
	Columns           ColumnMapping     `json:"columns,omitempty" yaml:"columns,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Slots is the Beancount field contract. When set, the template only fills
	// these fields and metadata; it does not assign accounts.
	Slots           SlotMapping `json:"slots,omitempty" yaml:"slots,omitempty"`
	// Direction says which way money moves. It replaces AmountSign, which
	// read the sign back out of metadata; AmountSign is still honoured when
	// Direction is empty.
	Direction  Direction  `json:"direction,omitempty" yaml:"direction,omitempty"`
	AmountSign AmountSign `json:"amountSign,omitempty" yaml:"amountSign,omitempty"`
	// Vars are template-level values usable as <var.x> in slots and legs,
	// such as a commodity symbol built from a security code. They are never
	// accounts; validation rejects account names here.
	Vars VarSets `json:"vars,omitempty" yaml:"vars,omitempty"`
	// Legs declares multi-leg transactions by role. The first branch whose
	// `when` holds is used; a branch without `when` is the default. With no
	// matching branch the transaction is the plain two-leg from/to shape.
	Legs []LegBranch `json:"legs,omitempty" yaml:"legs,omitempty"`
	DefaultMinus    string      `json:"defaultMinusAccount,omitempty" yaml:"defaultMinusAccount,omitempty"`
	DefaultPlus     string      `json:"defaultPlusAccount,omitempty" yaml:"defaultPlusAccount,omitempty"`
	DefaultCurrency string      `json:"defaultCurrency,omitempty" yaml:"defaultCurrency,omitempty"`
}

// SlotMapping binds bill columns to Beancount transaction fields.
// Provider-specific columns (支付方式, 收/支, …) belong in Metadata, not here.
type SlotMapping struct {
	Date      string         `json:"date,omitempty" yaml:"date,omitempty"`
	Payee     string         `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration string         `json:"narration,omitempty" yaml:"narration,omitempty"`
	Amount    string         `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency  string         `json:"currency,omitempty" yaml:"currency,omitempty"`
	Flag      string         `json:"flag,omitempty" yaml:"flag,omitempty"`
	Tags      string         `json:"tags,omitempty" yaml:"tags,omitempty"`
	Links     string         `json:"links,omitempty" yaml:"links,omitempty"`
	Metadata  orderedStrings `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// AmountSign negates the parsed amount when a metadata value matches.
// 收/支 stays metadata; the sign is applied onto the posting amount.
type AmountSign struct {
	Metadata string   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Negate   []string `json:"negate,omitempty" yaml:"negate,omitempty"`
}

func (a AmountSign) IsZero() bool { return a.Metadata == "" && len(a.Negate) == 0 }

// Direction decides whether an amount is an outflow (expense side) or an
// inflow. Exactly one form is used:
//
//	direction: { column: "<收/支>", outflow: [支出], inflow: [收入] }   values of one column
//	direction: { outflowColumn: "<支出金额>", inflowColumn: "<收入金额>" } two amount columns
//	direction: {}                                                  sign of slots.amount
type Direction struct {
	Column        string   `json:"column,omitempty" yaml:"column,omitempty"`
	Outflow       []string `json:"outflow,omitempty" yaml:"outflow,omitempty"`
	Inflow        []string `json:"inflow,omitempty" yaml:"inflow,omitempty"`
	OutflowColumn string   `json:"outflowColumn,omitempty" yaml:"outflowColumn,omitempty"`
	InflowColumn  string   `json:"inflowColumn,omitempty" yaml:"inflowColumn,omitempty"`
}

func (d Direction) IsZero() bool {
	return d.Column == "" && len(d.Outflow) == 0 && len(d.Inflow) == 0 && d.OutflowColumn == "" && d.InflowColumn == ""
}

// VarSet is a group of template vars, optionally guarded by a condition.
type VarSet struct {
	When string            `json:"when,omitempty" yaml:"when,omitempty"`
	Vars map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"`
}

// VarSets accepts either a plain mapping (one unconditional set) or a list
// of {when, vars} sets applied in order, later sets overriding earlier.
type VarSets []VarSet

func (v *VarSets) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.MappingNode:
		var vars map[string]flexibleString
		if err := value.Decode(&vars); err != nil {
			return err
		}
		*v = VarSets{{Vars: flexibleStringMap(vars)}}
		return nil
	case yaml.SequenceNode:
		type set struct {
			When flexibleString            `yaml:"when,omitempty"`
			Vars map[string]flexibleString `yaml:"vars,omitempty"`
		}
		var sets []set
		if err := value.Decode(&sets); err != nil {
			return err
		}
		out := make(VarSets, 0, len(sets))
		for _, s := range sets {
			out = append(out, VarSet{When: string(s.When), Vars: flexibleStringMap(s.Vars)})
		}
		*v = out
		return nil
	case 0:
		return nil
	default:
		return fmt.Errorf("vars must be a mapping or a list of {when, vars}")
	}
}

// LegBranch is one shape a template transaction can take.
type LegBranch struct {
	ID        string            `json:"id,omitempty" yaml:"id,omitempty"`
	When      string            `json:"when,omitempty" yaml:"when,omitempty"`
	Payee     string            `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration string            `json:"narration,omitempty" yaml:"narration,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Legs      []LegSpec         `json:"legs" yaml:"legs"`
}

// LegSpec is one posting described by role, never by account. Amount,
// Currency, Cost and Price are expressions rendered against the row. An
// empty Amount leaves the posting to Beancount's auto-balance (pnl legs).
type LegSpec struct {
	Role     string `json:"role" yaml:"role"`
	Amount   string `json:"amount" yaml:"amount"`
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
	Cost     string `json:"cost,omitempty" yaml:"cost,omitempty"`
	Price    string `json:"price,omitempty" yaml:"price,omitempty"`
}

func (l *LegSpec) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("leg must be a mapping with role and amount")
	}
	type leg struct {
		Role     flexibleString `yaml:"role"`
		Amount   flexibleString `yaml:"amount"`
		Currency flexibleString `yaml:"currency"`
		Cost     flexibleString `yaml:"cost"`
		Price    flexibleString `yaml:"price"`
	}
	for i := 0; i+1 < len(value.Content); i += 2 {
		switch key := value.Content[i].Value; key {
		case "role", "amount", "currency", "cost", "price":
		case "account":
			return fmt.Errorf("leg %q: templates bind roles, not accounts; put the account in the rules file under accounts:", value.Content[i+1].Value)
		default:
			return fmt.Errorf("leg: unknown key %q", key)
		}
	}
	var out leg
	if err := value.Decode(&out); err != nil {
		return err
	}
	*l = LegSpec{Role: string(out.Role), Amount: string(out.Amount), Currency: string(out.Currency), Cost: string(out.Cost), Price: string(out.Price)}
	return nil
}

// CoreRoles is the closed set of leg roles the engine understands. Each
// maps to the FIXME account used when nobody binds it. Roles outside the
// set must start with "x-" and must be bound by the rules file.
var CoreRoles = map[string]string{
	"from":     "", // decided by direction
	"to":       "", // decided by direction
	"cash":     "Assets:FIXME",
	"custody":  "Assets:FIXME",
	"position": "Assets:FIXME",
	"fee":      "Expenses:FIXME",
	"gas":      "Expenses:FIXME",
	"pnl":      "Income:FIXME",
}

func validateRole(role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return fmt.Errorf("leg role is required")
	}
	if _, ok := CoreRoles[role]; ok {
		return nil
	}
	if strings.HasPrefix(role, "x-") && len(role) > 2 {
		return nil
	}
	return fmt.Errorf("unknown leg role %q; use one of from, to, cash, custody, position, fee, gas, pnl, or an x- prefixed custom role", role)
}

// orderedStrings keeps YAML mapping order for metadata keys.
type orderedStrings struct {
	Keys   []string
	Values map[string]string
}

func (m *orderedStrings) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("metadata must be a mapping")
	}
	m.Values = make(map[string]string, len(value.Content)/2)
	for i := 0; i+1 < len(value.Content); i += 2 {
		key := strings.TrimSpace(value.Content[i].Value)
		var text string
		if err := value.Content[i+1].Decode(&text); err != nil {
			return err
		}
		if key == "" {
			continue
		}
		if _, ok := m.Values[key]; !ok {
			m.Keys = append(m.Keys, key)
		}
		m.Values[key] = text
	}
	return nil
}

func (t Template) HasSlotContract() bool {
	s := t.Slots
	return s.Date != "" || s.Payee != "" || s.Narration != "" || s.Amount != "" ||
		s.Currency != "" || s.Flag != "" || s.Tags != "" || s.Links != "" ||
		len(s.Metadata.Keys) > 0
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
	Date         string            `json:"date,omitempty" yaml:"date,omitempty"`
	Type         string            `json:"type,omitempty" yaml:"type,omitempty"`
	Note         string            `json:"note,omitempty" yaml:"note,omitempty"`
	From         TransferSide      `json:"from,omitempty" yaml:"from,omitempty"`
	To           TransferSide      `json:"to,omitempty" yaml:"to,omitempty"`
	Payee        string            `json:"payee,omitempty" yaml:"payee,omitempty"`
	Narration    string            `json:"narration,omitempty" yaml:"narration,omitempty"`
	Amount       string            `json:"amount,omitempty" yaml:"amount,omitempty"`
	Currency     string            `json:"currency,omitempty" yaml:"currency,omitempty"`
	Tag          string            `json:"tag,omitempty" yaml:"tag,omitempty"`
	Tags         []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Ignore       bool              `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	Flag         string            `json:"flag,omitempty" yaml:"flag,omitempty"`
	Link         string            `json:"link,omitempty" yaml:"link,omitempty"`
	Vars         map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	MetadataDrop []string          `json:"metadataDrop,omitempty" yaml:"metadataDrop,omitempty"`
	// Accounts binds leg roles for this transaction. from/to are shorthand
	// for accounts: {from: …, to: …}.
	Accounts map[string]string `json:"accounts,omitempty" yaml:"accounts,omitempty"`
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
		Date         flexibleString            `yaml:"date,omitempty"`
		Type         flexibleString            `yaml:"type,omitempty"`
		Note         flexibleString            `yaml:"note,omitempty"`
		From         TransferSide              `yaml:"from,omitempty"`
		To           TransferSide              `yaml:"to,omitempty"`
		Payee        flexibleString            `yaml:"payee,omitempty"`
		Narration    flexibleString            `yaml:"narration,omitempty"`
		Amount       flexibleString            `yaml:"amount,omitempty"`
		Currency     flexibleString            `yaml:"currency,omitempty"`
		Tag          flexibleString            `yaml:"tag,omitempty"`
		Tags         []string                  `yaml:"tags,omitempty"`
		Ignore       bool                      `yaml:"ignore,omitempty"`
		Flag         flexibleString            `yaml:"flag,omitempty"`
		Link         flexibleString            `yaml:"link,omitempty"`
		Vars         map[string]flexibleString `yaml:"vars,omitempty"`
		Metadata     map[string]flexibleString `yaml:"metadata,omitempty"`
		MetadataDrop []string                  `yaml:"metadataDrop,omitempty"`
		Accounts     map[string]flexibleString `yaml:"accounts,omitempty"`
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
		MetadataDrop: out.MetadataDrop,
		Accounts:     flexibleStringMap(out.Accounts),
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
		len(actions.MetadataDrop) == 0 &&
		len(actions.Accounts) == 0 &&
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
	if t.HasSlotContract() {
		if t.Slots.Date == "" {
			return fmt.Errorf("template slots.date is required")
		}
		d := t.Direction
		twoColumns := d.OutflowColumn != "" || d.InflowColumn != ""
		if twoColumns && (d.OutflowColumn == "" || d.InflowColumn == "") {
			return fmt.Errorf("template direction needs both outflowColumn and inflowColumn")
		}
		if twoColumns && (d.Column != "" || len(d.Outflow) > 0 || len(d.Inflow) > 0) {
			return fmt.Errorf("template direction: use either column/outflow/inflow or outflowColumn/inflowColumn, not both")
		}
		if d.Column == "" && (len(d.Outflow) > 0 || len(d.Inflow) > 0) {
			return fmt.Errorf("template direction.outflow/inflow need direction.column")
		}
		if t.Slots.Amount == "" && !twoColumns {
			return fmt.Errorf("template slots.amount is required (or direction.outflowColumn/inflowColumn)")
		}
		for _, set := range t.Vars {
			for name, value := range set.Vars {
				if isAccountName(strings.TrimSpace(value)) {
					return fmt.Errorf("template vars.%s is an account name; templates bind roles, accounts belong in the rules file", name)
				}
			}
		}
		for i, branch := range t.Legs {
			if len(branch.Legs) == 0 {
				return fmt.Errorf("template legs[%d]: branch has no legs", i)
			}
			for j, leg := range branch.Legs {
				if err := validateRole(leg.Role); err != nil {
					return fmt.Errorf("template legs[%d].legs[%d]: %w", i, j, err)
				}
			}
		}
		return nil
	}
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
	"when.starts_with":                {},
	"when.ends_with":                  {},
	"when.regex":                      {},
	"when.literalFieldLookup":         {},
	"actions.flag":                    {},
	"actions.link":                    {},
	"actions.replace":                 {},
	"actions.quoted_literal":          {},
	"actions.rawColumnRef":            {},
	"actions.postingsMode":            {},
	"actions.postingPriceCost":        {},
	"actions.dynamicPostingPriceCost": {},
	"rule.templateId":                 {},
	"template.headerLocate":           {},
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
