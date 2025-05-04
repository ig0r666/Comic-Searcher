package core

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Add(ctx context.Context, comics Comics) error {
	args := m.Called(ctx, comics)
	return args.Error(0)
}

func (m *MockDB) Stats(ctx context.Context) (DBStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(DBStats), args.Error(1)
}

func (m *MockDB) IDs(ctx context.Context) ([]int, error) {
	args := m.Called(ctx)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockDB) Drop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockXKCD struct {
	mock.Mock
}

func (m *MockXKCD) Get(ctx context.Context, id int) (XKCDInfo, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(XKCDInfo), args.Error(1)
}

func (m *MockXKCD) LastID(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type MockWords struct {
	mock.Mock
}

func (m *MockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	args := m.Called(ctx, phrase)
	return args.Get(0).([]string), args.Error(1)
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(db *MockDB, xkcd *MockXKCD, words *MockWords)
		wantErr     bool
		expectLock  bool
		concurrency int
	}{
		{
			name: "successful update",
			setupMocks: func(db *MockDB, xkcd *MockXKCD, words *MockWords) {
				xkcd.On("LastID", mock.Anything).Return(2, nil)
				db.On("IDs", mock.Anything).Return([]int{}, nil)

				xkcd.On("Get", mock.Anything, 1).Return(XKCDInfo{
					ID:         1,
					Title:      "Barrel - Part 1",
					URL:        "https://imgs.xkcd.com/comics/barrel_cropped_(1).jpg",
					Transcript: "[[A boy sits in a barrel which is floating in an ocean.]]\nBoy: I wonder where I'll float next?\n[[The barrel drifts into the distance. Nothing else can be seen.]]\n{{Alt: Don't we all.}}",
					SafeTitle:  "Barrel - Part 1",
					Alt:        "Don't we all.",
				}, nil)
				xkcd.On("Get", mock.Anything, 2).Return(XKCDInfo{
					ID:         2,
					Title:      "Petit Trees (sketch)",
					URL:        "https://imgs.xkcd.com/comics/tree_cropped_(1).jpg",
					Transcript: "[[Two trees are growing on opposite sides of a sphere.]]\n{{Alt-title: 'Petit' being a reference to Le Petit Prince, which I only thought about halfway through the sketch}}",
					SafeTitle:  "Petit Trees (sketch)",
					Alt:        "'Petit' being a reference to Le Petit Prince, which I only thought about halfway through the sketch",
				}, nil)

				words.On("Norm", mock.Anything, mock.Anything).Return([]string{"word1", "word2"}, nil).Twice()
				db.On("Add", mock.Anything, mock.Anything).Return(nil).Twice()
			},
			wantErr:     false,
			expectLock:  true,
			concurrency: 2,
		},
		{
			name:        "Already running",
			setupMocks:  func(db *MockDB, xkcd *MockXKCD, words *MockWords) {},
			wantErr:     true,
			expectLock:  false,
			concurrency: 1,
		},
		{
			name: "404 comic",
			setupMocks: func(db *MockDB, xkcd *MockXKCD, words *MockWords) {
				xkcd.On("LastID", mock.Anything).Return(1, nil)
				db.On("IDs", mock.Anything).Return([]int{}, nil)
				xkcd.On("Get", mock.Anything, 1).Return(XKCDInfo{}, Err404Comics)
				db.On("Add", mock.Anything, Comics{ID: 404}).Return(nil)
			},
			wantErr:     false,
			expectLock:  true,
			concurrency: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MockDB{}
			xkcd := &MockXKCD{}
			words := &MockWords{}

			tt.setupMocks(db, xkcd, words)

			service := &Service{
				log:         slog.Default(),
				db:          db,
				xkcd:        xkcd,
				words:       words,
				concurrency: tt.concurrency,
				idsExists:   make(map[int]struct{}),
				mu:          sync.Mutex{},
			}

			if !tt.expectLock {
				service.mu.Lock()
			}

			err := service.Update(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrAlreadyExists))
			} else {
				assert.NoError(t, err)
			}

			if tt.expectLock {
				assert.True(t, service.mu.TryLock())
				service.mu.Unlock()
			}

			db.AssertExpectations(t)
			xkcd.AssertExpectations(t)
			words.AssertExpectations(t)
		})
	}
}

func TestService_Stats(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(db *MockDB, xkcd *MockXKCD)
		want       ServiceStats
		wantErr    bool
	}{
		{
			name: "successful get stats",
			setupMocks: func(db *MockDB, xkcd *MockXKCD) {
				db.On("Stats", mock.Anything).Return(DBStats{
					ComicsFetched: 111,
					WordsTotal:    666,
					WordsUnique:   333,
				}, nil)
				xkcd.On("LastID", mock.Anything).Return(3076, nil)
			},
			want: ServiceStats{
				DBStats: DBStats{
					ComicsFetched: 111,
					WordsTotal:    666,
					WordsUnique:   333,
				},
				ComicsTotal: 3076,
			},
			wantErr: false,
		},
		{
			name: "DB error",
			setupMocks: func(db *MockDB, xkcd *MockXKCD) {
				db.On("Stats", mock.Anything).Return(DBStats{}, errors.New("db error"))
			},
			want:    ServiceStats{},
			wantErr: true,
		},
		{
			name: "XKCD error",
			setupMocks: func(db *MockDB, xkcd *MockXKCD) {
				db.On("Stats", mock.Anything).Return(DBStats{
					ComicsFetched: 111,
					WordsTotal:    666,
					WordsUnique:   333,
				}, nil)
				xkcd.On("LastID", mock.Anything).Return(0, errors.New("xkcd error"))
			},
			want:    ServiceStats{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MockDB{}
			xkcd := &MockXKCD{}
			tt.setupMocks(db, xkcd)

			service := &Service{
				log:  slog.Default(),
				db:   db,
				xkcd: xkcd,
			}

			got, err := service.Stats(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			db.AssertExpectations(t)
			xkcd.AssertExpectations(t)
		})
	}
}

func TestService_Status(t *testing.T) {
	service := &Service{}

	status := service.Status(context.Background())
	assert.Equal(t, StatusIdle, status)

	service.mu.Lock()
	defer service.mu.Unlock()

	status = service.Status(context.Background())
	assert.Equal(t, StatusRunning, status)
}

func TestService_Drop(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(db *MockDB)
		wantErr    bool
	}{
		{
			name: "Successful drop",
			setupMocks: func(db *MockDB) {
				db.On("Drop", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MockDB{}
			tt.setupMocks(db)

			service := &Service{
				log:       slog.Default(),
				db:        db,
				idsExists: map[int]struct{}{1: {}},
			}

			err := service.Drop(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Empty(t, service.idsExists)

			db.AssertExpectations(t)
		})
	}
}
