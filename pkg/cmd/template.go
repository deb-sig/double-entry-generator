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
	Use:               "template",
	Short:             "Inspect runtime import templates / 查看运行时模板",
	ValidArgsFunction: cobra.NoFileCompletions,
}

var templateListCmd = &cobra.Command{
	Use:               "list",
	Short:             "List templates from the online registry / 查看线上模板列表",
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := importer.LoadRemoteRegistry(templateRegistryURL)
		logErrorIfNotNil(err)
		fmt.Println("id\tname\tlatest\tprofile_path\tdescription")
		for _, template := range registry.Templates {
			latest := template.Latest
			if latest == "" {
				latest = template.Version
			}
			pin := template.ID
			if latest != "" {
				pin = template.ID + "@" + latest
			}
			desc := template.Description
			if latest != "" {
				desc = strings.TrimSpace(desc + " | pin: " + pin)
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
				template.ID,
				template.Name,
				latest,
				template.Path,
				desc,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateListCmd.Flags().StringVar(&templateRegistryURL, "registry", "", "registry URL / 模板索引地址")
}
