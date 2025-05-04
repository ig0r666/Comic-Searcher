package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	idsExists   map[int]struct{}
	mu          sync.Mutex
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		idsExists:   make(map[int]struct{}),
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	if !s.mu.TryLock() {
		return ErrAlreadyExists
	}

	defer s.mu.Unlock()
	var wg sync.WaitGroup
	id, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get LastID", "error", err)
	}

	ids, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("failed to get ids from db", "error", err)
	}

	for i := range ids {
		if _, ok := s.idsExists[i]; !ok {
			s.idsExists[i] = struct{}{}
		}
	}

	sema := make(chan struct{}, s.concurrency)
	for i := 1; i <= id; i++ {
		if _, ok := s.idsExists[i]; ok {
			continue
		}
		wg.Add(1)
		sema <- struct{}{}
		go func() {
			defer func() {
				<-sema
			}()
			defer wg.Done()

			info, err := s.xkcd.Get(ctx, i)
			if err != nil {
				if errors.Is(err, Err404Comics) {
					err = s.db.Add(ctx, Comics{ID: 404})
					if err != nil {
						s.log.Error("failed to save comics", "error", err)
						return
					}
					return
				}
				s.log.Error("failed to get comics", "error", err)
				return
			}

			phrase := info.Title + " " + info.Transcript + " " + info.SafeTitle + " " + info.Alt
			words, err := s.words.Norm(ctx, phrase)
			if err != nil {
				s.log.Error("failed to normalize words for ", "error", err)
				return
			}

			comics := Comics{
				ID:    info.ID,
				URL:   info.URL,
				Words: words,
			}
			err = s.db.Add(ctx, comics)
			if err != nil {
				s.log.Error("failed to save comics", "error", err)
			}
		}()
	}
	wg.Wait()
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	stats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("failed to get stats", "error", err)
		return ServiceStats{}, err
	}

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get lastID", "error", err)
		return ServiceStats{}, err
	}

	return ServiceStats{
		DBStats:     stats,
		ComicsTotal: comicsTotal,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	var status ServiceStatus

	if s.mu.TryLock() {
		status = StatusIdle
		s.mu.Unlock()
	} else {
		status = StatusRunning
	}

	return status
}

func (s *Service) Drop(ctx context.Context) error {
	err := s.db.Drop(ctx)
	if err != nil {
		s.log.Error("failed to drop", "error", err)
	}

	s.idsExists = make(map[int]struct{})
	return nil
}
