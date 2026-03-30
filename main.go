package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neur0map/wallf/config"
	"github.com/neur0map/wallf/download"
	"github.com/neur0map/wallf/source"
	"github.com/neur0map/wallf/tui"
)

func main() {
	srcFlag := flag.String("s", "", "source (wallhaven, reddit, bing)")
	countFlag := flag.Int("n", 10, "number of results")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "wallf -- wallpaper fetcher\n\nUsage: wallf [query] [-s source] [-n count]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	needWizard := !config.Exists()
	var cfg config.Config
	if !needWizard {
		cfg, _ = config.Load(config.Path())
	}
	if cfg == (config.Config{}) {
		cfg = config.Default()
	}

	os.MkdirAll(cfg.DownloadDir(), 0755)
	hashIndex, _ := download.BuildHashIndex(cfg.DownloadDir())

	sources := map[string]source.Source{
		"wallhaven": source.NewWallhaven(""),
		"reddit":    source.NewReddit(),
		"bing":      source.NewBing(),
	}

	query := strings.TrimSpace(strings.Join(flag.Args(), " "))

	opts := tui.AppOpts{
		NeedWizard: needWizard,
		Config:     cfg,
		Sources:    sources,
		Query:      query,
		Source:     *srcFlag,
		Count:      *countFlag,
		Downloader: download.New(cfg.DownloadDir(), hashIndex),
	}

	p := tea.NewProgram(tui.NewApp(opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
