package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTemplateRef(t *testing.T) {
	tests := []struct {
		ref, wantID, wantVersion string
	}{
		{ref: "wechat", wantID: "wechat"},
		{ref: "wechat@2026-04-28", wantID: "wechat", wantVersion: "2026-04-28"},
		{ref: " wechat @ 2025-07-15 ", wantID: "wechat", wantVersion: "2025-07-15"},
	}
	for _, tt := range tests {
		id, version := ParseTemplateRef(tt.ref)
		if id != tt.wantID || version != tt.wantVersion {
			t.Fatalf("ParseTemplateRef(%q) = (%q, %q), want (%q, %q)", tt.ref, id, version, tt.wantID, tt.wantVersion)
		}
	}
}

func TestApplyVersionToPath(t *testing.T) {
	tests := []struct {
		path, latest, version, want string
	}{
		{path: "templates/wechat/latest.yaml", latest: "2026-04-28", version: "", want: "templates/wechat/latest.yaml"},
		{path: "wechat/latest/template.yaml", latest: "2026-04-28", version: "2025-07-15", want: "wechat/2025-07-15/template.yaml"},
		{path: "templates/wechat/2026-04-28.yaml", latest: "2026-04-28", version: "2025-07-15", want: "templates/wechat/2025-07-15.yaml"},
		{path: "templates/wechat/latest.yaml", latest: "2026-04-28", version: "2025-07-15", want: "templates/wechat/2025-07-15.yaml"},
	}
	for _, tt := range tests {
		if got := applyVersionToPath(tt.path, tt.latest, tt.version); got != tt.want {
			t.Fatalf("applyVersionToPath(%q, %q, %q) = %q, want %q", tt.path, tt.latest, tt.version, got, tt.want)
		}
	}
}

func TestLookupRegistryTemplate(t *testing.T) {
	registry := &Registry{
		Templates: []RegistryTemplate{{
			ID:     "wechat",
			Latest: "2026-04-28",
			Path:   "wechat/latest/template.yaml",
		}},
	}
	template, version, err := lookupRegistryTemplate(registry, "wechat@2025-07-15")
	if err != nil {
		t.Fatal(err)
	}
	if template.ID != "wechat" || version != "2025-07-15" {
		t.Fatalf("unexpected lookup result: %#v %q", template, version)
	}
	if _, _, err := lookupRegistryTemplate(registry, "wechat@"); err == nil {
		t.Fatal("expected error for empty @version")
	}
}

func TestTemplateURLFromRegistryPinsVersion(t *testing.T) {
	registry := &Registry{
		Templates: []RegistryTemplate{{
			ID:     "wechat",
			Latest: "2026-04-28",
			Path:   "wechat/latest/template.yaml",
		}},
	}
	template, version, err := lookupRegistryTemplate(registry, "wechat@2025-07-15")
	if err != nil {
		t.Fatal(err)
	}
	url := resolveRegistryAssetURL("", template.Path, template.Latest, version)
	want := registryBaseURL("") + "wechat/2025-07-15/template.yaml"
	if url != want {
		t.Fatalf("url = %q, want %q", url, want)
	}
}

func TestTemplateURLFromRegistryAllowsPinnedLatest(t *testing.T) {
	registry := &Registry{
		Templates: []RegistryTemplate{{
			ID:     "wechat",
			Latest: "2026-04-28",
			Path:   "wechat/latest/template.yaml",
		}},
	}
	template, version, err := lookupRegistryTemplate(registry, "wechat@2026-04-28")
	if err != nil {
		t.Fatal(err)
	}
	url := resolveRegistryAssetURL("", template.Path, template.Latest, version)
	if !strings.HasSuffix(url, "wechat/2026-04-28/template.yaml") {
		t.Fatalf("url = %q", url)
	}
	if version != template.Latest {
		t.Fatalf("version = %q, latest = %q", version, template.Latest)
	}
}

func TestStarterRulesURLPinsVersionedFolder(t *testing.T) {
	template := RegistryTemplate{
		ID:           "wechat",
		Latest:       "2026-04-28",
		StarterRules: "wechat/latest/rules.yaml",
	}
	got := resolveRegistryAssetURL("", template.StarterRules, template.Latest, "2025-07-15")
	want := registryBaseURL("") + "wechat/2025-07-15/rules.yaml"
	if got != want {
		t.Fatalf("starter rules url = %q, want %q", got, want)
	}
}

func TestLocalRegistryResolvesTemplateAndStarterRules(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "registry.yaml"), `version: 1
templates:
  - id: oklink
    latest: "2026-05-26"
    versions:
      - "2026-05-26"
    path: oklink/latest/template.yaml
    starterRules: oklink/latest/rules.yaml
`)
	writeFile(t, filepath.Join(root, "oklink", "2026-05-26", "template.yaml"), `schema: https://deg.dev/template-profile/v1
id: oklink
template:
  fileFormat: csv
  columns:
    date: time
    amount: amount
`)
	writeFile(t, filepath.Join(root, "oklink", "2026-05-26", "rules.yaml"), `personalRules: []`)

	t.Setenv(RegistryURLEnv, filepath.Join(root, "registry.yaml"))
	profileRef, err := TemplateURLFromRegistry("oklink@2026-05-26")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "oklink", "2026-05-26", "template.yaml"); profileRef != want {
		t.Fatalf("profile ref = %q, want %q", profileRef, want)
	}
	rulesRef, err := StarterRulesURLFromRegistry("oklink@2026-05-26")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "oklink", "2026-05-26", "rules.yaml"); rulesRef != want {
		t.Fatalf("rules ref = %q, want %q", rulesRef, want)
	}
	if _, err := LoadProfileRef("oklink@2026-05-26"); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
