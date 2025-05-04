package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

type config struct {
	Address string `yaml:"address" env:"WORDS_ADDRESS"`
}

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	out := words.NormalizedString(in.Phrase)
	return &wordspb.WordsReply{
		Words: out,
	}, nil
}

func main() {
	var cfg config
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath != "" {
		if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
			log.Printf("error reading config file: %v\n", err)
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Printf("error reading env: %v\n", err)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	log.Printf("server is listening on address: %s", cfg.Address)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
