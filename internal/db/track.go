package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type TrackFilter struct {
	Genre       string
	CamelotCode string
	Title       string
	Artist      string
	ArtistID    string
	AlbumID     string
	Limit       int
	Offset      int
}

func listTrackDetails(db *sql.DB, f TrackFilter) ([]TrackDetail, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	conds, args := []string{}, []any{}
	if f.Genre != "" {
		conds = append(conds, "t.genre = ?")
		args = append(args, f.Genre)
	}
	if f.CamelotCode != "" {
		conds = append(conds, "t.camelot_code = ?")
		args = append(args, f.CamelotCode)
	}
	if f.Title != "" {
		conds = append(conds, "t.title LIKE ?")
		args = append(args, "%"+f.Title+"%")
	}
	if f.Artist != "" {
		conds = append(conds, "ar.name LIKE ?")
		args = append(args, "%"+f.Artist+"%")
	}
	if f.ArtistID != "" {
		conds = append(conds, "t.artist_id = ?")
		args = append(args, f.ArtistID)
	}
	if f.AlbumID != "" {
		conds = append(conds, "t.album_id = ?")
		args = append(args, f.AlbumID)
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	base := `FROM tracks t
		JOIN artists ar ON ar.id = t.artist_id
		JOIN albums al  ON al.id = t.album_id `

	var total int
	if err := db.QueryRow("SELECT COUNT(*) "+base+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := db.Query(`SELECT t.id, t.album_id, t.artist_id, t.title, t.duration_ms,
		t.tempo, t.key_of, t.camelot_code, t.genre, t.track_number,
		ar.name, al.title `+base+where+` ORDER BY t.title LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tracks []TrackDetail
	for rows.Next() {
		var td TrackDetail
		if err := rows.Scan(
			&td.ID, &td.AlbumID, &td.ArtistID, &td.Title, &td.DurationMS,
			&td.Tempo, &td.KeyOf, &td.CamelotCode, &td.Genre, &td.TrackNumber,
			&td.ArtistName, &td.AlbumTitle,
		); err != nil {
			return nil, 0, err
		}
		tracks = append(tracks, td)
	}
	return tracks, total, rows.Err()
}

func ListTracks(db *sql.DB, f TrackFilter) ([]TrackDetail, int, error) {
	return listTrackDetails(db, f)
}

func GetTrack(db *sql.DB, id string) (*TrackDetail, error) {
	rows, err := db.Query(`SELECT t.id, t.album_id, t.artist_id, t.title, t.duration_ms,
		t.tempo, t.key_of, t.camelot_code, t.genre, t.track_number,
		ar.name, al.title
		FROM tracks t
		JOIN artists ar ON ar.id = t.artist_id
		JOIN albums al  ON al.id = t.album_id
		WHERE t.id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var td TrackDetail
	if err := rows.Scan(
		&td.ID, &td.AlbumID, &td.ArtistID, &td.Title, &td.DurationMS,
		&td.Tempo, &td.KeyOf, &td.CamelotCode, &td.Genre, &td.TrackNumber,
		&td.ArtistName, &td.AlbumTitle,
	); err != nil {
		return nil, err
	}
	return &td, rows.Err()
}

func InsertTrack(db *sql.DB, t Track) error {
	_, err := db.Exec(`INSERT INTO tracks (id, album_id, artist_id, title, duration_ms, tempo, key_of, camelot_code, genre, track_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.AlbumID, t.ArtistID, t.Title, t.DurationMS, t.Tempo, t.KeyOf, t.CamelotCode, t.Genre, t.TrackNumber)
	return err
}

func UpdateTrack(db *sql.DB, t Track) error {
	res, err := db.Exec(`UPDATE tracks SET album_id=?, artist_id=?, title=?, duration_ms=?, tempo=?,
		key_of=?, camelot_code=?, genre=?, track_number=? WHERE id=?`,
		t.AlbumID, t.ArtistID, t.Title, t.DurationMS, t.Tempo, t.KeyOf, t.CamelotCode, t.Genre, t.TrackNumber, t.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func DeleteTrack(db *sql.DB, id string) error {
	res, err := db.Exec("DELETE FROM tracks WHERE id=?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func BulkInsertTracks(db *sql.DB, tracks []Track) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO tracks
		(id, album_id, artist_id, title, duration_ms, tempo, key_of, camelot_code, genre, track_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, t := range tracks {
		if _, err := stmt.Exec(t.ID, t.AlbumID, t.ArtistID, t.Title, t.DurationMS, t.Tempo, t.KeyOf, t.CamelotCode, t.Genre, t.TrackNumber); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// CountByGenreAndCamelot returns a map of camelot_code -> count for the given genre.
func CountByGenreAndCamelot(db *sql.DB, genre string) (map[string]int, error) {
	rows, err := db.Query("SELECT camelot_code, COUNT(*) FROM tracks WHERE genre=? GROUP BY camelot_code", genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]int{}
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		m[code] = count
	}
	return m, rows.Err()
}
