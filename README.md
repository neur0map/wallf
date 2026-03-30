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

### Binary (recommended)

Grab a prebuilt binary from [releases](https://github.com/neur0map/wallf/releases):

```
# linux amd64
curl -L https://github.com/neur0map/wallf/releases/latest/download/wallf_linux_amd64.tar.gz | tar xz
sudo mv wallf /usr/local/bin/

# linux arm64
curl -L https://github.com/neur0map/wallf/releases/latest/download/wallf_linux_arm64.tar.gz | tar xz
sudo mv wallf /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/neur0map/wallf/releases/latest/download/wallf_darwin_arm64.tar.gz | tar xz
sudo mv wallf /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/neur0map/wallf/releases/latest/download/wallf_darwin_amd64.tar.gz | tar xz
sudo mv wallf /usr/local/bin/
```

### Go install

Requires Go 1.21+.

```
go install github.com/neur0map/wallf@latest
```

### From source

```
git clone https://github.com/neur0map/wallf
cd wallf
go build -o wallf .
```

## Usage

```
wallf
```

The TUI walks you through everything.

On first run, a setup wizard asks for your download directory and preferred resolution. Config is saved to `~/.config/wallf/config.toml` (respects `$XDG_CONFIG_HOME`).

### Config

```toml
[general]
download_dir = "~/Pictures/Wallpapers"
default_source = "wallhaven"
min_resolution = "2560x1440"
```

## License

MIT
