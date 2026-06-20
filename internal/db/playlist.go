package db

import (
	"database/sql"
	"fmt"
)

type PlaylistFilter struct {
	Genre  string
	Name   string
	Limit  int
	Offset int
}

func ListPlaylists(db *sql.DB, f PlaylistFilter) ([]Playlist, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	conds, args := []string{}, []any{}
	if f.Genre != "" {
		conds = append(conds, "genre = ?")
		args = append(args, f.Genre)
	}
	if f.Name != "" {
		conds = append(conds, "name LIKE ?")
		args = append(args, "%"+f.Name+"%")
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + conds[0]
		for _, c := range conds[1:] {
			where += " AND " + c
		}
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM playlists "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := db.Query("SELECT id, name, genre, COALESCE(description,'') FROM playlists "+where+" ORDER BY name LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pls []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.Genre, &p.Description); err != nil {
			return nil, 0, err
		}
		pls = append(pls, p)
	}
	return pls, total, rows.Err()
}

func GetPlaylist(db *sql.DB, id string) (*Playlist, error) {
	var p Playlist
	err := db.QueryRow("SELECT id, name, genre, COALESCE(description,'') FROM playlists WHERE id=?", id).
		Scan(&p.ID, &p.Name, &p.Genre, &p.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func InsertPlaylist(db *sql.DB, p Playlist) error {
	_, err := db.Exec("INSERT INTO playlists (id, name, genre, description) VALUES (?, ?, ?, ?)",
		p.ID, p.Name, p.Genre, p.Description)
	return err
}

func UpdatePlaylist(db *sql.DB, p Playlist) error {
	res, err := db.Exec("UPDATE playlists SET name=?, genre=?, description=? WHERE id=?",
		p.Name, p.Genre, p.Description, p.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func DeletePlaylist(db *sql.DB, id string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM playlist_tracks WHERE playlist_id=?", id); err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.Exec("DELETE FROM playlists WHERE id=?", id)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		tx.Rollback()
		return fmt.Errorf("not found")
	}
	return tx.Commit()
}

type PlaylistTrackItem struct {
	Position int        `json:"position"`
	Track    TrackDetail `json:"track"`
}

func GetPlaylistTracks(db *sql.DB, playlistID string, limit, offset int) ([]PlaylistTrackItem, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id=?", playlistID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(`SELECT pt.position, t.id, t.album_id, t.artist_id, t.title, t.duration_ms,
		t.tempo, t.key_of, t.camelot_code, t.genre, t.track_number, ar.name, al.title
		FROM playlist_tracks pt
		JOIN tracks t   ON t.id  = pt.track_id
		JOIN artists ar ON ar.id = t.artist_id
		JOIN albums al  ON al.id = t.album_id
		WHERE pt.playlist_id = ?
		ORDER BY pt.position
		LIMIT ? OFFSET ?`, playlistID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []PlaylistTrackItem
	for rows.Next() {
		var item PlaylistTrackItem
		var td TrackDetail
		if err := rows.Scan(
			&item.Position,
			&td.ID, &td.AlbumID, &td.ArtistID, &td.Title, &td.DurationMS,
			&td.Tempo, &td.KeyOf, &td.CamelotCode, &td.Genre, &td.TrackNumber,
			&td.ArtistName, &td.AlbumTitle,
		); err != nil {
			return nil, 0, err
		}
		item.Track = td
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func AddTrackToPlaylist(db *sql.DB, playlistID, trackID string, position int) error {
	_, err := db.Exec("INSERT OR REPLACE INTO playlist_tracks (playlist_id, track_id, position) VALUES (?, ?, ?)",
		playlistID, trackID, position)
	return err
}

func RemoveTrackFromPlaylist(db *sql.DB, playlistID, trackID string) error {
	res, err := db.Exec("DELETE FROM playlist_tracks WHERE playlist_id=? AND track_id=?", playlistID, trackID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func BulkInsertPlaylists(db *sql.DB, playlists []Playlist) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO playlists (id, name, genre, description) VALUES (?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, p := range playlists {
		if _, err := stmt.Exec(p.ID, p.Name, p.Genre, p.Description); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func BulkInsertPlaylistTracks(db *sql.DB, pts []PlaylistTrack) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO playlist_tracks (playlist_id, track_id, position) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, pt := range pts {
		if _, err := stmt.Exec(pt.PlaylistID, pt.TrackID, pt.Position); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// PlaylistTrackCount returns count per playlist.
func PlaylistTrackCount(db *sql.DB, playlistID string) (int, error) {
	var n int
	err := db.QueryRow("SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id=?", playlistID).Scan(&n)
	return n, err
}

// GenreStats returns per-genre summary counts.
type GenreStat struct {
	Genre     string
	Artists   int
	Albums    int
	Tracks    int
	Playlists int
}

func GetGenreStats(db *sql.DB) ([]GenreStat, error) {
	rows, err := db.Query(`
		SELECT g.genre,
			(SELECT COUNT(*) FROM artists WHERE genre=g.genre) as artists,
			(SELECT COUNT(*) FROM albums   WHERE genre=g.genre) as albums,
			(SELECT COUNT(*) FROM tracks   WHERE genre=g.genre) as tracks,
			(SELECT COUNT(*) FROM playlists WHERE genre=g.genre) as playlists
		FROM (SELECT DISTINCT genre FROM tracks ORDER BY genre) g`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []GenreStat
	for rows.Next() {
		var s GenreStat
		if err := rows.Scan(&s.Genre, &s.Artists, &s.Albums, &s.Tracks, &s.Playlists); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

