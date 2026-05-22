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
	var profile Profile
	if err := yaml.Unmarshal(b, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
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

func TemplateURLFromRegistry(id string) (string, error) {
	registry, err := LoadRemoteRegistry("")
	if err != nil {
		return "", err
	}
	name := id
	version := ""
	if base, v, ok := strings.Cut(id, "@"); ok {
		name = base
		version = v
	}
	for _, template := range registry.Templates {
		if template.ID != name {
			continue
		}
		if template.URL != "" {
			return template.URL, nil
		}
		path := template.Path
		if version != "" {
			path = strings.Replace(path, template.Latest, version, 1)
		}
		return strings.TrimSuffix(DefaultRegistryURL, "registry.yaml") + path, nil
	}
	return "", fmt.Errorf("template %q not found in registry", id)
}

func StarterRulesURLFromRegistry(id string) (string, error) {
	registry, err := LoadRemoteRegistry("")
	if err != nil {
		return "", err
	}
	name := id
	if base, _, ok := strings.Cut(id, "@"); ok {
		name = base
	}
	for _, template := range registry.Templates {
		if template.ID != name {
			continue
		}
		if template.StarterRules == "" {
			return "", fmt.Errorf("template %q has no starterRules in registry", id)
		}
		if IsHTTPURL(template.StarterRules) {
			return template.StarterRules, nil
		}
		return strings.TrimSuffix(DefaultRegistryURL, "registry.yaml") + template.StarterRules, nil
	}
	return "", fmt.Errorf("template %q not found in registry", id)
}
