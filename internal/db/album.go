package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type AlbumFilter struct {
	ArtistID string
	Genre    string
	Limit    int
	Offset   int
}

func ListAlbums(db *sql.DB, f AlbumFilter) ([]Album, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	conds, args := []string{}, []any{}
	if f.ArtistID != "" {
		conds = append(conds, "artist_id = ?")
		args = append(args, f.ArtistID)
	}
	if f.Genre != "" {
		conds = append(conds, "genre = ?")
		args = append(args, f.Genre)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM albums "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := db.Query("SELECT id, artist_id, title, release_year, genre FROM albums "+where+" ORDER BY title LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var albums []Album
	for rows.Next() {
		var a Album
		if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.ReleaseYear, &a.Genre); err != nil {
			return nil, 0, err
		}
		albums = append(albums, a)
	}
	return albums, total, rows.Err()
}

func GetAlbum(db *sql.DB, id string) (*Album, error) {
	var a Album
	err := db.QueryRow("SELECT id, artist_id, title, release_year, genre FROM albums WHERE id=?", id).
		Scan(&a.ID, &a.ArtistID, &a.Title, &a.ReleaseYear, &a.Genre)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func InsertAlbum(db *sql.DB, a Album) error {
	_, err := db.Exec("INSERT INTO albums (id, artist_id, title, release_year, genre) VALUES (?, ?, ?, ?, ?)",
		a.ID, a.ArtistID, a.Title, a.ReleaseYear, a.Genre)
	return err
}

func UpdateAlbum(db *sql.DB, a Album) error {
	res, err := db.Exec("UPDATE albums SET artist_id=?, title=?, release_year=?, genre=? WHERE id=?",
		a.ArtistID, a.Title, a.ReleaseYear, a.Genre, a.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func DeleteAlbum(db *sql.DB, id string) error {
	res, err := db.Exec("DELETE FROM albums WHERE id=?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func ListAlbumTracks(db *sql.DB, albumID string, limit, offset int) ([]TrackDetail, int, error) {
	return listTrackDetails(db, TrackFilter{AlbumID: albumID, Limit: limit, Offset: offset})
}

func BulkInsertAlbums(db *sql.DB, albums []Album) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO albums (id, artist_id, title, release_year, genre) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, a := range albums {
		if _, err := stmt.Exec(a.ID, a.ArtistID, a.Title, a.ReleaseYear, a.Genre); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
