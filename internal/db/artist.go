package db

import (
	"database/sql"
	"fmt"
)

type ArtistFilter struct {
	Genre  string
	Limit  int
	Offset int
}

func ListArtists(db *sql.DB, f ArtistFilter) ([]Artist, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	where, args := "", []any{}
	if f.Genre != "" {
		where = "WHERE genre = ?"
		args = append(args, f.Genre)
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM artists "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := db.Query("SELECT id, name, genre FROM artists "+where+" ORDER BY name LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var artists []Artist
	for rows.Next() {
		var a Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.Genre); err != nil {
			return nil, 0, err
		}
		artists = append(artists, a)
	}
	return artists, total, rows.Err()
}

func GetArtist(db *sql.DB, id string) (*Artist, error) {
	var a Artist
	err := db.QueryRow("SELECT id, name, genre FROM artists WHERE id = ?", id).Scan(&a.ID, &a.Name, &a.Genre)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func InsertArtist(db *sql.DB, a Artist) error {
	_, err := db.Exec("INSERT INTO artists (id, name, genre) VALUES (?, ?, ?)", a.ID, a.Name, a.Genre)
	return err
}

func UpdateArtist(db *sql.DB, a Artist) error {
	res, err := db.Exec("UPDATE artists SET name=?, genre=? WHERE id=?", a.Name, a.Genre, a.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func DeleteArtist(db *sql.DB, id string) error {
	res, err := db.Exec("DELETE FROM artists WHERE id=?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func ListArtistTracks(db *sql.DB, artistID string, limit, offset int) ([]TrackDetail, int, error) {
	return listTrackDetails(db, TrackFilter{ArtistID: artistID, Limit: limit, Offset: offset})
}

// BulkInsertArtists inserts artists in batches within a transaction.
func BulkInsertArtists(db *sql.DB, artists []Artist) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO artists (id, name, genre) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, a := range artists {
		if _, err := stmt.Exec(a.ID, a.Name, a.Genre); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
