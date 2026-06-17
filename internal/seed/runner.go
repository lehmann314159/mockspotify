package seed

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"

	"github.com/lehmann314159/mockspotify/internal/db"
)

// Run executes the full seed pipeline and returns per-genre summaries.
func Run(database *sql.DB, cfg Config) ([]Summary, error) {
	wb, err := loadWordBank()
	if err != nil {
		return nil, fmt.Errorf("load words: %w", err)
	}

	r := rand.New(rand.NewSource(cfg.Seed))

	artistsPerGenre := cfg.ArtistsPerGenre
	if artistsPerGenre <= 0 {
		artistsPerGenre = 50
	}
	albumsPerArtist := cfg.AlbumsPerArtist
	if albumsPerArtist <= 0 {
		albumsPerArtist = 4
	}

	var summaries []Summary

	for _, genre := range cfg.Genres {
		// Generate artists
		artists := make([]db.Artist, artistsPerGenre)
		for i := range artists {
			artists[i] = db.Artist{
				ID:    fmt.Sprintf("art_%s_%03d", genre[:3], i+1),
				Name:  wb.artistName(r),
				Genre: genre,
			}
		}
		if err := db.BulkInsertArtists(database, artists); err != nil {
			return nil, fmt.Errorf("insert artists %s: %w", genre, err)
		}

		// Generate albums
		albums := make([]db.Album, 0, artistsPerGenre*albumsPerArtist)
		albumIdx := 0
		for _, artist := range artists {
			for j := 0; j < albumsPerArtist; j++ {
				albumIdx++
				year := 1980 + r.Intn(45) // 1980–2024
				albums = append(albums, db.Album{
					ID:          fmt.Sprintf("alb_%s_%04d", genre[:3], albumIdx),
					ArtistID:    artist.ID,
					Title:       wb.albumTitle(r),
					ReleaseYear: year,
					Genre:       genre,
				})
			}
		}
		if err := db.BulkInsertAlbums(database, albums); err != nil {
			return nil, fmt.Errorf("insert albums %s: %w", genre, err)
		}

		// Generate tracks
		tracks := generateTracksForGenre(r, wb, genre, artists, albums, cfg.TracksPerGenre)
		if err := db.BulkInsertTracks(database, tracks); err != nil {
			return nil, fmt.Errorf("insert tracks %s: %w", genre, err)
		}

		// Generate playlists and assignments
		playlists, pts := generatePlaylistsForGenre(r, wb, genre, tracks, 20)
		if err := db.BulkInsertPlaylists(database, playlists); err != nil {
			return nil, fmt.Errorf("insert playlists %s: %w", genre, err)
		}
		if err := db.BulkInsertPlaylistTracks(database, pts); err != nil {
			return nil, fmt.Errorf("insert playlist_tracks %s: %w", genre, err)
		}

		// Compute summary
		slotCount, err := db.CountByGenreAndCamelot(database, genre)
		if err != nil {
			return nil, fmt.Errorf("count camelot %s: %w", genre, err)
		}
		minSlot, maxSlot := int(^uint(0)>>1), 0
		for _, slot := range camelotSlots {
			c := slotCount[slot.Code]
			if c < minSlot {
				minSlot = c
			}
			if c > maxSlot {
				maxSlot = c
			}
		}

		summaries = append(summaries, Summary{
			Genre:     genre,
			Artists:   len(artists),
			Albums:    len(albums),
			Tracks:    len(tracks),
			Playlists: len(playlists),
			MinSlot:   minSlot,
			MaxSlot:   maxSlot,
		})
	}

	return summaries, nil
}

// PrintSummary prints the validation table to stdout.
func PrintSummary(summaries []Summary) bool {
	fmt.Printf("%-12s  %-7s  %-6s  %-6s  %-9s  %-8s  %-8s\n",
		"Genre", "Artists", "Albums", "Tracks", "Playlists", "Min/slot", "Max/slot")
	fmt.Println(strings.Repeat("-", 70))
	ok := true
	for _, s := range summaries {
		marker := ""
		if s.MinSlot < 80 {
			marker = " WARNING: slot below 80"
			ok = false
		}
		fmt.Printf("%-12s  %-7d  %-6d  %-6d  %-9d  %-8d  %-8d%s\n",
			s.Genre, s.Artists, s.Albums, s.Tracks, s.Playlists, s.MinSlot, s.MaxSlot, marker)
	}
	return ok
}
