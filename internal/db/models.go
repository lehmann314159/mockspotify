package db

type Artist struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Genre string `json:"genre"`
}

type Album struct {
	ID          string `json:"id"`
	ArtistID    string `json:"artist_id"`
	Title       string `json:"title"`
	ReleaseYear int    `json:"release_year"`
	Genre       string `json:"genre"`
}

type Track struct {
	ID          string  `json:"id"`
	AlbumID     string  `json:"album_id"`
	ArtistID    string  `json:"artist_id"`
	Title       string  `json:"title"`
	DurationMS  int     `json:"duration_ms"`
	Tempo       float64 `json:"tempo"`
	KeyOf       string  `json:"key_of"`
	CamelotCode string  `json:"camelot_code"`
	Genre       string  `json:"genre"`
	TrackNumber int     `json:"track_number"`
}

type Playlist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
}

type PlaylistTrack struct {
	PlaylistID string `json:"playlist_id"`
	TrackID    string `json:"track_id"`
	Position   int    `json:"position"`
}

// TrackDetail joins track with artist and album names for API responses.
type TrackDetail struct {
	Track
	ArtistName string `json:"artist"`
	AlbumTitle string `json:"album"`
}
