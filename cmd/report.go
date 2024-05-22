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
		filePath := "cmd/default/dashboard-template.json" // the path to default dashboard template file.
		username := "admin"
		password := "admin"

		// Prepare for authentication
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		// Read dashboard template from JSON file
		dashboardData, err := readJSONFile(filePath)
		if err != nil {
			return err
		}

		// Create Dashboard
		resp, err := createDashboard(grafanaURL, auth, dashboardData)
		if err != nil {
			return err
		}

		// Fetch the dashboard
		dashboardData, err = fetchDashboard(grafanaURL, auth, resp.UID)
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

type DashboardResponse struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func readJSONFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func createDashboard(grafanaURL, auth string, dashboardData []byte) (DashboardResponse, error) {
	var response DashboardResponse
	createURL := grafanaURL + "/api/dashboards/db"
	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(dashboardData))
	if err != nil {
		return response, err
	}
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response, fmt.Errorf("error parsing JSON response: %v", err)
	}

	return response, nil
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
