package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.service.Search(ctx, int(in.Limit), in.Phrase)
	if err != nil {
		return nil, err
	}

	searchReply := &searchpb.SearchReply{
		Comics: make([]*searchpb.Comics, 0, len(comics)),
		Total:  int64(len(comics)),
	}

	for _, comic := range comics {
		searchReply.Comics = append(searchReply.Comics, &searchpb.Comics{
			Id:  int64(comic.ID),
			Url: comic.URL,
		})
	}
	return searchReply, nil
}

func (s *Server) IndexSearch(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.service.IndexSearch(ctx, int(in.Limit), in.Phrase)
	if err != nil {
		return nil, err
	}

	searchReply := &searchpb.SearchReply{
		Comics: make([]*searchpb.Comics, 0, len(comics)),
		Total:  int64(len(comics)),
	}

	for _, comic := range comics {
		searchReply.Comics = append(searchReply.Comics, &searchpb.Comics{
			Id:  int64(comic.ID),
			Url: comic.URL,
		})
	}
	return searchReply, nil
}
