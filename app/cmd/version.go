package cmd

import (
	"fmt"

	"github.com/chrisyxlee/snippets/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Summary())
	},
}
