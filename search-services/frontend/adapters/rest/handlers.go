package rest

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"yadro.com/course/frontend/core"
)

func getToken(r *http.Request) string {
	cookie, err := r.Cookie("admin_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SearchHandler(templatePath string, log *slog.Logger, api core.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		result, err := api.Search(query)
		if err != nil {
			log.Error("failed to search", "error", err)
			http.Error(w, "search error", http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles(
			filepath.Join(templatePath, "index.html"),
			filepath.Join(templatePath, "results.html"),
		))

		data := struct {
			Query  string
			Comics []struct {
				ID       int
				ImageURL string
			}
		}{
			Query: query,
		}

		for _, comic := range result.Comics {
			data.Comics = append(data.Comics, struct {
				ID       int
				ImageURL string
			}{
				ID:       comic.ID,
				ImageURL: comic.ImageURL,
			})
		}

		if err := tmpl.ExecuteTemplate(w, "results.html", data); err != nil {
			log.Error("template error", "error", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}
}

func MainPageHandler(templatePath string, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(templatePath + "/index.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			log.Error("template error", "error", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}
}

func AdminPageHandler(templatePath string, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(templatePath + "/login.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			log.Error("template error", "error", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}
}

func AdminLoginHandler(templatePath string, log *slog.Logger, api core.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := api.Login(r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			tmpl := template.Must(template.ParseFiles(templatePath + "/login.html"))
			if err := tmpl.Execute(w, struct{ Error string }{"Invalid credentials"}); err != nil {
				log.Error("template error", "error", err)
				http.Error(w, "template error", http.StatusInternalServerError)
			}
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_token",
			Value:    token,
			Path:     "/admin",
			HttpOnly: true,
			Secure:   true,
		})

		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	}
}

func DashboardHandler(templatePath string, log *slog.Logger, api core.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := api.GetStats()
		if err != nil {
			log.Error("failed to get stats", "error", err)
			http.Error(w, "failed to load stats", http.StatusInternalServerError)
			return
		}

		status, err := api.GetStatus()
		if err != nil {
			log.Error("failed to get status", "error", err)
			http.Error(w, "failed to load status", http.StatusInternalServerError)
			return
		}

		data := struct {
			Stats  core.Stats
			Status core.Status
		}{
			Stats:  stats,
			Status: status,
		}

		tmpl := template.Must(template.ParseFiles(
			filepath.Join(templatePath, "dashboard.html"),
		))
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			log.Error("template error", "error", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}
}

func AdminUpdateHandler(log *slog.Logger, api core.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)
		if token == "" {
			http.Redirect(w, r, "/admin/login", http.StatusUnauthorized)
			return
		}

		err := api.Update(token)

		if err != nil {
			log.Error("failed to update", "error", err)
			http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		}
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	}
}

func AdminDropHandler(log *slog.Logger, api core.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)
		if token == "" {
			http.Redirect(w, r, "/admin/login", http.StatusUnauthorized)
			return
		}

		err := api.Drop(token)
		if err != nil {
			log.Error("failed to update", "error", err)
			http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		}

		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	}
}
