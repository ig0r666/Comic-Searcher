package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
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

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	status, err := c.client.Status(ctx, nil)
	if err != nil {
		c.log.Error("failed to get status", "error", err)
		return core.StatusUpdateUnknown, err
	}

	var out core.UpdateStatus
	switch status.Status {
	case updatepb.Status_STATUS_IDLE:
		out = core.StatusUpdateIdle
	case updatepb.Status_STATUS_RUNNING:
		out = core.StatusUpdateRunning
	default:
		out = core.StatusUpdateUnknown
	}
	return out, nil
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	stats, err := c.client.Stats(ctx, nil)
	if err != nil {
		c.log.Error("failed to get stats", "error", err)
		return core.UpdateStats{}, err
	}
	return core.UpdateStats{
		WordsTotal:    int(stats.WordsTotal),
		WordsUnique:   int(stats.WordsUnique),
		ComicsFetched: int(stats.ComicsFetched),
		ComicsTotal:   int(stats.ComicsTotal),
	}, nil
}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, nil)
	if err != nil {
		c.log.Error("failed to update db", "error", err)
		return err
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, nil)
	if err != nil {
		c.log.Error("failed to drop", "error", err)
		return err
	}
	return nil
}
