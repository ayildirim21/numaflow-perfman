package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ayildirim21/numaflow-perfman/report"
	"github.com/ayildirim21/numaflow-perfman/util"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reporting dashboard snapshot url",
	Long:  "The report command generates a url for user to open and see the snapshot of the reporting dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: can all be moved to viper configuration
		grafanaURL := "http://localhost:3000"
		filePath := "default/dashboard-template.json" // the path to default dashboard template file.
		username := "admin"
		password := util.GrafanaPassword

		// Prepare for authentication
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		// Create the Prometheus data source
		dsId, err := report.CreateGrafanaDataSource(grafanaURL, auth)
		if err != nil {
			if strings.Contains(err.Error(), "data source with the same name already exists") {
				log.Warn("Prometheus data source has already been configured.")
				// Attempt to fetch the UID of the existing data source
				dsId, err = report.FetchGrafanaDataSourceUID(grafanaURL, auth)
				if err != nil {
					return fmt.Errorf("error fetching existing data source UID: %w", err)
				}
			} else {
				return fmt.Errorf("error creating data source: %w", err)
			}
		}

		// Read dashboard template from JSON file
		dashboardData, err := report.ReadJSONFile(filePath)
		if err != nil {
			return err
		}

		// Configure the dashboard template to read from the data source created above
		dashboardData = []byte(strings.Replace(string(dashboardData), "prometheus-datasource-uid-placeholder", dsId, -1))

		// Create Dashboard
		// TODO - handle case when the dashboard already exists.
		// (we should move data source and dashboard creation into setup and let report to only generate snapshot.)
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
