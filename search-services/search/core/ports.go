package core

import "context"

type DB interface {
	SearchComics(ctx context.Context, limit int, keywords []string) ([]Comics, error)
	GetImageURL(ctx context.Context, id int) (string, error)
	GetComics(ctx context.Context) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Searcher interface {
	Search(ctx context.Context, limit int, phrase string) ([]Comics, error)
	IndexSearch(ctx context.Context, limit int, phrase string) ([]Comics, error)
}

type Index interface {
	SearchByIndex(ctx context.Context, limit int, keywords []string) ([]Comics, error)
}
