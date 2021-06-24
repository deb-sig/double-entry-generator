package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/deb-sig/double-entry-generator/pkg/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of double-entry-generator",
	Long:  `All software has versions. This is double-entry-generator's`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("double-entry-generator Version: %s, Commit: %s, Build Path: %s",
			version.VERSION, version.COMMIT, version.REPOROOT)
	},
}
