package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lehmann314159/mockspotify/internal/db"
	"github.com/lehmann314159/mockspotify/internal/seed"
)

func main() {
	dbPath := flag.String("db", "./data/mockspotify.db", "path to SQLite file")
	genresFlag := flag.String("genres", "pop,rock,jazz,electronic,classical", "comma-separated genres")
	tracksPerGenre := flag.Int("tracks-per-genre", 2000, "target track count per genre")
	seedVal := flag.Int64("seed", 0, "random seed (0 = time-based)")
	clear := flag.Bool("clear", false, "drop and recreate all data")
	flag.Parse()

	if *seedVal == 0 {
		*seedVal = time.Now().UnixNano()
	}

	genres := strings.Split(*genresFlag, ",")
	for i, g := range genres {
		genres[i] = strings.TrimSpace(g)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if *clear {
		tables := []string{"playlist_tracks", "playlists", "tracks", "albums", "artists"}
		for _, t := range tables {
			if _, err := database.Exec("DELETE FROM " + t); err != nil {
				fmt.Fprintf(os.Stderr, "clear %s: %v\n", t, err)
				os.Exit(1)
			}
		}
		fmt.Println("Cleared all data.")
	}

	cfg := seed.Config{
		Genres:         genres,
		TracksPerGenre: *tracksPerGenre,
		Seed:           *seedVal,
	}

	fmt.Printf("Seeding with seed=%d, genres=%v, tracks/genre=%d\n", *seedVal, genres, *tracksPerGenre)

	summaries, err := seed.Run(database, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	ok := seed.PrintSummary(summaries)
	if !ok {
		fmt.Fprintln(os.Stderr, "\nValidation failed: one or more Camelot slots below minimum of 80.")
		os.Exit(1)
	}
	fmt.Println("\nSeed complete.")
}
