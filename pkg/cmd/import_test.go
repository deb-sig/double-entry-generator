package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
)

func TestLoadRegistryStarterRulesThenPersonalRules(t *testing.T) {
	root := t.TempDir()
	writeCmdTestFile(t, filepath.Join(root, "registry.yaml"), `version: 1
templates:
  - id: oklink
    latest: "2026-05-26"
    path: oklink/latest/template.yaml
    starterRules: oklink/latest/rules.yaml
`)
	writeCmdTestFile(t, filepath.Join(root, "oklink", "2026-05-26", "rules.yaml"), `templateRuleOverrides:
  - id: template-direction
    actions:
      type: send
personalRules:
  - id: template-asset
    actions:
      from: Assets:Crypto
`)
	personalPath := filepath.Join(root, "personal.yaml")
	writeCmdTestFile(t, personalPath, `personalRules:
  - id: user-food
    actions:
      to: Expenses:Food
`)

	t.Setenv(importer.RegistryURLEnv, filepath.Join(root, "registry.yaml"))
	starter, err := loadRegistryStarterRules("oklink@2026-05-26")
	if err != nil {
		t.Fatal(err)
	}
	personal, err := loadRuleFile(personalPath)
	if err != nil {
		t.Fatal(err)
	}

	profile := &importer.Profile{}
	appendRulesToProfile(profile, starter)
	appendRulesToProfile(profile, personal)

	if len(profile.TemplateRuleOverrides) != 1 || profile.TemplateRuleOverrides[0].ID != "template-direction" {
		t.Fatalf("template overrides not loaded first: %#v", profile.TemplateRuleOverrides)
	}
	if got := []string{profile.PersonalRules[0].ID, profile.PersonalRules[1].ID}; got[0] != "template-asset" || got[1] != "user-food" {
		t.Fatalf("rules loaded in wrong order: %#v", got)
	}
}

func writeCmdTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
