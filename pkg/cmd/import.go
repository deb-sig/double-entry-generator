package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	importRules  string
	importTarget string
	importOutput string
	importAppend bool
)

type projectImportConfig struct {
	DefaultTemplate string `yaml:"defaultTemplate"`
	DefaultRules    string `yaml:"defaultRules"`
	DefaultTarget   string `yaml:"defaultTarget"`
	DefaultOutput   string `yaml:"defaultOutput"`
	Append          bool   `yaml:"append"`
}

var importCmd = &cobra.Command{
	Use:   "import [flags] <template> <path to bill file>",
	Short: msg("Import bills with a runtime template", "使用运行时模板导入账单"),
	Long: strings.TrimSpace(`
Import bills using a template from the online registry, a pinned version, or a local profile YAML.

Template reference:
  wechat                         use registry latest
  wechat@2026-04-28              pin a specific profile version
  ./profiles/wechat.yaml           local profile file

Omit @version to follow registry "latest"; pin @version to keep old bill formats working after the default template changes.

Examples:
  double-entry-generator import wechat bill.csv -o out.bean
  double-entry-generator import wechat@2025-07-15 bill.csv --rules rules.yaml -o out.bean
`),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("accepts 2 args, received %d", len(args))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		runImport(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importRules, "rules", "", msg("personal rules YAML file", "个人规则 YAML 文件"))
	importCmd.Flags().StringVarP(&importTarget, "target", "t", "", msg("target output format", "输出格式"))
	importCmd.Flags().StringVarP(&importOutput, "output", "o", "", msg("output file", "输出文件"))
	importCmd.Flags().BoolVarP(&importAppend, "append", "a", false, msg("append to output file", "追加写入输出文件"))
}

func runImport(templateRef, filename string) {
	projectCfg := loadProjectImportConfig()
	templateRef = firstNonEmpty(templateRef, projectCfg.DefaultTemplate)
	if id, version := importer.ParseTemplateRef(templateRef); version != "" {
		log.Printf("Using pinned template %s@%s", id, version)
	}
	profile, err := importer.LoadProfileRef(templateRef)
	logErrorIfNotNil(err)

	if isRegistryTemplateRef(templateRef) {
		ruleCfg, err := loadRegistryStarterRules(templateRef)
		logErrorIfNotNil(err)
		appendTemplateRulesToProfile(profile, ruleCfg)
	}

	rulesPath := firstNonEmpty(importRules, projectCfg.DefaultRules)
	if rulesPath != "" {
		ruleCfg, err := loadRuleFile(rulesPath)
		logErrorIfNotNil(err)
		appendRulesToProfile(profile, ruleCfg)
	}

	i, err := importer.ImportFile(profile, filename)
	logErrorIfNotNil(err)

	c := &config.Config{
		Title:               firstNonEmpty(profile.Name, profile.ID, "DEG Import"),
		DefaultMinusAccount: profile.Template.DefaultMinus,
		DefaultPlusAccount:  profile.Template.DefaultPlus,
		DefaultCurrency:     firstNonEmpty(profile.Template.DefaultCurrency, "CNY"),
	}
	target := firstNonEmpty(importTarget, projectCfg.DefaultTarget, consts.CompilerBeanCount)
	output := firstNonEmpty(importOutput, projectCfg.DefaultOutput, "default_output.beancount")
	appendMode := importAppend || projectCfg.Append

	cpl, err := compiler.New(importer.DefaultProviderName, target, output, appendMode, c, i)
	logErrorIfNotNil(err)
	logErrorIfNotNil(cpl.Compile())
}

func loadProjectImportConfig() projectImportConfig {
	var cfg projectImportConfig
	b, err := os.ReadFile(".deg/config.yaml")
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(b, &cfg)
	return cfg
}

type importRuleConfig struct {
	TemplateRules         []importer.Rule `yaml:"templateRules"`
	TemplateRuleOverrides []importer.Rule `yaml:"templateRuleOverrides"`
	PersonalRules         []importer.Rule `yaml:"personalRules"`
	Options               importOptions   `yaml:"options"`
}

type importOptions struct {
	Title             string `yaml:"title"`
	OperatingCurrency string `yaml:"operatingCurrency"`
}

func loadRuleFile(path string) (importRuleConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return importRuleConfig{}, err
	}
	return parseRuleBytes(b)
}

func loadRegistryStarterRules(templateRef string) (importRuleConfig, error) {
	rulesRef, err := importer.StarterRulesURLFromRegistry(templateRef)
	if err != nil {
		return importRuleConfig{}, err
	}
	b, err := importer.ReadRef(rulesRef)
	if err != nil {
		return importRuleConfig{}, err
	}
	return parseRuleBytes(b)
}

func parseRuleBytes(b []byte) (importRuleConfig, error) {
	var wrapper importRuleConfig
	if err := yaml.Unmarshal(b, &wrapper); err != nil {
		return importRuleConfig{}, err
	}
	return wrapper, nil
}

func appendRulesToProfile(profile *importer.Profile, ruleCfg importRuleConfig) {
	profile.TemplateRules = append(profile.TemplateRules, ruleCfg.TemplateRules...)
	profile.TemplateRuleOverrides = append(profile.TemplateRuleOverrides, ruleCfg.TemplateRuleOverrides...)
	profile.PersonalRules = append(profile.PersonalRules, ruleCfg.PersonalRules...)
	if ruleCfg.Options.Title != "" {
		profile.Name = ruleCfg.Options.Title
	}
	if ruleCfg.Options.OperatingCurrency != "" {
		profile.Template.DefaultCurrency = ruleCfg.Options.OperatingCurrency
	}
}

func appendTemplateRulesToProfile(profile *importer.Profile, ruleCfg importRuleConfig) {
	profile.TemplateRules = append(profile.TemplateRules, ruleCfg.TemplateRules...)
	if ruleCfg.Options.Title != "" {
		profile.Name = ruleCfg.Options.Title
	}
	if ruleCfg.Options.OperatingCurrency != "" {
		profile.Template.DefaultCurrency = ruleCfg.Options.OperatingCurrency
	}
}

func isRegistryTemplateRef(ref string) bool {
	return !importer.IsHTTPURL(ref) && !importer.IsLocalPathRef(ref)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
