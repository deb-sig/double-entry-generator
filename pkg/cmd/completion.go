package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|powershell]",
	Short: msg("Generate shell completion scripts", "生成 shell 补全脚本"),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg, received %d", len(args))
		}
		switch args[0] {
		case "bash", "zsh", "powershell":
			return nil
		default:
			return fmt.Errorf("unsupported shell %q", args[0])
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			err = rootCmd.GenZshCompletion(os.Stdout)
		case "powershell":
			err = rootCmd.GenPowerShellCompletion(os.Stdout)
		}
		logErrorIfNotNil(err)
	},
	ValidArgsFunction: cobra.FixedCompletions([]cobra.Completion{
		cobra.CompletionWithDesc("bash", msg("Generate bash completion", "生成 bash 补全脚本")),
		cobra.CompletionWithDesc("zsh", msg("Generate zsh completion", "生成 zsh 补全脚本")),
		cobra.CompletionWithDesc("powershell", msg("Generate PowerShell completion", "生成 PowerShell 补全脚本")),
	}, cobra.ShellCompDirectiveNoFileComp),
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func completeTemplateRefs(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if strings.Contains(toComplete, "@") {
		return completeTemplateVersions(toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	registry, err := importer.LoadRemoteRegistry("")
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	completions := make([]cobra.Completion, 0, len(registry.Templates))
	for _, template := range registry.Templates {
		desc := templateCompletionDescription(template)
		if template.Latest != "" {
			desc += " latest " + template.Latest
		}
		completions = append(completions, cobra.CompletionWithDesc(template.ID, desc))
	}
	return completions, cobra.ShellCompDirectiveDefault
}

func completeTemplateVersions(toComplete string) []cobra.Completion {
	id, prefix := importer.ParseTemplateRef(toComplete)
	if id == "" {
		return nil
	}
	registry, err := importer.LoadRemoteRegistry("")
	if err != nil {
		return nil
	}
	for _, template := range registry.Templates {
		if template.ID != id {
			continue
		}
		versions := template.Versions
		if len(versions) == 0 && template.Latest != "" {
			versions = []string{template.Latest}
		}
		out := make([]cobra.Completion, 0, len(versions))
		for _, version := range versions {
			if prefix != "" && !strings.HasPrefix(version, prefix) {
				continue
			}
			desc := templateCompletionDescription(template)
			if version == template.Latest {
				desc = strings.TrimSpace(desc + " latest")
			}
			out = append(out, cobra.CompletionWithDesc(id+"@"+version, desc))
		}
		return out
	}
	return nil
}

func completeImportArgs(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeTemplateRefs(toComplete)
	case 1:
		return []cobra.Completion{"csv", "xlsx", "xls"}, cobra.ShellCompDirectiveFilterFileExt
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeConfigInitArgs(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeTemplateRefs(toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeYAMLFiles(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
}

func completeOutputFiles(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(toComplete)), ".")
	if ext == "bean" || ext == "beancount" || ext == "ledger" {
		return nil, cobra.ShellCompDirectiveDefault
	}
	return []cobra.Completion{"bean", "beancount", "ledger"}, cobra.ShellCompDirectiveFilterFileExt
}

func templateCompletionDescription(template importer.RegistryTemplate) string {
	if isChineseLocale() {
		return firstNonEmpty(template.Name, template.Description, template.ID)
	}
	return firstNonEmpty(template.ID, template.Name, template.Description)
}
