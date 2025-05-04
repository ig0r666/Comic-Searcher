package search

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Println("search client address", address)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, nil)
	if err != nil {
		c.log.Error("failed to ping", "error", err)
		return err
	}
	return nil
}

func (c Client) Search(ctx context.Context, limit int, phrase string) ([]core.Comics, error) {
	req := &searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  int64(limit),
	}

	resp, err := c.client.Search(ctx, req)
	if err != nil {
		c.log.Error("failed to search comics", "error", err)
		return nil, err
	}

	comics := make([]core.Comics, len(resp.GetComics()))
	for i, comic := range resp.GetComics() {
		comics[i] = core.Comics{
			ID:  int(comic.GetId()),
			URL: comic.GetUrl(),
		}
	}

	return comics, nil
}

func (c Client) IndexSearch(ctx context.Context, limit int, phrase string) ([]core.Comics, error) {
	req := &searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  int64(limit),
	}

	resp, err := c.client.IndexSearch(ctx, req)
	if err != nil {
		c.log.Error("failed to search comics", "error", err)
		return nil, err
	}

	comics := make([]core.Comics, len(resp.GetComics()))
	for i, comic := range resp.GetComics() {
		comics[i] = core.Comics{
			ID:  int(comic.GetId()),
			URL: comic.GetUrl(),
		}
	}

	return comics, nil
}
