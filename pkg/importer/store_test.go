package importer

import "testing"

func TestParseTemplateRef(t *testing.T) {
	tests := []struct {
		ref, wantID, wantVersion string
	}{
		{ref: "wechat", wantID: "wechat"},
		{ref: "wechat@2026.05", wantID: "wechat", wantVersion: "2026.05"},
		{ref: " wechat @ 2025.12 ", wantID: "wechat", wantVersion: "2025.12"},
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
		{path: "templates/wechat/2026.05.yaml", latest: "2026.05", version: "", want: "templates/wechat/2026.05.yaml"},
		{path: "templates/wechat/2026.05.yaml", latest: "2026.05", version: "2025.12", want: "templates/wechat/2025.12.yaml"},
		{path: "templates/wechat/latest.yaml", latest: "2026.05", version: "2025.12", want: "templates/wechat/2025.12.yaml"},
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
			Latest: "2026.05",
			Path:   "templates/wechat/2026.05.yaml",
		}},
	}
	template, version, err := lookupRegistryTemplate(registry, "wechat@2025.12")
	if err != nil {
		t.Fatal(err)
	}
	if template.ID != "wechat" || version != "2025.12" {
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
			Latest: "2026.05",
			Path:   "templates/wechat/2026.05.yaml",
		}},
	}
	template, version, err := lookupRegistryTemplate(registry, "wechat@2025.12")
	if err != nil {
		t.Fatal(err)
	}
	url := resolveRegistryAssetURL("", template.Path, template.Latest, version)
	want := registryBaseURL("") + "templates/wechat/2025.12.yaml"
	if url != want {
		t.Fatalf("url = %q, want %q", url, want)
	}
}
