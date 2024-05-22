package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayildirim21/numaflow-perfman/report"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reporting dashboard snapshot url",
	Long:  "The report command generates a url for user to open and see the snapshot of the reporting dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaURL := "http://localhost:3000"
		filePath := "cmd/default/dashboard-template.json" // the path to default dashboard template file.
		username := "admin"
		password := "admin"

		// Prepare for authentication
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		// Read dashboard template from JSON file
		dashboardData, err := report.ReadJSONFile(filePath)
		if err != nil {
			return err
		}

		// Create Dashboard
		resp, err := report.CreateDashboard(grafanaURL, auth, dashboardData)
		if err != nil {
			return err
		}

		// Fetch the dashboard
		dashboardData, err = report.FetchDashboard(grafanaURL, auth, resp.UID)
		if err != nil {
			return err
		}

		// Create a snapshot
		reportUrl, err := report.CreateSnapshot(grafanaURL, auth, dashboardData)
		if err != nil {
			return err
		}

		fmt.Println(reportUrl)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
