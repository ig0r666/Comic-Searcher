package xkcd

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/update/core"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://xkcd.com",
			timeout: time.Second,
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			timeout: time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.url, tt.timeout, slog.Default())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/777/info.0.json":
			info := core.JsonXKCDInfo{
				ID:    777,
				Title: "Pore Strips",
				URL:   "https://imgs.xkcd.com/comics/pore_strips.png",
				Alt:   "I'm sure they're a harmful tool of the cosmetics-industrial complex and all, but my goodness do those strips ever work to pull gunk out of your pores. I was shocked, disgusted, and vaguely fascinated by the result.",
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(info); err != nil {
				log.Printf("failed to encode")
				return
			}
		case "/404/info.0.json":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, time.Second, slog.Default())
	assert.NoError(t, err)

	tests := []struct {
		name    string
		id      int
		want    core.XKCDInfo
		wantErr error
	}{
		{
			name: "success get 777 comic",
			id:   777,
			want: core.XKCDInfo{
				ID:    777,
				Title: "Pore Strips",
				URL:   "https://imgs.xkcd.com/comics/pore_strips.png",
				Alt:   "I'm sure they're a harmful tool of the cosmetics-industrial complex and all, but my goodness do those strips ever work to pull gunk out of your pores. I was shocked, disgusted, and vaguely fascinated by the result.",
			},
			wantErr: nil,
		},
		{
			name:    "404 comic",
			id:      404,
			want:    core.XKCDInfo{},
			wantErr: core.Err404Comics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.Get(context.Background(), tt.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_LastID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/info.0.json":
			info := core.JsonXKCDInfo{
				ID: 3076,
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(info); err != nil {
				log.Printf("failed to encode")
				return
			}
		case "/error/info.0.json":
			w.WriteHeader(http.StatusInternalServerError)
		case "/invalidjson/info.0.json":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("invalid json")); err != nil {
				log.Printf("failed to write")
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		want    int
		wantErr string
	}{
		{
			name:    "success",
			url:     server.URL,
			want:    3076,
			wantErr: "",
		},
		{
			name:    "server err",
			url:     server.URL + "/error",
			want:    0,
			wantErr: "unexpected status code",
		},
		{
			name:    "invalid json",
			url:     server.URL + "/invalidjson",
			want:    0,
			wantErr: "failed to decode JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url, time.Second, slog.Default())
			assert.NoError(t, err)

			got, err := client.LastID(context.Background())
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
