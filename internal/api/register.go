package api

import (
	"database/sql"
	"net/http"
)

// Register adds all authenticated API routes to mux.
// The caller is responsible for wrapping mux with AuthMiddleware.
// The /api/token endpoint must be registered separately without auth.
func Register(mux *http.ServeMux, database *sql.DB) {
	// Artists
	mux.HandleFunc("GET /api/artists", listArtists(database))
	mux.HandleFunc("POST /api/artists", createArtist(database))
	mux.HandleFunc("GET /api/artists/{id}", getArtist(database))
	mux.HandleFunc("PUT /api/artists/{id}", updateArtist(database))
	mux.HandleFunc("DELETE /api/artists/{id}", deleteArtist(database))
	mux.HandleFunc("GET /api/artists/{id}/tracks", getArtistTracks(database))

	// Albums
	mux.HandleFunc("GET /api/albums", listAlbums(database))
	mux.HandleFunc("POST /api/albums", createAlbum(database))
	mux.HandleFunc("GET /api/albums/{id}", getAlbum(database))
	mux.HandleFunc("PUT /api/albums/{id}", updateAlbum(database))
	mux.HandleFunc("DELETE /api/albums/{id}", deleteAlbum(database))
	mux.HandleFunc("GET /api/albums/{id}/tracks", getAlbumTracks(database))

	// Tracks
	mux.HandleFunc("GET /api/tracks", listTracks(database))
	mux.HandleFunc("POST /api/tracks", createTrack(database))
	mux.HandleFunc("GET /api/tracks/{id}", getTrack(database))
	mux.HandleFunc("GET /api/tracks/{id}/audio", getTrackAudio(database))
	mux.HandleFunc("PUT /api/tracks/{id}", updateTrack(database))
	mux.HandleFunc("DELETE /api/tracks/{id}", deleteTrack(database))

	// Playlists
	mux.HandleFunc("GET /api/playlists", listPlaylists(database))
	mux.HandleFunc("POST /api/playlists", createPlaylist(database))
	mux.HandleFunc("GET /api/playlists/{id}", getPlaylist(database))
	mux.HandleFunc("PUT /api/playlists/{id}", updatePlaylist(database))
	mux.HandleFunc("DELETE /api/playlists/{id}", deletePlaylist(database))
	mux.HandleFunc("GET /api/playlists/{id}/tracks", getPlaylistTracks(database))
	mux.HandleFunc("POST /api/playlists/{id}/tracks", addPlaylistTrack(database))
	mux.HandleFunc("DELETE /api/playlists/{id}/tracks/{track_id}", removePlaylistTrack(database))
}
