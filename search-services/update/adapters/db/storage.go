package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
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

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	query := `
		INSERT INTO comics (comic_id, image_url, keywords)
		VALUES (:id, :url, :words)
		ON CONFLICT (comic_id) DO NOTHING;`

	_, err := db.conn.NamedExecContext(ctx, query, comics)
	if err != nil {
		db.log.Error("failed to insert comic", "error", err, "comic_id", comics.ID)
		return err
	}
	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	err := db.conn.GetContext(ctx, &stats.ComicsFetched, `SELECT COALESCE(COUNT(*), 0) FROM comics;`)
	if err != nil {
		db.log.Error("failed to get comics count", "error", err)
		return core.DBStats{}, err
	}

	query := `
		SELECT
    	 COALESCE(SUM(array_length(keywords, 1)), 0) AS words_total,
    	 COALESCE(COUNT(DISTINCT keyword), 0) AS words_unique
		FROM comics,
		LATERAL unnest(keywords) AS keyword;`

	err = db.conn.GetContext(ctx, &stats, query)
	if err != nil {
		db.log.Error("failed to get words stats", "error", err)
		return core.DBStats{}, err
	}

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, `SELECT comic_id FROM comics;`)
	if err != nil {
		db.log.Error("failed to query comic IDs", "error", err)
		return nil, err
	}

	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, `TRUNCATE TABLE comics;`)
	if err != nil {
		db.log.Error("failed to drop table", "error", err)
		return err
	}
	return nil
}
