package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/spf13/cobra"
)

var (
	configInitOutput string
	configInitForce  bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage DEG local config / 管理 DEG 本地配置",
}

var configInitCmd = &cobra.Command{
	Use:   "init <template>",
	Short: "Create a personal rule skeleton / 生成个人规则骨架",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg, received %d", len(args))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		output, err := initPersonalRules(args[0], configInitOutput, configInitForce)
		logErrorIfNotNil(err)
		fmt.Printf("personal rules written: %s\n", output)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configInitCmd.Flags().StringVarP(&configInitOutput, "output", "o", "", "output personal rules YAML path")
	configInitCmd.Flags().BoolVarP(&configInitForce, "force", "f", false, "overwrite existing output file")
}

func initPersonalRules(templateRef, output string, force bool) (string, error) {
	if strings.TrimSpace(templateRef) == "" {
		return "", fmt.Errorf("template is required")
	}
	rulesURL, err := importer.StarterRulesURLFromRegistry(templateRef)
	if err != nil {
		return "", err
	}
	b, err := importer.ReadURL(rulesURL)
	if err != nil {
		return "", err
	}
	if output == "" {
		name := templateRef
		if base, _, ok := strings.Cut(templateRef, "@"); ok {
			name = base
		}
		output = name + "-rules.yaml"
	}
	if filepath.Ext(output) != ".yaml" && filepath.Ext(output) != ".yml" {
		return "", fmt.Errorf("output must be a YAML file")
	}
	if !force {
		if _, err := os.Stat(output); err == nil {
			return "", fmt.Errorf("output already exists: %s; use --force to overwrite", output)
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(output, b, 0o644); err != nil {
		return "", err
	}
	return output, nil
}
