package cmd

import (
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reporting dashboard snapshot url",
	Long:  "The report command generates a url for user to open and see the snapshot of the reporting dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		print("dummy-url")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
