package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reporting dashboard snapshot url",
	Long:  "The report command generates a url for user to open and see the snapshot of the reporting dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaURL := "http://localhost:3000"
		dashboardID := "admef4ycri77ka"
		username := "admin"
		password := "admin"

		// Prepare for authentication
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		// Fetch the dashboard
		dashboardData, err := fetchDashboard(grafanaURL, auth, dashboardID)
		if err != nil {
			return err
		}

		// Create a snapshot
		reportUrl, err := createSnapshot(grafanaURL, auth, dashboardData)
		if err != nil {
			return err
		}

		fmt.Println(reportUrl)
		return nil
	},
}

func fetchDashboard(grafanaURL, auth, dashboardID string) ([]byte, error) {
	// dashboardURL := fmt.Sprintf("%s/api/dashboards/db/%s", grafanaURL, dashboardName)
	dashboardURL := fmt.Sprintf("%s/api/dashboards/uid/%s", grafanaURL, dashboardID)

	req, _ := http.NewRequest("GET", dashboardURL, nil)
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func createSnapshot(grafanaURL, auth string, dashboardData []byte) (string, error) {
	snapshotURL := grafanaURL + "/api/snapshots"
	req, _ := http.NewRequest("POST", snapshotURL, bytes.NewBuffer(dashboardData))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %v", err)
	}
	if result.URL == "" {
		return "", fmt.Errorf("snapshot URL not found in response")
	}
	return result.URL, nil
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
