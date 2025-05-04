package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.url, id)
	resp, err := http.Get(url)
	if err != nil {
		c.log.Error("failed to send req", "info", err)
		return core.XKCDInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if id == 404 {
			return core.XKCDInfo{}, core.Err404Comics
		}
		return core.XKCDInfo{}, err
	}

	var jsonInfo core.JsonXKCDInfo
	if err := json.NewDecoder(resp.Body).Decode(&jsonInfo); err != nil {
		return core.XKCDInfo{}, err
	}

	info := core.XKCDInfo(jsonInfo)

	return info, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/info.0.json", c.url)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to send req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jsonInfo core.JsonXKCDInfo
	if err := json.NewDecoder(resp.Body).Decode(&jsonInfo); err != nil {
		return 0, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return jsonInfo.ID, nil
}
