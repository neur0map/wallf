package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neur0map/wallf/config"
	"github.com/neur0map/wallf/download"
	"github.com/neur0map/wallf/source"
)

type appState int

const (
	stateWizard appState = iota
	stateSearch
	stateDownload
	stateSummary
)

// AppOpts holds startup configuration.
type AppOpts struct {
	NeedWizard bool
	Config     config.Config
	Sources    map[string]source.Source
	Query      string
	Source     string
	Count      int
	Downloader *download.Downloader
}

// DownloadRecord tracks the outcome of a single wallpaper.
type DownloadRecord struct {
	Result     source.WallpaperResult
	Downloaded bool
	Skipped    bool
	Error      string
	Path       string
}

// App is the root bubbletea model.
type App struct {
	state    appState
	opts     AppOpts
	width    int
	height   int
	wizard   config.WizardModel
	search   SearchModel
	swipe    SwipeModel
	summary  SummaryModel
	loading  bool
	spinner  spinner.Model
}

// --- Messages ---

type searchResultsMsg struct {
	results []source.WallpaperResult
	err     error
}

// NewApp creates the root app model.
func NewApp(opts AppOpts) App {
	initialState := stateSearch
	if opts.NeedWizard {
		initialState = stateWizard
	}

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)

	a := App{
		state:   initialState,
		opts:    opts,
		spinner: sp,
	}

	if initialState == stateWizard {
		a.wizard = config.NewWizard()
	} else {
		a.search = NewSearchModel(opts.Source, opts.Count)
		a.search.SetSources(opts.Sources)
	}

	return a
}

func (a App) Init() tea.Cmd {
	return a.spinner.Tick
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	switch a.state {
	case stateWizard:
		return a.updateWizard(msg)
	case stateSearch:
		return a.updateSearch(msg)
	case stateDownload:
		return a.updateDownload(msg)
	case stateSummary:
		return a.updateSummary(msg)
	}
	return a, nil
}

func (a App) View() string {
	var content string
	switch a.state {
	case stateWizard:
		content = a.wizard.View()
	case stateSearch:
		if a.loading {
			content = stylePanel.Render(
				styleTitle.Render("  wallf") + "\n\n" +
					"  " + a.spinner.View() + styleDim.Render(" searching...") + "\n",
			)
		} else {
			content = a.search.View()
		}
	case stateDownload:
		content = a.swipe.View()
	case stateSummary:
		content = a.summary.View()
	}
	if a.width > 0 && a.height > 0 {
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

// --- State: Wizard ---

func (a App) updateWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case config.WizardDoneMsg:
		cfg := msg.Config
		_ = config.Save(cfg, config.Path())
		a.opts.Config = cfg
		a.search = NewSearchModel(a.opts.Source, a.opts.Count)
		a.search.SetSources(a.opts.Sources)
		a.state = stateSearch
		return a, nil
	default:
		var cmd tea.Cmd
		a.wizard, cmd = a.wizard.Update(msg)
		return a, cmd
	}
}

// --- State: Search ---

func (a App) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchDoneMsg:
		a.loading = true
		return a, a.performSearch(msg.Result)
	case searchResultsMsg:
		a.loading = false
		if msg.err != nil {
			a.swipe = NewSwipeModel(nil)
			a.swipe.status = fmt.Sprintf("Search error: %v", msg.err)
			a.state = stateDownload
			return a, nil
		}
		a.swipe = NewSwipeModel(msg.results)
		a.state = stateDownload
		// Auto-start downloading
		return a, a.swipe.Init()
	default:
		var cmd tea.Cmd
		a.search, cmd = a.search.Update(msg)
		return a, cmd
	}
}

// --- State: Download ---

func (a App) updateDownload(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SwipeDoneMsg:
		if a.swipe.Total() == 0 {
			a.search = NewSearchModel(a.opts.Source, a.opts.Count)
			a.state = stateSearch
			return a, nil
		}
		a.summary = NewSummaryModel(msg.Records, a.opts.Config.DownloadDir())
		a.state = stateSummary
		download.InvalidateSkwdCache(skwdCachePath())
		return a, nil

	case StartDownloadMsg:
		return a, a.downloadItem(msg.Index, msg.Result)

	default:
		var cmd tea.Cmd
		a.swipe, cmd = a.swipe.Update(msg)
		return a, cmd
	}
}

// --- State: Summary ---

func (a App) updateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.summary, cmd = a.summary.Update(msg)
	return a, cmd
}

// --- Commands ---

func (a App) performSearch(sr SearchResult) tea.Cmd {
	sources := a.opts.Sources
	return func() tea.Msg {
		src, ok := sources[sr.Source]
		if !ok {
			return searchResultsMsg{err: fmt.Errorf("unknown source: %s", sr.Source)}
		}
		minRes := sr.MinRes
		if minRes == "" || minRes == "any" {
			minRes = ""
		}
		opts := source.SearchOpts{
			Query:     sr.Query,
			Sort:      sr.Sort,
			Count:     sr.Count,
			MinRes:    minRes,
			Colors:    sr.Colors,
			Subreddit: sr.Subreddit,
		}
		results, err := src.Search(opts)
		if err == nil && sr.Count > 0 && len(results) > sr.Count {
			results = results[:sr.Count]
		}
		return searchResultsMsg{results: results, err: err}
	}
}

func (a App) downloadItem(index int, r source.WallpaperResult) tea.Cmd {
	dl := a.opts.Downloader
	return func() tea.Msg {
		if dl == nil {
			return DownloadProgressMsg{Index: index, Err: fmt.Errorf("downloader not configured")}
		}
		path, err := dl.Fetch(r)
		return DownloadProgressMsg{Index: index, Path: path, Err: err}
	}
}

// --- Helpers ---

func skwdCachePath() string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "skwd", "wallpaper", "checksum.txt")
}

func sourceNames(sources map[string]source.Source) []string {
	ordered := []string{"wallhaven", "reddit", "bing"}
	var names []string
	for _, name := range ordered {
		if _, ok := sources[name]; ok {
			names = append(names, name)
		}
	}
	return names
}
