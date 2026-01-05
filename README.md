# stremio-tui

Terminal UI for searching and streaming content via Stremio addons.

<img src="https://raw.githubusercontent.com/rshero/stremio-tui/refs/heads/master/demo.gif" alt="Demo" height="512" />

## Requirements

- Go 1.21+
- mpv (for playback)

## Install

```bash
go build -o stremio-tui
```

## Run

```bash
./stremio-tui
```

## Config

Set environment variables to override defaults:

```bash
export IMDB_API_URL="https://your-imdb-api-url"
export ALC_ADDON_URL="your-addon-url"
```

## Keys

| Key | Action |
|-----|--------|
| `Enter` | Select / Play |
| `Esc` | Go back |
| `/` | Filter list |
| `p` | Play stream |
| `d` | Download stream |
| `j/k` | Navigate |
| `q` | Quit |

## Downloads

Files save to `./downloads/` in current directory.
