package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lehmann314159/mockspotify/internal/db"
)


func listAlbums(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := db.AlbumFilter{
			ArtistID: r.URL.Query().Get("artist_id"),
			Genre:    r.URL.Query().Get("genre"),
			Limit:    queryInt(r, "limit", 20),
			Offset:   queryInt(r, "offset", 0),
		}
		albums, total, err := db.ListAlbums(database, f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"albums": albums, "total": total})
	}
}

func getAlbum(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		a, err := db.GetAlbum(database, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if a == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, a)
	}
}

func createAlbum(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a db.Album
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := db.InsertAlbum(database, a); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

func updateAlbum(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a db.Album
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		a.ID = r.PathValue("id")
		if err := db.UpdateAlbum(database, a); err != nil {
			if err.Error() == "not found" {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, a)
	}
}

func deleteAlbum(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := db.DeleteAlbum(database, id); err != nil {
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

func getAlbumTracks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		tracks, total, err := db.ListAlbumTracks(database, id, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": tracks, "total": total})
	}
}
