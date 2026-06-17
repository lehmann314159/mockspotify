package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lehmann314159/mockspotify/internal/db"
)


func listTracks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f := db.TrackFilter{
			Genre:       q.Get("genre"),
			CamelotCode: q.Get("camelot_code"),
			Title:       q.Get("title"),
			Artist:      q.Get("artist"),
			Limit:       queryInt(r, "limit", 20),
			Offset:      queryInt(r, "offset", 0),
		}
		tracks, total, err := db.ListTracks(database, f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": tracks, "total": total})
	}
}

func getTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		t, err := db.GetTrack(database, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, t)
	}
}

func getTrackAudio(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		t, err := db.GetTrack(database, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":           t.ID,
			"tempo":        t.Tempo,
			"key_of":       t.KeyOf,
			"camelot_code": t.CamelotCode,
		})
	}
}

func createTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t db.Track
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := db.InsertTrack(database, t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)
	}
}

func updateTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t db.Track
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		t.ID = r.PathValue("id")
		if err := db.UpdateTrack(database, t); err != nil {
			if err.Error() == "not found" {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	}
}

func deleteTrack(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := db.DeleteTrack(database, id); err != nil {
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
