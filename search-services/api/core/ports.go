package core

import "context"

type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Searcher interface {
	Search(context.Context, int, string) ([]Comics, error)
	IndexSearch(context.Context, int, string) ([]Comics, error)
}

type Loginer interface {
	Login(string, string) (string, error)
}

type TokenVerifier interface {
	Verify(string) error
}
