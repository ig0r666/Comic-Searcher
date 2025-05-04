package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) SearchComics(ctx context.Context, limit int, keywords []string) ([]core.Comics, error) {
	query := `
	SELECT comic_id, image_url
	FROM comics
	WHERE keywords && $1
	ORDER BY (
		SELECT COUNT(*) 
		FROM unnest(keywords) AS kw
		WHERE kw = ANY($1)
	) DESC
	LIMIT $2
	`

	var dbComics []core.DbComics
	err := db.conn.SelectContext(ctx, &dbComics, query, pq.Array(keywords), limit)
	if err != nil {
		db.log.Error("failed to do query", "error", err)
		return nil, err
	}

	comics := make([]core.Comics, len(dbComics))
	for i, c := range dbComics {
		comics[i] = core.Comics{
			ID:  c.ID,
			URL: c.URL,
		}
	}
	return comics, nil
}

func (db *DB) GetImageURL(ctx context.Context, id int) (string, error) {
	query := `SELECT image_url FROM comics WHERE comic_id = $1`

	var imageURL string
	err := db.conn.GetContext(ctx, &imageURL, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		db.log.Error("failed to get image_url", "error", err)
		return "", err
	}

	return imageURL, nil

}

func (db *DB) GetComics(ctx context.Context) ([]core.Comics, error) {
	query := `
        SELECT 
            comic_id, 
            ARRAY_TO_JSON(COALESCE(keywords, ARRAY[]::TEXT[])) AS keywords
        FROM comics
    `

	var comics []core.DbComics
	err := db.conn.SelectContext(ctx, &comics, query)
	if err != nil {
		db.log.Error("failed to get comics", "error", err)
		return []core.Comics{}, err
	}

	out := make([]core.Comics, len(comics))
	for i, c := range comics {
		out[i] = core.Comics{
			ID:       c.ID,
			Keywords: c.Keywords,
		}
	}

	return out, nil
}
