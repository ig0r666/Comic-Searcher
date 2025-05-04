package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/core"
)

func NewLoginHandler(log *slog.Logger, a core.Loginer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Error("failed to decode req", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		token, err := a.Login(user.Name, user.Password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		if _, err = w.Write([]byte(token)); err != nil {
			log.Error("failed to send token", "error", err)
		}
	}
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := make(map[string]string)

		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				replies[name] = "unavailable"
				log.Error("Service unavailable", "service", name, "error", err)
			} else {
				replies[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"replies": replies}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}

}

func NewUpdateHandler(log *slog.Logger, updater core.Updater, verifier core.TokenVerifier) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		err := updater.Update(r.Context())
		if err != nil {
			http.Error(w, "update is already exists", http.StatusAccepted)
		}
	}

	return middleware.Auth(handler, verifier)
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get stats", "error", err)
			http.Error(w, "failed to get stats", http.StatusInternalServerError)
			return
		}

		resp := map[string]int{
			"words_total":    stats.WordsTotal,
			"words_unique":   stats.WordsUnique,
			"comics_fetched": stats.ComicsFetched,
			"comics_total":   stats.ComicsTotal,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("failed to encode stats", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("failed to get status", "error", err)
			http.Error(w, "failed to get status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]core.UpdateStatus{"status": status}); err != nil {
			log.Error("failed to encode status response", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater, verifier core.TokenVerifier) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		err := updater.Drop(r.Context())
		if err != nil {
			log.Error("failed to drop", "error", err)
			http.Error(w, "failed to drop", http.StatusInternalServerError)
		}
	}

	return middleware.Auth(handler, verifier)
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher, concurrencyLimit int) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "10"
		}

		num, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		comics, err := searcher.Search(r.Context(), num, phrase)
		if err != nil {
			log.Error("arguments are not acceptable", "error", err)
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"comics": make([]map[string]interface{}, 0, len(comics)),
			"total":  len(comics),
		}

		for _, comic := range comics {
			resp["comics"] = append(resp["comics"].([]map[string]interface{}), map[string]interface{}{
				"id":  comic.ID,
				"url": comic.URL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}

	return middleware.Concurrency(handler, concurrencyLimit)
}

func NewIndexSearchHandler(log *slog.Logger, searcher core.Searcher, rateLimit int) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "10"
		}

		num, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		comics, err := searcher.IndexSearch(r.Context(), num, phrase)
		if err != nil {
			if len(comics) == 0 {
				return
			}
			log.Error("arguments are not acceptable", "error", err)
			http.Error(w, "Bad arguments", http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"comics": make([]map[string]interface{}, 0, len(comics)),
			"total":  len(comics),
		}

		for _, comic := range comics {
			resp["comics"] = append(resp["comics"].([]map[string]interface{}), map[string]interface{}{
				"id":  comic.ID,
				"url": comic.URL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}

	return middleware.Rate(handler, rateLimit)
}
