package db

const schema = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS artists (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    genre TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS albums (
    id           TEXT PRIMARY KEY,
    artist_id    TEXT NOT NULL REFERENCES artists(id),
    title        TEXT NOT NULL,
    release_year INTEGER,
    genre        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tracks (
    id           TEXT PRIMARY KEY,
    album_id     TEXT NOT NULL REFERENCES albums(id),
    artist_id    TEXT NOT NULL REFERENCES artists(id),
    title        TEXT NOT NULL,
    duration_ms  INTEGER NOT NULL,
    tempo        REAL NOT NULL,
    key_of       TEXT NOT NULL,
    camelot_code TEXT NOT NULL,
    genre        TEXT NOT NULL,
    track_number INTEGER
);

CREATE TABLE IF NOT EXISTS playlists (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    genre       TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id TEXT    NOT NULL REFERENCES playlists(id),
    track_id    TEXT    NOT NULL REFERENCES tracks(id),
    position    INTEGER NOT NULL,
    PRIMARY KEY (playlist_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_tracks_genre_camelot ON tracks(genre, camelot_code);
CREATE INDEX IF NOT EXISTS idx_tracks_genre_tempo   ON tracks(genre, tempo);
CREATE INDEX IF NOT EXISTS idx_tracks_genre         ON tracks(genre);
CREATE INDEX IF NOT EXISTS idx_tracks_title_artist  ON tracks(title, artist_id);
`