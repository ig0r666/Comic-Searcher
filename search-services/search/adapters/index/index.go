package index

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"time"

	"yadro.com/course/search/core"
)

type Index struct {
	log      *slog.Logger
	indexTTL time.Duration
	db       core.DB
	storage  map[string][]int
}

func NewIndex(log *slog.Logger, db core.DB, indexTTL time.Duration) *Index {
	index := &Index{
		log:      log,
		indexTTL: indexTTL,
		db:       db,
	}

	go index.UpdateIndex()
	return index
}

func (index *Index) UpdateIndex() {
	ticker := time.NewTicker(index.indexTTL)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		index.log.Info("updating index")
		comics, err := index.db.GetComics(ctx)
		if err != nil {
			index.log.Error("failed to get comics from db", "error", err)
		}

		if err = index.BuildIndex(comics); err != nil {
			index.log.Error("failed to build index", "error", err)
		}
	}
}

func (index *Index) BuildIndex(comics []core.Comics) error {
	newIndex := make(map[string][]int)
	for _, comic := range comics {
		var keywords []string
		if err := json.Unmarshal([]byte(comic.Keywords), &keywords); err != nil {
			index.log.Error("failed to unmarshal keywords", "error", err)
			continue
		}

		for _, word := range keywords {
			newIndex[word] = append(newIndex[word], comic.ID)
		}
	}

	index.storage = newIndex
	return nil
}

func (index Index) SearchByIndex(ctx context.Context, limit int, keywords []string) ([]core.Comics, error) {
	comicCount := make(map[int]int)

	for _, keyword := range keywords {
		if ids, exists := index.storage[keyword]; exists {
			for _, id := range ids {
				comicCount[id]++
			}
		}
	}

	if len(comicCount) == 0 {
		return []core.Comics{}, nil
	}

	type comicRate struct {
		id            int
		keywordsCount int
	}

	var sortedComics []comicRate
	for id, score := range comicCount {
		sortedComics = append(sortedComics, comicRate{id, score})
	}

	sort.Slice(sortedComics, func(i, j int) bool {
		if sortedComics[i].keywordsCount == sortedComics[j].keywordsCount {
			return sortedComics[i].id > sortedComics[j].id
		}
		return sortedComics[i].keywordsCount > sortedComics[j].keywordsCount
	})

	resultCount := min(limit, len(sortedComics))
	result := make([]core.Comics, 0, resultCount)

	for i := 0; i < resultCount; i++ {
		imageUrl, err := index.db.GetImageURL(ctx, sortedComics[i].id)
		if err != nil {
			index.log.Error("failed to get image from db", "error", err)
			return []core.Comics{}, err
		}

		result = append(result, core.Comics{ID: sortedComics[i].id, URL: imageUrl})
	}

	return result, nil
}
