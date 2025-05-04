package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"yadro.com/course/frontend/adapters/api"
	"yadro.com/course/frontend/adapters/rest"
	"yadro.com/course/frontend/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	apiClient := api.NewClient(cfg.APIAddress, log)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /admin", rest.AdminPageHandler(cfg.TemplatePath, log))
	mux.HandleFunc("POST /admin/login", rest.AdminLoginHandler(cfg.TemplatePath, log, apiClient))
	mux.HandleFunc("GET /admin/dashboard", rest.DashboardHandler(cfg.TemplatePath, log, apiClient))
	mux.HandleFunc("POST /admin/update", rest.AdminUpdateHandler(log, apiClient))
	mux.HandleFunc("POST /admin/drop", rest.AdminDropHandler(log, apiClient))

	mux.HandleFunc("GET /", rest.MainPageHandler(cfg.TemplatePath, log))
	mux.HandleFunc("GET /search", rest.SearchHandler(cfg.TemplatePath, log, apiClient))

	srv := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: mux,
	}

	log.Info("Starting server", "address", cfg.HTTPAddress)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
