# wallf

Bulk wallpaper downloader with a terminal UI. Searches Wallhaven, Reddit, and Bing for wallpapers and downloads them to a local directory with dedup.

Built as part of [I.A.M](https://github.com/neur0map/project-i-a-m) — an opinionated Arch Linux rice for cybersecurity students.

## What it does

- Step-by-step search: query, count, resolution, color filter
- Shows how many results are available on the server before you pick a count
- Downloads wallpapers sequentially with a progress bar and live status feed
- SHA-256 content dedup so you never save the same image twice
- Catppuccin-inspired color scheme

### Sources

| Source | Key required | Notes |
|--------|-------------|-------|
| Wallhaven | No | Full text search, color filtering, resolution filtering |
| Reddit | No | Pulls from r/wallpapers, r/wallpaper, r/unixporn |
| Bing | No | Daily curated wallpapers (up to 8) |

## Install

Requires Go 1.21+.

```
go install github.com/neur0map/wallf@latest
```

Or build from source:

```
git clone https://github.com/neur0map/wallf
cd wallf
go build -o wallf .
```

## Usage

```
wallf
```

That's it. The TUI walks you through everything.

On first run, a setup wizard asks for your download directory and preferred resolution. Config is saved to `~/.config/wallf/config.toml`.

### Config

```toml
[general]
download_dir = "~/Pictures/Wallpapers"
default_source = "wallhaven"
min_resolution = "2560x1440"
```

## Requirements

- Any terminal emulator
- Go 1.21+ for building

## License

MIT
