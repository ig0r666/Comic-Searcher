package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockTokenVerifier struct {
	shouldError bool
}

func (m *MockTokenVerifier) Verify(token string) error {
	if m.shouldError {
		return errors.New("invalid token")
	}
	return nil
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockError      bool
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "not enough parts in header",
			authHeader:     "Token",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "too many parts in token",
			authHeader:     "Token test test",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "invalid token",
			authHeader:     "Token false",
			mockError:      true,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "valid token",
			authHeader:     "Token valid",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := &MockTokenVerifier{shouldError: tt.mockError}

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			Auth(next, verifier).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code")
			}

			if nextCalled != tt.shouldCallNext {
				t.Errorf("next handler called")
			}
		})
	}
}
