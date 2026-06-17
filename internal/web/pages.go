package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lehmann314159/mockspotify/internal/db"
)

type pageBase struct {
	Title       string
	Page        int
	Limit       int
	TotalPages  int
	BasePath    string
	QueryParams string
}

func calcPages(total, limit int) int {
	if limit == 0 {
		return 1
	}
	p := total / limit
	if total%limit != 0 {
		p++
	}
	return p
}

func queryStr(r *http.Request, keys ...string) string {
	q := r.URL.Query()
	out := ""
	for _, k := range keys {
		v := q.Get(k)
		if v != "" {
			if out != "" {
				out += "&"
			}
			out += k + "=" + v
		}
	}
	return out
}

func pageFromRequest(r *http.Request, limit int) (int, int) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	page := offset/limit + 1
	return page, offset
}

// DashboardHandler serves GET /
func DashboardHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := db.GetGenreStats(database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderPage(w, "dashboard", map[string]any{
			"Title": "Dashboard",
			"Stats": stats,
		})
	}
}

// ArtistsHandler serves GET /artists
func ArtistsPageHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const limit = 50
		genre := r.URL.Query().Get("genre")
		page, offset := pageFromRequest(r, limit)

		artists, total, err := db.ListArtists(database, db.ArtistFilter{
			Genre: genre, Limit: limit, Offset: offset,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title":       "Artists",
			"Artists":     artists,
			"Genre":       genre,
			"Page":        page,
			"Limit":       limit,
			"TotalPages":  calcPages(total, limit),
			"BasePath":    "/artists",
			"QueryParams": queryStr(r, "genre"),
		}

		if isHTMX(r) {
			renderPartial(w, "artists", data)
		} else {
			renderPage(w, "artists", data)
		}
	}
}

// AlbumsPageHandler serves GET /albums
func AlbumsPageHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const limit = 50
		genre := r.URL.Query().Get("genre")
		artistID := r.URL.Query().Get("artist_id")
		page, offset := pageFromRequest(r, limit)

		albums, total, err := db.ListAlbums(database, db.AlbumFilter{
			Genre: genre, ArtistID: artistID, Limit: limit, Offset: offset,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title":       "Albums",
			"Albums":      albums,
			"Genre":       genre,
			"ArtistID":    artistID,
			"Page":        page,
			"Limit":       limit,
			"TotalPages":  calcPages(total, limit),
			"BasePath":    "/albums",
			"QueryParams": queryStr(r, "genre", "artist_id"),
		}

		if isHTMX(r) {
			renderPartial(w, "albums", data)
		} else {
			renderPage(w, "albums", data)
		}
	}
}

// TracksPageHandler serves GET /tracks
func TracksPageHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const limit = 50
		q := r.URL.Query()
		genre := q.Get("genre")
		camelot := q.Get("camelot_code")
		page, offset := pageFromRequest(r, limit)

		tracks, total, err := db.ListTracks(database, db.TrackFilter{
			Genre:       genre,
			CamelotCode: camelot,
			Limit:       limit,
			Offset:      offset,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title":       "Tracks",
			"Tracks":      tracks,
			"Genre":       genre,
			"CamelotCode": camelot,
			"BPMMin":      q.Get("bpm_min"),
			"BPMMax":      q.Get("bpm_max"),
			"Page":        page,
			"Limit":       limit,
			"TotalPages":  calcPages(total, limit),
			"BasePath":    "/tracks",
			"QueryParams": queryStr(r, "genre", "camelot_code", "bpm_min", "bpm_max"),
		}

		if isHTMX(r) {
			renderPartial(w, "tracks", data)
		} else {
			renderPage(w, "tracks", data)
		}
	}
}

type playlistWithCount struct {
	db.Playlist
	TrackCount int
}

// PlaylistsPageHandler serves GET /playlists
func PlaylistsPageHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const limit = 50
		genre := r.URL.Query().Get("genre")
		page, offset := pageFromRequest(r, limit)

		pls, total, err := db.ListPlaylists(database, db.PlaylistFilter{
			Genre: genre, Limit: limit, Offset: offset,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		withCounts := make([]playlistWithCount, len(pls))
		for i, p := range pls {
			c, _ := db.PlaylistTrackCount(database, p.ID)
			withCounts[i] = playlistWithCount{p, c}
		}

		data := map[string]any{
			"Title":       "Playlists",
			"Playlists":   withCounts,
			"Genre":       genre,
			"Page":        page,
			"Limit":       limit,
			"TotalPages":  calcPages(total, limit),
			"BasePath":    "/playlists",
			"QueryParams": queryStr(r, "genre"),
		}

		if isHTMX(r) {
			renderPartial(w, "playlists", data)
		} else {
			renderPage(w, "playlists", data)
		}
	}
}

// PlaylistDetailHandler serves GET /playlists/{id}
func PlaylistDetailHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const limit = 100
		id := r.PathValue("id")
		pl, err := db.GetPlaylist(database, id)
		if err != nil || pl == nil {
			http.NotFound(w, r)
			return
		}

		page, offset := pageFromRequest(r, limit)
		items, total, err := db.GetPlaylistTracks(database, id, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		renderPage(w, "playlist_detail", map[string]any{
			"Title":       fmt.Sprintf("Playlist: %s", pl.Name),
			"Playlist":    pl,
			"Items":       items,
			"Total":       total,
			"Page":        page,
			"Limit":       limit,
			"TotalPages":  calcPages(total, limit),
			"BasePath":    "/playlists/" + id,
			"QueryParams": "",
		})
	}
}
