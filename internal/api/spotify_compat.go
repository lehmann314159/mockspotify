package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/lehmann314159/mockspotify/internal/db"
)

// RegisterV1 adds Spotify API-compatible routes under /v1/.
// These endpoints accept the same Bearer token as the /api/ routes.
func RegisterV1(mux *http.ServeMux, database *sql.DB) {
	mux.HandleFunc("GET /v1/search", spotifySearch(database))
	mux.HandleFunc("GET /v1/playlists/{id}/tracks", spotifyPlaylistTracks(database))
}

type spotifyArtist struct {
	Name string `json:"name"`
}

type spotifyTrack struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	DurationMS int             `json:"duration_ms"`
	Artists    []spotifyArtist `json:"artists"`
}

func toSpotifyTrack(t db.TrackDetail) spotifyTrack {
	return spotifyTrack{
		ID:         t.ID,
		Name:       t.Title,
		DurationMS: t.DurationMS,
		Artists:    []spotifyArtist{{Name: t.ArtistName}},
	}
}

func spotifySearch(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		searchType := r.URL.Query().Get("type")
		limit := queryInt(r, "limit", 20)
		offset := queryInt(r, "offset", 0)

		switch searchType {
		case "track":
			genre := extractGenre(q)
			tracks, total, err := db.ListTracks(database, db.TrackFilter{Genre: genre, Limit: limit, Offset: offset})
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			items := make([]spotifyTrack, len(tracks))
			for i, t := range tracks {
				items[i] = toSpotifyTrack(t)
			}
			var next *string
			if offset+limit < total {
				s := fmt.Sprintf("/search?q=%s&type=track&market=US&limit=%d&offset=%d", q, limit, offset+limit)
				next = &s
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"tracks": map[string]any{
					"items":  items,
					"total":  total,
					"limit":  limit,
					"offset": offset,
					"next":   next,
				},
			})

		case "playlist":
			pls, total, err := db.ListPlaylists(database, db.PlaylistFilter{Genre: q, Limit: limit, Offset: offset})
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			items := make([]map[string]any, len(pls))
			for i, p := range pls {
				items[i] = map[string]any{"id": p.ID, "name": p.Name}
			}
			var next *string
			if offset+limit < total {
				s := fmt.Sprintf("/search?q=%s&type=playlist&limit=%d&offset=%d", q, limit, offset+limit)
				next = &s
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"playlists": map[string]any{
					"items": items,
					"next":  next,
				},
			})

		default:
			writeError(w, http.StatusBadRequest, "unsupported type")
		}
	}
}

func spotifyPlaylistTracks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		limit := queryInt(r, "limit", 100)
		offset := queryInt(r, "offset", 0)

		items, total, err := db.GetPlaylistTracks(database, id, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		wrapped := make([]map[string]any, len(items))
		for i, item := range items {
			wrapped[i] = map[string]any{"track": toSpotifyTrack(item.Track)}
		}

		var next *string
		if offset+limit < total {
			s := fmt.Sprintf("/playlists/%s/tracks?limit=%d&offset=%d", id, limit, offset+limit)
			next = &s
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items": wrapped,
			"total": total,
			"next":  next,
		})
	}
}

// extractGenre parses "genre:pop" → "pop"; otherwise returns q unchanged.
func extractGenre(q string) string {
	if after, ok := strings.CutPrefix(q, "genre:"); ok {
		return after
	}
	return q
}