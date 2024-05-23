package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DashboardResponse struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func ReadJSONFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func CreateDashboard(grafanaURL, auth string, dashboardData []byte) (DashboardResponse, error) {
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

func FetchDashboard(grafanaURL, auth, dashboardID string) ([]byte, error) {
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

func CreateSnapshot(grafanaURL, auth string, dashboardData []byte) (string, error) {
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
