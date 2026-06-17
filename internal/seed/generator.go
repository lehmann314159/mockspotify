package seed

import (
	"embed"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/lehmann314159/mockspotify/internal/db"
)

//go:embed words/*.txt
var wordFS embed.FS

type Config struct {
	Genres          []string
	TracksPerGenre  int
	Seed            int64
	ArtistsPerGenre int // derived: 50
	AlbumsPerArtist int // derived: 4
}

type genreParams struct {
	meanBPM   float64
	stddevBPM float64
}

var genreDefaults = map[string]genreParams{
	"pop":        {118, 15},
	"rock":       {130, 20},
	"jazz":       {145, 30},
	"electronic": {128, 12},
	"classical":  {90, 25},
}

func defaultBPMParams(genre string) genreParams {
	if p, ok := genreDefaults[genre]; ok {
		return p
	}
	return genreParams{120, 15}
}

// loadWords reads a word list from the embedded FS, one word per line, stripping blank lines.
func loadWords(filename string) ([]string, error) {
	data, err := wordFS.ReadFile("words/" + filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var words []string
	for _, l := range lines {
		w := strings.TrimSpace(l)
		if w != "" {
			words = append(words, w)
		}
	}
	return words, nil
}

// Generate produces all entities for the given config and writes them to the db.
// Returns a summary table for printing.
func Generate(database interface {
	// accepts *sql.DB but avoids importing database/sql here
}, cfg Config) ([]Summary, error) {
	panic("use RunGenerate")
}

// Summary holds per-genre stats for the validation table.
type Summary struct {
	Genre     string
	Artists   int
	Albums    int
	Tracks    int
	Playlists int
	MinSlot   int
	MaxSlot   int
}

// wordBank holds all word lists.
type wordBank struct {
	adjectives []string
	nouns      []string
	titleWords []string
}

func loadWordBank() (*wordBank, error) {
	adj, err := loadWords("adjectives.txt")
	if err != nil {
		return nil, fmt.Errorf("adjectives: %w", err)
	}
	nouns, err := loadWords("nouns.txt")
	if err != nil {
		return nil, fmt.Errorf("nouns: %w", err)
	}
	titles, err := loadWords("title_words.txt")
	if err != nil {
		return nil, fmt.Errorf("title_words: %w", err)
	}
	return &wordBank{adj, nouns, titles}, nil
}

func (wb *wordBank) artistName(r *rand.Rand) string {
	return wb.adjectives[r.Intn(len(wb.adjectives))] + " " + wb.nouns[r.Intn(len(wb.nouns))]
}

func (wb *wordBank) albumTitle(r *rand.Rand) string {
	n := 1 + r.Intn(3)
	words := make([]string, n)
	for i := range words {
		words[i] = wb.titleWords[r.Intn(len(wb.titleWords))]
	}
	return strings.Join(words, " ")
}

func (wb *wordBank) trackTitle(r *rand.Rand) string {
	n := 2 + r.Intn(3) // 2–4 words
	words := make([]string, n)
	for i := range words {
		words[i] = wb.titleWords[r.Intn(len(wb.titleWords))]
	}
	return strings.Join(words, " ")
}

func clampBPM(v float64) float64 {
	if v < 60 {
		return 60
	}
	if v > 200 {
		return 200
	}
	return v
}

// buildWeightedCamelotPicker returns cumulative weights and parallel slot indices.
func buildWeightedCamelotPicker() (codes []string, cumWeights []float64) {
	var cum float64
	for _, s := range camelotSlots {
		cum += s.Weight
		codes = append(codes, s.Code)
		cumWeights = append(cumWeights, cum)
	}
	return
}

func pickCamelot(r *rand.Rand, codes []string, cumWeights []float64) string {
	total := cumWeights[len(cumWeights)-1]
	v := r.Float64() * total
	for i, cw := range cumWeights {
		if v <= cw {
			return codes[i]
		}
	}
	return codes[len(codes)-1]
}

// generateTracksForGenre generates tracks with weighted Camelot distribution,
// then top-fills slots below the floor.
func generateTracksForGenre(r *rand.Rand, wb *wordBank, genre string, artists []db.Artist, albums []db.Album, total int) []db.Track {
	params := defaultBPMParams(genre)
	codes, cumWeights := buildWeightedCamelotPicker()
	tracksPerAlbum := total / len(albums)

	tracks := make([]db.Track, 0, total)
	trackNum := 0

	// First pass: distribute naturally
	for ai, album := range albums {
		count := tracksPerAlbum
		if ai == len(albums)-1 {
			count = total - len(tracks) // absorb remainder in last album
		}
		for i := 0; i < count; i++ {
			trackNum++
			code := pickCamelot(r, codes, cumWeights)
			key := camelotByCode[code].KeyOf
			bpm := clampBPM(r.NormFloat64()*params.stddevBPM + params.meanBPM)
			dur := 120000 + r.Intn(240001) // [120000, 360000]
			tracks = append(tracks, db.Track{
				ID:          fmt.Sprintf("trk_%s_%05d", genre[:3], trackNum),
				AlbumID:     album.ID,
				ArtistID:    album.ArtistID,
				Title:       wb.trackTitle(r),
				DurationMS:  dur,
				Tempo:       math.Round(bpm*10) / 10,
				KeyOf:       key,
				CamelotCode: code,
				Genre:       genre,
				TrackNumber: (i % 10) + 1,
			})
		}
	}

	// Second pass: enforce minimum 80 per slot
	const minPerSlot = 80
	slotCount := map[string]int{}
	for _, t := range tracks {
		slotCount[t.CamelotCode]++
	}

	for _, slot := range camelotSlots {
		deficit := minPerSlot - slotCount[slot.Code]
		if deficit <= 0 {
			continue
		}
		// top-fill: append extra tracks spread across albums
		for i := 0; i < deficit; i++ {
			trackNum++
			album := albums[trackNum%len(albums)]
			bpm := clampBPM(r.NormFloat64()*params.stddevBPM + params.meanBPM)
			dur := 120000 + r.Intn(240001)
			tracks = append(tracks, db.Track{
				ID:          fmt.Sprintf("trk_%s_%05d", genre[:3], trackNum),
				AlbumID:     album.ID,
				ArtistID:    album.ArtistID,
				Title:       wb.trackTitle(r),
				DurationMS:  dur,
				Tempo:       math.Round(bpm*10) / 10,
				KeyOf:       slot.KeyOf,
				CamelotCode: slot.Code,
				Genre:       genre,
				TrackNumber: (i % 10) + 1,
			})
		}
	}

	return tracks
}

func generatePlaylistsForGenre(r *rand.Rand, wb *wordBank, genre string, tracks []db.Track, count int) ([]db.Playlist, []db.PlaylistTrack) {
	playlists := make([]db.Playlist, count)
	for i := range playlists {
		playlists[i] = db.Playlist{
			ID:          fmt.Sprintf("pl_%s_%03d", genre[:3], i+1),
			Name:        wb.adjectives[r.Intn(len(wb.adjectives))] + " " + strings.Title(genre) + " Mix " + fmt.Sprintf("%d", i+1),
			Genre:       genre,
			Description: "A curated " + genre + " playlist.",
		}
	}

	// Assign each track to 1–3 playlists
	positionOf := make([]int, count) // next position per playlist
	for i := range positionOf {
		positionOf[i] = 1
	}

	var pts []db.PlaylistTrack
	for _, track := range tracks {
		n := 1 + r.Intn(3) // assign to 1–3 playlists
		chosen := r.Perm(count)[:min(n, count)]
		for _, idx := range chosen {
			pts = append(pts, db.PlaylistTrack{
				PlaylistID: playlists[idx].ID,
				TrackID:    track.ID,
				Position:   positionOf[idx],
			})
			positionOf[idx]++
		}
	}

	return playlists, pts
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
