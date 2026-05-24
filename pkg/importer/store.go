package importer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultRegistryURL = "https://raw.githubusercontent.com/deb-sig/deg-provider-template/main/registry.yaml"

type Registry struct {
	Version   int                `json:"version" yaml:"version"`
	Templates []RegistryTemplate `json:"templates" yaml:"templates"`
}

type RegistryTemplate struct {
	ID           string   `json:"id" yaml:"id"`
	Name         string   `json:"name,omitempty" yaml:"name,omitempty"`
	Category     string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Version      string   `json:"version,omitempty" yaml:"version,omitempty"`
	Latest       string   `json:"latest,omitempty" yaml:"latest,omitempty"`
	Versions     []string `json:"versions,omitempty" yaml:"versions,omitempty"`
	Path         string   `json:"path" yaml:"path"`
	StarterRules string   `json:"starterRules,omitempty" yaml:"starterRules,omitempty"`
	URL          string   `json:"url,omitempty" yaml:"url,omitempty"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	SHA256       string   `json:"sha256,omitempty" yaml:"sha256,omitempty"`
}

func LoadProfileRef(ref string) (*Profile, error) {
	if ref == "" {
		return nil, fmt.Errorf("template is required")
	}
	if strings.HasPrefix(ref, "ftp://") {
		return nil, fmt.Errorf("ftp template URLs are not supported")
	}
	if IsHTTPURL(ref) {
		return loadProfileURL(ref)
	}
	if IsLocalPathRef(ref) {
		return loadProfilePath(ref)
	}
	rawURL, err := TemplateURLFromRegistry(ref)
	if err != nil {
		return nil, err
	}
	return loadProfileURL(rawURL)
}

func IsHTTPURL(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

func IsLocalPathRef(ref string) bool {
	return filepath.IsAbs(ref) ||
		strings.HasPrefix(ref, "./") ||
		strings.HasPrefix(ref, "../") ||
		strings.HasPrefix(ref, "~/")
}

func loadProfilePath(path string) (*Profile, error) {
	if err := requireYAMLPath(path); err != nil {
		return nil, err
	}
	return LoadProfile(expandHome(path))
}

func loadProfileURL(rawURL string) (*Profile, error) {
	if err := requireYAMLPath(rawURL); err != nil {
		return nil, err
	}
	b, err := readURL(rawURL)
	if err != nil {
		return nil, err
	}
	return loadProfileBytes(b, strings.TrimSuffix(filepath.Base(strings.Split(rawURL, "?")[0]), filepath.Ext(strings.Split(rawURL, "?")[0])))
}

func LoadRemoteRegistry(rawURL string) (*Registry, error) {
	if rawURL == "" {
		rawURL = DefaultRegistryURL
	}
	b, err := readURL(rawURL)
	if err != nil {
		return nil, err
	}
	var registry Registry
	if err := yaml.Unmarshal(b, &registry); err != nil {
		return nil, err
	}
	return &registry, nil
}

func ReadURL(rawURL string) ([]byte, error) {
	return readURL(rawURL)
}

func readURL(rawURL string) ([]byte, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download %s: %s", rawURL, resp.Status)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func requireYAMLPath(path string) error {
	ext := strings.ToLower(filepath.Ext(strings.Split(path, "?")[0]))
	if !slices.Contains([]string{".yaml", ".yml"}, ext) {
		return fmt.Errorf("template profile must be a YAML file")
	}
	return nil
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

// ParseTemplateRef splits "wechat@2026-04-28" into id and optional pinned version.
func ParseTemplateRef(ref string) (id, version string) {
	ref = strings.TrimSpace(ref)
	if base, v, ok := strings.Cut(ref, "@"); ok {
		return strings.TrimSpace(base), strings.TrimSpace(v)
	}
	return ref, ""
}

func registryBaseURL(rawURL string) string {
	if rawURL == "" {
		rawURL = DefaultRegistryURL
	}
	return strings.TrimSuffix(rawURL, "registry.yaml")
}

func lookupRegistryTemplate(registry *Registry, ref string) (RegistryTemplate, string, error) {
	id, version := ParseTemplateRef(ref)
	if id == "" {
		return RegistryTemplate{}, "", fmt.Errorf("template id is required")
	}
	if strings.Contains(ref, "@") && version == "" {
		return RegistryTemplate{}, "", fmt.Errorf("template version is empty in %q, use id@version (e.g. wechat@2026-04-28)", ref)
	}
	for _, template := range registry.Templates {
		if template.ID != id {
			continue
		}
		return template, version, nil
	}
	return RegistryTemplate{}, version, fmt.Errorf("template %q not found in registry", id)
}

func applyVersionToPath(path, latest, version string) string {
	if version == "" {
		return path
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		if part == "latest" {
			out := append([]string(nil), parts...)
			out[i] = version
			return strings.Join(out, "/")
		}
	}
	if latest != "" && strings.Contains(path, latest) {
		return strings.Replace(path, latest, version, 1)
	}
	dir, file := filepath.Split(path)
	ext := filepath.Ext(file)
	return filepath.ToSlash(filepath.Join(strings.TrimSuffix(dir, "/"), version+ext))
}

func resolveRegistryAssetURL(registryURL, assetPath, latest, version string) string {
	return registryBaseURL(registryURL) + applyVersionToPath(assetPath, latest, version)
}

func TemplateURLFromRegistry(id string) (string, error) {
	registry, err := LoadRemoteRegistry("")
	if err != nil {
		return "", err
	}
	template, version, err := lookupRegistryTemplate(registry, id)
	if err != nil {
		return "", err
	}
	if version != "" && template.URL != "" {
		return "", fmt.Errorf("template %q uses a fixed URL and does not support version pinning; use a path-based registry entry or omit @version", template.ID)
	}
	if template.URL != "" {
		return template.URL, nil
	}
	if template.Path == "" {
		return "", fmt.Errorf("template %q has no path in registry", template.ID)
	}
	url := resolveRegistryAssetURL("", template.Path, template.Latest, version)
	if version != "" && version != template.Latest && template.Latest != "" && strings.Contains(url, template.Latest) {
		return "", fmt.Errorf("template %s@%s: version %q did not match registry path (latest=%q)", template.ID, version, version, template.Latest)
	}
	return url, nil
}

func StarterRulesURLFromRegistry(id string) (string, error) {
	registry, err := LoadRemoteRegistry("")
	if err != nil {
		return "", err
	}
	template, version, err := lookupRegistryTemplate(registry, id)
	if err != nil {
		return "", err
	}
	if template.StarterRules == "" {
		return "", fmt.Errorf("template %q has no starterRules in registry", template.ID)
	}
	if IsHTTPURL(template.StarterRules) {
		if version != "" {
			return "", fmt.Errorf("template %q starter rules use a fixed URL and do not support @version", template.ID)
		}
		return template.StarterRules, nil
	}
	return resolveRegistryAssetURL("", template.StarterRules, template.Latest, version), nil
}
