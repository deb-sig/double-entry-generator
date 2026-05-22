package cmd

import (
	"fmt"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/spf13/cobra"
)

var (
	templateRegistryURL string
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Inspect runtime import templates / 查看运行时模板",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List templates from the online registry / 查看线上模板列表",
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := importer.LoadRemoteRegistry(templateRegistryURL)
		logErrorIfNotNil(err)
		for _, template := range registry.Templates {
			fields := []string{template.ID}
			if template.Name != "" {
				fields = append(fields, template.Name)
			}
			if template.Version != "" {
				fields = append(fields, template.Version)
			}
			if template.Category != "" {
				fields = append(fields, template.Category)
			}
			if len(template.Tags) > 0 {
				fields = append(fields, strings.Join(template.Tags, ","))
			}
			fmt.Println(strings.Join(fields, "\t"))
		}
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateListCmd.Flags().StringVar(&templateRegistryURL, "registry", "", "registry URL / 模板索引地址")
}
