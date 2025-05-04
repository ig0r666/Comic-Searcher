package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
	index Index
}

func NewService(log *slog.Logger, db DB, words Words, index Index) (*Service, error) {
	service := &Service{
		log:   log,
		db:    db,
		words: words,
		index: index,
	}

	return service, nil
}

func (s Service) Search(ctx context.Context, limit int, phrase string) ([]Comics, error) {
	normReq, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to normalize req", "error", err)
		return []Comics{}, err
	}

	comics, err := s.db.SearchComics(ctx, limit, normReq)
	if err != nil {
		s.log.Error("failed to search comics in db", "error", err)
		return []Comics{}, err
	}

	return comics, nil
}

func (s Service) IndexSearch(ctx context.Context, limit int, phrase string) ([]Comics, error) {
	normReq, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to normalize req", "error", err)
		return []Comics{}, err
	}

	comics, err := s.index.SearchByIndex(ctx, limit, normReq)
	if err != nil {
		s.log.Error("failed to isearch comics in db", "error", err)
		return []Comics{}, err
	}

	return comics, nil
}
