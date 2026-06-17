package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lehmann314159/mockspotify/internal/api"
	"github.com/lehmann314159/mockspotify/internal/db"
	"github.com/lehmann314159/mockspotify/internal/web"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	dbPath := getenv("MOCKSPOTIFY_DB", "./data/mockspotify.db")
	port := getenv("MOCKSPOTIFY_PORT", "8090")
	token := getenv("MOCKSPOTIFY_TOKEN", "mockspotify-dev-token-2026")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := web.LoadTemplatesFromDir("templates"); err != nil {
		log.Fatalf("load templates: %v", err)
	}

	// API mux: all authenticated routes live here.
	apiMux := http.NewServeMux()
	api.Register(apiMux, database)

	// Root mux
	mux := http.NewServeMux()

	// Token endpoint — no auth
	mux.HandleFunc("POST /api/token", api.TokenHandler(token))

	// All other /api/ routes — auth required
	mux.Handle("/api/", api.AuthMiddleware(token, apiMux))

	// Web UI
	mux.HandleFunc("GET /{$}", web.DashboardHandler(database))
	mux.HandleFunc("GET /artists", web.ArtistsPageHandler(database))
	mux.HandleFunc("GET /albums", web.AlbumsPageHandler(database))
	mux.HandleFunc("GET /tracks", web.TracksPageHandler(database))
	mux.HandleFunc("GET /playlists", web.PlaylistsPageHandler(database))
	mux.HandleFunc("GET /playlists/{id}", web.PlaylistDetailHandler(database))

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("MockSpotify listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
