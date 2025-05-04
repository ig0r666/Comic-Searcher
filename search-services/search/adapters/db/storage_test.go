package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"yadro.com/course/search/core"
)

func TestDB_SearchComics(t *testing.T) {
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
		name     string
		limit    int
		keywords []string
		mock     func()
		want     []core.Comics
		wantErr  bool
	}{
		{
			name:     "successful search",
			limit:    10,
			keywords: []string{"keyword1", "keyword2"},
			mock: func() {
				rows := sqlxmock.NewRows([]string{"comic_id", "image_url", "keywords"}).
					AddRow(1, "https://imgs.xkcd.com/comics/barrel_cropped_(1).jpg", "keywords1").
					AddRow(2, "https://imgs.xkcd.com/comics/tree_cropped_(1).jpg", "keywords2")
				mock.ExpectQuery(`SELECT comic_id, image_url FROM comics WHERE keywords && \$1.*LIMIT \$2`).
					WithArgs(sqlxmock.AnyArg(), 10).
					WillReturnRows(rows)
			},
			want: []core.Comics{
				{ID: 1, URL: "https://imgs.xkcd.com/comics/barrel_cropped_(1).jpg"},
				{ID: 2, URL: "https://imgs.xkcd.com/comics/tree_cropped_(1).jpg"},
			},
			wantErr: false,
		},
		{
			name:     "empty",
			limit:    10,
			keywords: []string{"test"},
			mock: func() {
				rows := sqlxmock.NewRows([]string{"comic_id", "image_url"})
				mock.ExpectQuery(`SELECT comic_id, image_url FROM comics WHERE keywords && \$1.*LIMIT \$2`).
					WithArgs(sqlxmock.AnyArg(), 10).
					WillReturnRows(rows)
			},
			want:    []core.Comics{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := storage.SearchComics(context.Background(), tt.limit, tt.keywords)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchComics error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDB_GetImageURL(t *testing.T) {
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
		id      int
		mock    func()
		want    string
		wantErr bool
	}{
		{
			name: "successful get",
			id:   1,
			mock: func() {
				row := mock.NewRows([]string{"image_url"}).
					AddRow("https://imgs.xkcd.com/comics/barrel_cropped_(1).jpg")
				mock.ExpectQuery(`SELECT image_url FROM comics WHERE comic_id = \$1`).
					WithArgs(1).
					WillReturnRows(row)
			},
			want:    "https://imgs.xkcd.com/comics/barrel_cropped_(1).jpg",
			wantErr: false,
		},
		{
			name: "not found",
			id:   3069,
			mock: func() {
				mock.ExpectQuery(`SELECT image_url FROM comics WHERE comic_id = \$1`).
					WithArgs(3069).
					WillReturnError(sql.ErrNoRows)
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "database error",
			id:   1,
			mock: func() {
				mock.ExpectQuery(`SELECT image_url FROM comics WHERE comic_id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := storage.GetImageURL(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageURL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
