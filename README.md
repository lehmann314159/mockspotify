# MockSpotify

A self-contained Go HTTP service that serves as a drop-in mock replacement for the Spotify API and GetSongBPM. Stores plausible-looking music metadata in SQLite and exposes a REST API plus a human-browsable web UI.

## Quick start

```sh
# Build
go build -o mockspotify-server ./cmd/server
go build -o mockspotify-seed ./cmd/seed

# Seed the database (2000 tracks per genre, reproducible)
./mockspotify-seed --db ./data/mockspotify.db --seed 42

# Run the server
./mockspotify-server
# → http://localhost:8090
```

## Environment variables

| Variable            | Default                       | Description               |
|---------------------|-------------------------------|---------------------------|
| `MOCKSPOTIFY_DB`    | `./data/mockspotify.db`       | SQLite file path          |
| `MOCKSPOTIFY_PORT`  | `8090`                        | Listen port               |
| `MOCKSPOTIFY_TOKEN` | `mockspotify-dev-token-2026`  | Static bearer token       |

## API authentication

All `/api/*` endpoints (except `POST /api/token`) require:
```
Authorization: Bearer <MOCKSPOTIFY_TOKEN>
```

The token endpoint mimics the Spotify client credentials flow:
```sh
curl -X POST http://localhost:8090/api/token
# → {"access_token":"mockspotify-dev-token-2026","token_type":"Bearer","expires_in":86400}
```

## Key endpoints

```
GET  /api/tracks?genre=pop&camelot_code=8B&title=sky&artist=echo
GET  /api/tracks/{id}/audio          # BPM + key data
GET  /api/playlists/{id}/tracks
GET  /api/artists/{id}/tracks
```

## Seeder flags

```
--db              SQLite path (default: ./data/mockspotify.db)
--genres          Comma-separated genres (default: pop,rock,jazz,electronic,classical)
--tracks-per-genre  Target tracks per genre (default: 2000)
--seed            int64 seed (default: time-based)
--clear           Drop and recreate all data
```

## Docker

```sh
docker compose up -d
# Seed inside the container:
docker compose exec mockspotify ./mockspotify-seed --db /app/data/mockspotify.db --seed 42
```

## Web UI

Browse to `http://localhost:8090` for the dark-themed dashboard with genre stats, artist/album/track/playlist browsers with live HTMX search and filtering.
