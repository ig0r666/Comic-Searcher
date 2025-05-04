package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"yadro.com/course/api/core"
)

type MockLoginer struct{ mock.Mock }

func (m *MockLoginer) Login(name, password string) (string, error) {
	args := m.Called(name, password)
	return args.String(0), args.Error(1)
}

type MockPinger struct{ mock.Mock }

func (m *MockPinger) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockUpdater struct{ mock.Mock }

func (m *MockUpdater) Update(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockUpdater) Stats(ctx context.Context) (core.UpdateStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(core.UpdateStats), args.Error(1)
}
func (m *MockUpdater) Status(ctx context.Context) (core.UpdateStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(core.UpdateStatus), args.Error(1)
}
func (m *MockUpdater) Drop(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockSearcher struct{ mock.Mock }

func (m *MockSearcher) Search(ctx context.Context, limit int, phrase string) ([]core.Comics, error) {
	args := m.Called(ctx, limit, phrase)
	return args.Get(0).([]core.Comics), args.Error(1)
}
func (m *MockSearcher) IndexSearch(ctx context.Context, limit int, phrase string) ([]core.Comics, error) {
	args := m.Called(ctx, limit, phrase)
	return args.Get(0).([]core.Comics), args.Error(1)
}

type MockTokenVerifier struct{ mock.Mock }

func (m *MockTokenVerifier) Verify(token string) error {
	return m.Called(token).Error(0)
}

func TestNewLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]string
		mockToken  string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "successful login",
			input:      map[string]string{"name": "admin", "password": "password"},
			mockToken:  "token",
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid creds",
			input:      map[string]string{"name": "admin", "password": "wrong"},
			mockToken:  "",
			mockErr:    errors.New("invalid"),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoginer := &MockLoginer{}
			mockLoginer.On("Login", tt.input["name"], tt.input["password"]).Return(tt.mockToken, tt.mockErr)

			handler := NewLoginHandler(slog.Default(), mockLoginer)

			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.input); err != nil {
				log.Printf("Failed to encode")
				return
			}

			req := httptest.NewRequest("POST", "/api/login", &body)
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.mockToken, w.Body.String())
			}
			mockLoginer.AssertExpectations(t)
		})
	}
}

func TestNewPingHandler(t *testing.T) {
	mockPinger := &MockPinger{}
	mockPinger.On("Ping", mock.Anything).Return(nil)

	pingers := map[string]core.Pinger{
		"test": mockPinger,
	}

	handler := NewPingHandler(slog.Default(), pingers)

	req := httptest.NewRequest("GET", "/api/ping", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		log.Printf("Failed to encode")
		return
	}
	assert.Equal(t, "ok", response["replies"].(map[string]interface{})["test"])
	mockPinger.AssertExpectations(t)
}
func TestNewUpdateHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockUpdate error
		wantStatus int
		authHeader string
		mockVerify error
	}{
		{
			name:       "successful update",
			mockUpdate: nil,
			wantStatus: http.StatusOK,
			authHeader: "Token valid",
			mockVerify: nil,
		},
		{
			name:       "update in progress",
			mockUpdate: errors.New("already exists"),
			wantStatus: http.StatusAccepted,
			authHeader: "Token valid",
			mockVerify: nil,
		},
		{
			name:       "no token",
			mockUpdate: nil,
			wantStatus: http.StatusUnauthorized,
			authHeader: "",
			mockVerify: nil,
		},
		{
			name:       "invalid token",
			mockUpdate: nil,
			wantStatus: http.StatusUnauthorized,
			authHeader: "Token invalid",
			mockVerify: errors.New("invalid token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdater := &MockUpdater{}
			if tt.authHeader != "" && (tt.mockVerify == nil || tt.authHeader == "Token valid") {
				mockUpdater.On("Update", mock.Anything).Return(tt.mockUpdate)
			}

			mockVerifier := &MockTokenVerifier{}
			if tt.authHeader != "" {
				token := tt.authHeader[len("Token "):]
				mockVerifier.On("Verify", token).Return(tt.mockVerify)
			}

			handler := NewUpdateHandler(slog.Default(), mockUpdater, mockVerifier)

			req := httptest.NewRequest("POST", "/api/db/update", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus != http.StatusUnauthorized {
				mockUpdater.AssertExpectations(t)
			}
			if tt.authHeader != "" {
				mockVerifier.AssertExpectations(t)
			}
		})
	}
}

func TestNewDropHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockDrop   error
		wantStatus int
		authHeader string
		mockVerify error
	}{
		{
			name:       "successful drop",
			mockDrop:   nil,
			wantStatus: http.StatusOK,
			authHeader: "Token valid",
			mockVerify: nil,
		},
		{
			name:       "drop error",
			mockDrop:   errors.New("drop failed"),
			wantStatus: http.StatusInternalServerError,
			authHeader: "Token valid",
			mockVerify: nil,
		},
		{
			name:       "Unauthorized",
			mockDrop:   nil,
			wantStatus: http.StatusUnauthorized,
			authHeader: "Token invalid",
			mockVerify: errors.New("invalid token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdater := &MockUpdater{}
			if tt.authHeader == "Token valid" {
				mockUpdater.On("Drop", mock.Anything).Return(tt.mockDrop)
			}

			mockVerifier := &MockTokenVerifier{}
			if tt.authHeader != "" {
				token := tt.authHeader[len("Token "):]
				mockVerifier.On("Verify", token).Return(tt.mockVerify)
			}

			handler := NewDropHandler(slog.Default(), mockUpdater, mockVerifier)

			req := httptest.NewRequest("DELETE", "/api/db", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.authHeader == "Token valid" {
				mockUpdater.AssertExpectations(t)
			}
			if tt.authHeader != "" {
				mockVerifier.AssertExpectations(t)
			}
		})
	}
}

func TestNewSearchHandler(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		mockComics  []core.Comics
		mockErr     error
		wantStatus  int
	}{
		{
			name: "successful search",
			queryParams: map[string]string{
				"phrase": "Binary Christmas Tree",
				"limit":  "1",
			},
			mockComics: []core.Comics{
				{ID: 1, URL: "https://imgs.xkcd.com/comics/tree.png"},
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name: "miss phrase",
			queryParams: map[string]string{
				"limit": "1",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "invalid",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "search error",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "1",
			},
			mockComics: nil,
			mockErr:    errors.New("search error"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSearcher := &MockSearcher{}
			if phrase, ok := tt.queryParams["phrase"]; ok && phrase != "" {
				limit := 10
				if l, ok := tt.queryParams["limit"]; ok {
					if num, err := strconv.Atoi(l); err == nil {
						limit = num
					}
				}
				if _, err := strconv.Atoi(tt.queryParams["limit"]); err == nil || tt.queryParams["limit"] == "" {
					mockSearcher.On("Search", mock.Anything, limit, phrase).Return(tt.mockComics, tt.mockErr)
				}
			}

			handler := NewSearchHandler(slog.Default(), mockSearcher, 10)

			req := httptest.NewRequest("GET", "/api/search", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockComics), int(response["total"].(float64)))
			}

			if phrase, ok := tt.queryParams["phrase"]; ok && phrase != "" {
				if _, err := strconv.Atoi(tt.queryParams["limit"]); err == nil || tt.queryParams["limit"] == "" {
					mockSearcher.AssertExpectations(t)
				}
			}
		})
	}
}

func TestNewIndexSearchHandler(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		mockComics  []core.Comics
		mockErr     error
		wantStatus  int
	}{
		{
			name: "successful index search",
			queryParams: map[string]string{
				"phrase": "Binary Christmas Tree",
				"limit":  "1",
			},
			mockComics: []core.Comics{
				{ID: 1, URL: "https://imgs.xkcd.com/comics/tree.png"},
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSearcher := &MockSearcher{}
			mockSearcher.On("IndexSearch", mock.Anything, 1, "Binary Christmas Tree").Return(tt.mockComics, tt.mockErr)

			handler := NewIndexSearchHandler(slog.Default(), mockSearcher, 10)

			req := httptest.NewRequest("GET", "/api/isearch", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockSearcher.AssertExpectations(t)
		})
	}
}
