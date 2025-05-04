package index

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"yadro.com/course/search/core"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) SearchComics(ctx context.Context, limit int, keywords []string) ([]core.Comics, error) {
	args := m.Called(ctx, limit, keywords)
	return args.Get(0).([]core.Comics), args.Error(1)
}

func (m *MockDB) GetImageURL(ctx context.Context, id int) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockDB) GetComics(ctx context.Context) ([]core.Comics, error) {
	args := m.Called(ctx)
	return args.Get(0).([]core.Comics), args.Error(1)
}

func TestBuildIndex(t *testing.T) {
	tests := []struct {
		name    string
		comics  []core.Comics
		want    map[string][]int
		wantErr bool
	}{
		{
			name: "valid keywords",
			comics: []core.Comics{
				{ID: 1, Keywords: `["cat","dog"]`},
				{ID: 2, Keywords: `["cat"]`},
			},
			want: map[string][]int{
				"cat": {1, 2},
				"dog": {1},
			},
			wantErr: false,
		},
		{
			name: "invalid JSON",
			comics: []core.Comics{
				{ID: 1, Keywords: "invalid json"},
			},
			want:    map[string][]int{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := &Index{
				log:     slog.Default(),
				storage: make(map[string][]int),
			}

			err := idx.BuildIndex(tt.comics)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildIndex error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, idx.storage)
		})
	}
}

func TestSearchByIndex(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		storage   map[string][]int
		keywords  []string
		limit     int
		mockSetup func(*MockDB)
		want      []core.Comics
		wantErr   bool
	}{
		{
			name: "one keyword",
			storage: map[string][]int{
				"cat": {1, 2},
				"dog": {2, 3},
			},
			keywords: []string{"cat"},
			limit:    10,
			mockSetup: func(m *MockDB) {
				m.On("GetImageURL", ctx, 1).Return("url1", nil)
				m.On("GetImageURL", ctx, 2).Return("url2", nil)
			},
			want: []core.Comics{
				{ID: 2, URL: "url2"},
				{ID: 1, URL: "url1"},
			},
			wantErr: false,
		},
		{
			name: "two or more keywords",
			storage: map[string][]int{
				"cat": {1, 2},
				"dog": {2, 3},
			},
			keywords: []string{"cat", "dog"},
			limit:    10,
			mockSetup: func(m *MockDB) {
				m.On("GetImageURL", ctx, 2).Return("url2", nil)
				m.On("GetImageURL", ctx, 1).Return("url1", nil)
				m.On("GetImageURL", ctx, 3).Return("url3", nil)
			},
			want: []core.Comics{
				{ID: 2, URL: "url2"},
				{ID: 3, URL: "url3"},
				{ID: 1, URL: "url1"},
			},
			wantErr: false,
		},
		{
			name: "failed to get image url",
			storage: map[string][]int{
				"cat": {1},
			},
			keywords: []string{"cat"},
			limit:    10,
			mockSetup: func(m *MockDB) {
				m.On("GetImageURL", ctx, 1).Return("", assert.AnError)
			},
			want:    []core.Comics{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			idx := &Index{
				log:     slog.Default(),
				db:      mockDB,
				storage: tt.storage,
			}

			got, err := idx.SearchByIndex(ctx, tt.limit, tt.keywords)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchByIndex error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
			mockDB.AssertExpectations(t)
		})
	}
}
