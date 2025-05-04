package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"yadro.com/course/frontend/core"
)

type Client struct {
	log        *slog.Logger
	apiAddress string
	client     http.Client
}

func NewClient(apiAddress string, log *slog.Logger) *Client {
	return &Client{
		log:        log,
		apiAddress: apiAddress,
		client:     *http.DefaultClient,
	}
}

func (c Client) Search(phrase string) (core.SearchResponse, error) {
	searchURL := fmt.Sprintf("http://%s/api/search?phrase=%s", c.apiAddress, phrase)
	c.log.Debug("API request", "url", searchURL)

	resp, err := c.client.Get(searchURL)
	if err != nil {
		c.log.Error("failed to search", "error", err)
		return core.SearchResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("failed to search", "status", resp.StatusCode)
		return core.SearchResponse{}, err
	}

	var result core.SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.log.Error("failed to decode API response", "error", err)
		return core.SearchResponse{}, err
	}

	return result, nil
}

func (c Client) Update(token string) error {
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/api/db/update", c.apiAddress), nil)
	req.Header.Set("Authorization", "Token "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("update failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("update failed", "status", resp.StatusCode)
		return err
	}

	return nil
}

func (c Client) Drop(token string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://%s/api/db", c.apiAddress), nil)
	req.Header.Set("Authorization", "Token "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("drop failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("drop failed", "status", resp.StatusCode)
		return err
	}

	return nil
}

func (c Client) Login(username, password string) (string, error) {
	jsonBody := map[string]string{
		"name":     username,
		"password": password,
	}
	jsonData, _ := json.Marshal(jsonBody)

	resp, err := c.client.Post(
		fmt.Sprintf("http://%s/api/login", c.apiAddress),
		"application/json",
		strings.NewReader(string(jsonData)),
	)

	if err != nil {
		c.log.Error("failed to login", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Warn("failed to login", "status", resp.StatusCode)
		return "", err
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to get token", "error", err)
		return "", err
	}
	return string(token), nil
}

func (c Client) GetStatus() (core.Status, error) {
	statusReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/db/status", c.apiAddress), nil)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		c.log.Error("failed to get status", "error", err)
		return core.Status{}, err
	}
	defer statusResp.Body.Close()

	var status core.Status
	if err := json.NewDecoder(statusResp.Body).Decode(&status); err != nil {
		c.log.Error("failed to decode status", "error", err)
		return core.Status{}, err
	}
	return status, nil
}

func (c Client) GetStats() (core.Stats, error) {
	statsReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/db/stats", c.apiAddress), nil)
	statsResp, err := c.client.Do(statsReq)
	if err != nil {
		c.log.Error("failed to get stats", "error", err)
		return core.Stats{}, err
	}
	defer statsResp.Body.Close()

	var stats core.Stats
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		c.log.Error("failed to decode stats", "error", err)
		return core.Stats{}, err
	}

	return stats, nil
}
