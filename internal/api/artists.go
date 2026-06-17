package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lehmann314159/mockspotify/internal/db"
)


func listArtists(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := db.ArtistFilter{
			Genre:  r.URL.Query().Get("genre"),
			Limit:  queryInt(r, "limit", 20),
			Offset: queryInt(r, "offset", 0),
		}
		artists, total, err := db.ListArtists(database, f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"artists": artists, "total": total})
	}
}

func getArtist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		a, err := db.GetArtist(database, id)
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

func createArtist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a db.Artist
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := db.InsertArtist(database, a); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

func updateArtist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a db.Artist
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		a.ID = r.PathValue("id")
		if err := db.UpdateArtist(database, a); err != nil {
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

func deleteArtist(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := db.DeleteArtist(database, id); err != nil {
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

func getArtistTracks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		tracks, total, err := db.ListArtistTracks(database, id, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": tracks, "total": total})
	}
}
