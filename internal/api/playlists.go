package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lehmann314159/mockspotify/internal/db"
)


func listPlaylists(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := db.PlaylistFilter{
			Genre:  r.URL.Query().Get("genre"),
			Limit:  queryInt(r, "limit", 20),
			Offset: queryInt(r, "offset", 0),
		}
		pls, total, err := db.ListPlaylists(database, f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"playlists": pls, "total": total})
	}
}

func getPlaylist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p, err := db.GetPlaylist(database, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if p == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func createPlaylist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p db.Playlist
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := db.InsertPlaylist(database, p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func updatePlaylist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p db.Playlist
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		p.ID = r.PathValue("id")
		if err := db.UpdatePlaylist(database, p); err != nil {
			if err.Error() == "not found" {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func deletePlaylist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := db.DeletePlaylist(database, id); err != nil {
			if err.Error() == "not found" {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getPlaylistTracks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		items, total, err := db.GetPlaylistTracks(database, id, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
	}
}

func addPlaylistTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistID := r.PathValue("id")
		var body struct {
			TrackID  string `json:"track_id"`
			Position int    `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := db.AddTrackToPlaylist(database, playlistID, body.TrackID, body.Position); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func removePlaylistTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistID := r.PathValue("id")
		trackID := r.PathValue("track_id")
		if err := db.RemoveTrackFromPlaylist(database, playlistID, trackID); err != nil {
			if err.Error() == "not found" {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
