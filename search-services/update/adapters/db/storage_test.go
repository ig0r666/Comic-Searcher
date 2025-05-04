package db

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"yadro.com/course/update/core"
)

func TestDB_Stats(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("failed to mock db")
	}
	defer db.Close()

	storage := &DB{
		log:  slog.Default(),
		conn: db,
	}

	tests := []struct {
		name    string
		mock    func()
		want    core.DBStats
		wantErr bool
	}{
		{
			name: "successful",
			mock: func() {
				rows1 := sqlxmock.NewRows([]string{"count"}).AddRow(10)
				mock.ExpectQuery(`SELECT COALESCE\(COUNT\(\*\), 0\) FROM comics`).
					WillReturnRows(rows1)

				rows2 := sqlxmock.NewRows([]string{"words_total", "words_unique"}).
					AddRow(100, 50)
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(array_length\(keywords, 1\)\), 0\) AS words_total, COALESCE\(COUNT\(DISTINCT keyword\), 0\) AS words_unique FROM comics, LATERAL unnest\(keywords\) AS keyword`).
					WillReturnRows(rows2)
			},
			want: core.DBStats{
				ComicsFetched: 10,
				WordsTotal:    100,
				WordsUnique:   50,
			},
			wantErr: false,
		},
		{
			name: "failed to get count",
			mock: func() {
				mock.ExpectQuery(`SELECT COALESCE\(COUNT\(\*\), 0\) FROM comics`).
					WillReturnError(errors.New("db error"))
			},
			want:    core.DBStats{},
			wantErr: true,
		},
		{
			name: "failed to get stats",
			mock: func() {
				rows1 := sqlxmock.NewRows([]string{"count"}).AddRow(10)
				mock.ExpectQuery(`SELECT COALESCE\(COUNT\(\*\), 0\) FROM comics`).
					WillReturnRows(rows1)

				mock.ExpectQuery(`SELECT COALESCE\(SUM\(array_length\(keywords, 1\)\), 0\) AS words_total, COALESCE\(COUNT\(DISTINCT keyword\), 0\) AS words_unique FROM comics, LATERAL unnest\(keywords\) AS keyword`).
					WillReturnError(errors.New("db error"))
			},
			want:    core.DBStats{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := storage.Stats(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDB_IDs(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("failed to mock db")
	}
	defer db.Close()

	storage := &DB{
		log:  slog.Default(),
		conn: db,
	}

	tests := []struct {
		name    string
		mock    func()
		want    []int
		wantErr bool
	}{
		{
			name: "successful",
			mock: func() {
				rows := sqlxmock.NewRows([]string{"comic_id"}).
					AddRow(1).
					AddRow(2).
					AddRow(3)
				mock.ExpectQuery(`SELECT comic_id FROM comics`).
					WillReturnRows(rows)
			},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name: "empty",
			mock: func() {
				rows := sqlxmock.NewRows([]string{"comic_id"})
				mock.ExpectQuery(`SELECT comic_id FROM comics`).
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Db error",
			mock: func() {
				mock.ExpectQuery(`SELECT comic_id FROM comics`).
					WillReturnError(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := storage.IDs(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("IDs error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDB_Drop(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("failed to mock db")
	}
	defer db.Close()

	storage := &DB{
		log:  slog.Default(),
		conn: db,
	}

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful",
			mock: func() {
				mock.ExpectExec(`TRUNCATE TABLE comics`).
					WillReturnResult(sqlxmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name: "Db error",
			mock: func() {
				mock.ExpectExec(`TRUNCATE TABLE comics`).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.Drop(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Drop error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
