package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/spf13/cobra"
)

var (
	templateRegistryURL string
)

var templateCmd = &cobra.Command{
	Use:               "template",
	Short:             msg("Inspect runtime import templates", "查看运行时模板"),
	ValidArgsFunction: cobra.NoFileCompletions,
}

var templateListCmd = &cobra.Command{
	Use:               "list",
	Short:             msg("List templates from the online registry", "查看线上模板列表"),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := importer.LoadRemoteRegistry(templateRegistryURL)
		logErrorIfNotNil(err)
		printTemplateGroups(registry.Templates)
	},
}

var templateSearchCmd = &cobra.Command{
	Use:               "search <query>",
	Short:             msg("Search templates in the online registry", "搜索线上模板"),
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := importer.LoadRemoteRegistry(templateRegistryURL)
		logErrorIfNotNil(err)
		matches := searchTemplates(registry.Templates, args[0])
		if len(matches) == 0 {
			fmt.Printf("No templates matched %q\n", args[0])
			return
		}
		if len(matches) > 1 {
			printTemplateGroups(matches)
			return
		}
		printTemplateDetail(matches[0])
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateSearchCmd)
	templateListCmd.Flags().StringVar(&templateRegistryURL, "registry", "", msg("registry URL", "模板索引地址"))
	templateSearchCmd.Flags().StringVar(&templateRegistryURL, "registry", "", msg("registry URL", "模板索引地址"))
}

func printTemplateGroups(templates []importer.RegistryTemplate) {
	byCategory := map[string][]importer.RegistryTemplate{}
	for _, template := range templates {
		category := firstNonEmpty(template.Category, "uncategorized")
		byCategory[category] = append(byCategory[category], template)
	}
	categories := make([]string, 0, len(byCategory))
	for category := range byCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	for i, category := range categories {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(category)
		sort.Slice(byCategory[category], func(i, j int) bool {
			return byCategory[category][i].ID < byCategory[category][j].ID
		})
		for _, template := range byCategory[category] {
			parts := []string{template.ID}
			if template.Name != "" {
				parts = append(parts, template.Name)
			}
			if latest := templateLatest(template); latest != "" {
				parts = append(parts, "latest "+latest)
			}
			if len(template.Tags) > 0 {
				parts = append(parts, "#"+strings.Join(template.Tags, " #"))
			}
			fmt.Printf("  - %s\n", strings.Join(parts, "  "))
		}
	}
}

func searchTemplates(templates []importer.RegistryTemplate, query string) []importer.RegistryTemplate {
	query = strings.ToLower(strings.TrimSpace(query))
	matches := make([]importer.RegistryTemplate, 0)
	for _, template := range templates {
		haystack := strings.ToLower(strings.Join([]string{
			template.ID,
			template.Name,
			template.Category,
			strings.Join(template.Tags, " "),
			template.Description,
		}, " "))
		if strings.Contains(haystack, query) {
			matches = append(matches, template)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ID < matches[j].ID
	})
	return matches
}

func printTemplateDetail(template importer.RegistryTemplate) {
	fmt.Println(template.ID)
	if template.Name != "" {
		fmt.Printf("  name: %s\n", template.Name)
	}
	if template.Category != "" {
		fmt.Printf("  category: %s\n", template.Category)
	}
	if len(template.Tags) > 0 {
		fmt.Printf("  tags: %s\n", strings.Join(template.Tags, ", "))
	}
	if template.Description != "" {
		fmt.Printf("  description: %s\n", template.Description)
	}
	versions := template.Versions
	if len(versions) == 0 {
		versions = []string{templateLatest(template)}
	}
	for _, version := range versions {
		if version == "" {
			continue
		}
		ref := template.ID + "@" + version
		format := templateFormatForRef(ref)
		suffix := ""
		if version == templateLatest(template) {
			suffix = " latest"
		}
		if format != "" {
			fmt.Printf("  - @%s  format:%s%s\n", version, format, suffix)
		} else {
			fmt.Printf("  - @%s%s\n", version, suffix)
		}
	}
}

func templateLatest(template importer.RegistryTemplate) string {
	return firstNonEmpty(template.Latest, template.Version)
}

func templateFormatForRef(ref string) string {
	profile, err := importer.LoadProfileRef(ref)
	if err != nil {
		return ""
	}
	return firstNonEmpty(profile.Template.FileFormat, "csv")
}
