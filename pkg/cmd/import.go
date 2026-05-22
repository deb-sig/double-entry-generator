package cmd

import (
	"fmt"
	"os"

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
	Short: "Import bills with a runtime template / 使用运行时模板导入账单",
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
	importCmd.Flags().StringVar(&importRules, "rules", "", "personal rules YAML file / 个人规则 YAML 文件")
	importCmd.Flags().StringVarP(&importTarget, "target", "t", "", "target output format / 输出格式")
	importCmd.Flags().StringVarP(&importOutput, "output", "o", "", "output file / 输出文件")
	importCmd.Flags().BoolVarP(&importAppend, "append", "a", false, "append mode / 追加写入")
}

func runImport(templateRef, filename string) {
	projectCfg := loadProjectImportConfig()
	profile, err := importer.LoadProfileRef(firstNonEmpty(templateRef, projectCfg.DefaultTemplate))
	logErrorIfNotNil(err)

	rulesPath := firstNonEmpty(importRules, projectCfg.DefaultRules)
	if rulesPath != "" {
		ruleCfg, err := loadRuleFile(rulesPath)
		logErrorIfNotNil(err)
		profile.TemplateRuleOverrides = append(profile.TemplateRuleOverrides, ruleCfg.TemplateRuleOverrides...)
		profile.PersonalRules = append(profile.PersonalRules, ruleCfg.Rules...)
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
	Rules                 []importer.Rule `yaml:"rules"`
	TemplateRuleOverrides []importer.Rule `yaml:"templateRuleOverrides"`
	PersonalRules         []importer.Rule `yaml:"personalRules"`
	UserRules             []importer.Rule `yaml:"userRules"`
}

func loadRuleFile(path string) (importRuleConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return importRuleConfig{}, err
	}
	var rules []importer.Rule
	if err := yaml.Unmarshal(b, &rules); err == nil && len(rules) > 0 {
		return importRuleConfig{Rules: rules}, nil
	}
	var wrapper importRuleConfig
	if err := yaml.Unmarshal(b, &wrapper); err != nil {
		return importRuleConfig{}, err
	}
	wrapper.Rules = append(wrapper.Rules, wrapper.PersonalRules...)
	wrapper.Rules = append(wrapper.Rules, wrapper.UserRules...)
	return wrapper, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
