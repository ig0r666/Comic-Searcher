package core

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

type MockWords struct {
	mock.Mock
}

type MockIndex struct {
	mock.Mock
}

func (m *MockDB) SearchComics(ctx context.Context, limit int, keywords []string) ([]Comics, error) {
	args := m.Called(ctx, limit, keywords)
	return args.Get(0).([]Comics), args.Error(1)
}

func (m *MockDB) GetImageURL(ctx context.Context, id int) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockDB) GetComics(ctx context.Context) ([]Comics, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Comics), args.Error(1)
}

func (m *MockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	args := m.Called(ctx, phrase)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIndex) SearchByIndex(ctx context.Context, limit int, keywords []string) ([]Comics, error) {
	args := m.Called(ctx, limit, keywords)
	return args.Get(0).([]Comics), args.Error(1)
}

func TestService_Search(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		phrase      string
		limit       int
		mockNorm    []string
		mockNormErr error
		mockDBRes   []Comics
		mockDBErr   error
		want        []Comics
		wantErr     bool
	}{
		{
			name:     "successful search",
			phrase:   "funny cats",
			limit:    5,
			mockNorm: []string{"cat", "dog"},
			mockDBRes: []Comics{
				{ID: 1, URL: "url1"},
				{ID: 2, URL: "url2"},
			},
			want: []Comics{
				{ID: 1, URL: "url1"},
				{ID: 2, URL: "url2"},
			},
		},
		{
			name:        "failed to norm",
			phrase:      "invalid",
			limit:       5,
			mockNormErr: errors.New("failed to norm"),
			wantErr:     true,
		},
		{
			name:      "failed to search",
			phrase:    "dogs",
			limit:     3,
			mockNorm:  []string{"dog"},
			mockDBErr: errors.New("failed to search"),
			wantErr:   true,
		},
		{
			name:      "empty",
			phrase:    "test",
			limit:     10,
			mockNorm:  []string{"test"},
			mockDBRes: []Comics{},
			want:      []Comics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWords := new(MockWords)
			mockDB := new(MockDB)
			mockIndex := new(MockIndex)

			mockWords.On("Norm", ctx, tt.phrase).Return(tt.mockNorm, tt.mockNormErr)
			if tt.mockNormErr == nil {
				mockDB.On("SearchComics", ctx, tt.limit, tt.mockNorm).
					Return(tt.mockDBRes, tt.mockDBErr)
			}

			service := &Service{
				log:   slog.Default(),
				db:    mockDB,
				words: mockWords,
				index: mockIndex,
			}

			got, err := service.Search(ctx, tt.limit, tt.phrase)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockWords.AssertExpectations(t)
			if tt.mockNormErr == nil {
				mockDB.AssertExpectations(t)
			}
		})
	}
}

func TestService_IndexSearch(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		phrase       string
		limit        int
		mockNorm     []string
		mockNormErr  error
		mockIndexRes []Comics
		mockIndexErr error
		want         []Comics
		wantErr      bool
	}{
		{
			name:     "successful index search",
			phrase:   "cats dogs",
			limit:    5,
			mockNorm: []string{"cat", "dog"},
			mockIndexRes: []Comics{
				{ID: 1, URL: "url1"},
				{ID: 3, URL: "url3"},
			},
			want: []Comics{
				{ID: 1, URL: "url1"},
				{ID: 3, URL: "url3"},
			},
		},
		{
			name:        "failed to norm",
			phrase:      "test",
			limit:       5,
			mockNormErr: errors.New("failed to norm"),
			wantErr:     true,
		},
		{
			name:         "index search err",
			phrase:       "cats",
			limit:        3,
			mockNorm:     []string{"cat"},
			mockIndexErr: errors.New("failed to index search"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWords := new(MockWords)
			mockDB := new(MockDB)
			mockIndex := new(MockIndex)

			mockWords.On("Norm", ctx, tt.phrase).Return(tt.mockNorm, tt.mockNormErr)
			if tt.mockNormErr == nil {
				mockIndex.On("SearchByIndex", ctx, tt.limit, tt.mockNorm).
					Return(tt.mockIndexRes, tt.mockIndexErr)
			}

			service := &Service{
				log:   slog.Default(),
				db:    mockDB,
				words: mockWords,
				index: mockIndex,
			}

			got, err := service.IndexSearch(ctx, tt.limit, tt.phrase)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockWords.AssertExpectations(t)
			if tt.mockNormErr == nil {
				mockIndex.AssertExpectations(t)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	mockDB := new(MockDB)
	mockWords := new(MockWords)
	mockIndex := new(MockIndex)

	service, err := NewService(slog.Default(), mockDB, mockWords, mockIndex)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockWords, service.words)
	assert.Equal(t, mockIndex, service.index)
}
