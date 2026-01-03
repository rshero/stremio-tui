package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	apiutils "github.com/rshero/stremio-tui/utils"
)

type View int

const (
	SearchView View = iota
	ResultsView
	SeasonsView
	EpisodesView
	StreamsView
	DownloadView
)

// List item implementations
type imdbItem struct {
	result apiutils.ImdbSearchResult
}

func (i imdbItem) Title() string { return i.result.PrimaryTitle }
func (i imdbItem) Description() string {
	switch i.result.Type {
	case "tvSeries", "tvMiniSeries":
		return "TV"
	case "movie":
		return "Movie"
	case "tvMovie":
		return "TV Movie"
	case "short":
		return "Short"
	case "tvSpecial":
		return "TV Special"
	default:
		return i.result.Type
	}
}
func (i imdbItem) FilterValue() string { return i.result.PrimaryTitle + " " + i.result.OriginalTitle + " " + i.result.Type }

type streamItem struct {
	result apiutils.AlcSearchResult
}

func (i streamItem) Title() string {
	title := i.result.Name
	// Add size to title if available for visibility
	if i.result.BehaviorHints.VideoSize > 0 {
		title += " [" + formatSize(i.result.BehaviorHints.VideoSize) + "]"
	}
	return title
}
func (i streamItem) Description() string {
	return i.result.Description
}
func (i streamItem) FilterValue() string {
	// Search through name, description, and filename
	return i.result.Name + " " + i.result.Description + " " + i.result.BehaviorHints.Filename
}

type seasonItem struct {
	result apiutils.Season
}

func (i seasonItem) Title() string       { return "Season " + i.result.Season }
func (i seasonItem) Description() string { return fmt.Sprintf("%d episodes", i.result.EpisodeCount) }
func (i seasonItem) FilterValue() string { return i.result.Season }

type episodeItem struct {
	result apiutils.Episode
}

func (i episodeItem) Title() string {
	return fmt.Sprintf("E%d: %s", i.result.EpisodeNumber, i.result.Title)
}
func (i episodeItem) Description() string {
	desc := i.result.Plot
	if len(desc) > 80 {
		desc = desc[:80] + "..."
	}
	if i.result.Rating.AggregateRating > 0 {
		desc = fmt.Sprintf("★ %.1f • %s", i.result.Rating.AggregateRating, desc)
	}
	return desc
}
func (i episodeItem) FilterValue() string {
	return fmt.Sprintf("%d %s %s", i.result.EpisodeNumber, i.result.Title, i.result.Plot)
}

// Main model
type Model struct {
	view        View
	searchInput textinput.Model
	resultsList list.Model
	seasonsList list.Model
	episodesList list.Model
	streamsList list.Model
	spinner     spinner.Model
	progress    progress.Model

	// Data
	imdbResults     []apiutils.ImdbSearchResult
	seasons         []apiutils.Season
	episodes        []apiutils.Episode
	streams         []apiutils.AlcSearchResult
	selectedTitle   *apiutils.ImdbSearchResult
	selectedSeason  *apiutils.Season
	selectedEpisode *apiutils.Episode
	selectedStream  *apiutils.AlcSearchResult

	// Store all items for manual filtering
	allEpisodeItems []list.Item
	allStreamItems  []list.Item

	// Custom filter state
	filterInput textinput.Model
	isFiltering bool

	// Download state
	downloading  bool
	downloadProg float64
	downloadFile string
	downloadErr  error

	// Loading states
	loading    bool
	loadingMsg string

	// Status messages
	statusMsg string
	errorMsg  string

	// Dimensions
	width, height int
}

func NewModel() Model {
	// Search input
	ti := textinput.New()
	ti.Placeholder = "Search for movies or shows..."
	ti.Focus()
	ti.Width = 50
	ti.Prompt = "› "
	ti.PromptStyle = SelectedStyle
	ti.TextStyle = NormalStyle
	ti.PlaceholderStyle = DimStyle

	// Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	// Progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Create styled delegate for all lists
	createDelegate := func() list.DefaultDelegate {
		d := list.NewDefaultDelegate()
		d.Styles.SelectedTitle = SelectedStyle.Padding(0, 0, 0, 2)
		d.Styles.SelectedDesc = DimStyle.Padding(0, 0, 0, 2)
		d.Styles.NormalTitle = NormalStyle.Padding(0, 0, 0, 2)
		d.Styles.NormalDesc = DimStyle.Padding(0, 0, 0, 2)
		d.SetSpacing(1)
		return d
	}

	// Results list
	resultsList := list.New([]list.Item{}, createDelegate(), 0, 0)
	resultsList.Title = "Search Results"
	resultsList.SetShowStatusBar(false)
	resultsList.SetFilteringEnabled(false)
	resultsList.Styles.Title = TitleStyle.Padding(0, 0, 1, 2)

	// Seasons list
	seasonsList := list.New([]list.Item{}, createDelegate(), 0, 0)
	seasonsList.Title = "Select Season"
	seasonsList.SetShowStatusBar(false)
	seasonsList.SetFilteringEnabled(false)
	seasonsList.Styles.Title = TitleStyle.Padding(0, 0, 1, 2)

	// Episodes list
	episodesList := list.New([]list.Item{}, createDelegate(), 0, 0)
	episodesList.Title = "Select Episode"
	episodesList.SetShowStatusBar(false)
	episodesList.SetFilteringEnabled(false)
	episodesList.Styles.Title = TitleStyle.Padding(0, 0, 1, 2)

	// Streams list
	streamsList := list.New([]list.Item{}, createDelegate(), 0, 0)
	streamsList.Title = "Available Streams"
	streamsList.SetShowStatusBar(true)
	streamsList.SetFilteringEnabled(false)
	streamsList.Styles.Title = TitleStyle.Padding(0, 0, 1, 2)

	// Filter input for episodes/streams
	fi := textinput.New()
	fi.Placeholder = "Filter..."
	fi.Width = 30
	fi.Prompt = "/ "
	fi.PromptStyle = SelectedStyle
	fi.TextStyle = NormalStyle
	fi.PlaceholderStyle = DimStyle

	return Model{
		view:         SearchView,
		searchInput:  ti,
		filterInput:  fi,
		resultsList:  resultsList,
		seasonsList:  seasonsList,
		episodesList: episodesList,
		streamsList:  streamsList,
		spinner:      sp,
		progress:     prog,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resultsList.SetSize(msg.Width-4, msg.Height-8)
		m.seasonsList.SetSize(msg.Width-4, msg.Height-8)
		m.episodesList.SetSize(msg.Width-4, msg.Height-8)
		m.streamsList.SetSize(msg.Width-4, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.view == SearchView && !m.searchInput.Focused() {
				return m, tea.Quit
			}
			// Don't quit if filtering
			if m.isFiltering {
				break
			}
			if m.view != SearchView {
				return m, tea.Quit
			}
		}

		// View-specific handling
		switch m.view {
		case SearchView:
			return m.updateSearchView(msg)
		case ResultsView:
			return m.updateResultsView(msg)
		case SeasonsView:
			return m.updateSeasonsView(msg)
		case EpisodesView:
			return m.updateEpisodesView(msg)
		case StreamsView:
			return m.updateStreamsView(msg)
		case DownloadView:
			return m.updateDownloadView(msg)
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		if m.downloading {
			progressModel, cmd := m.progress.Update(msg)
			m.progress = progressModel.(progress.Model)
			cmds = append(cmds, cmd)
		}

	// Custom messages
	case searchResultsMsg:
		m.loading = false
		m.imdbResults = msg.results
		if len(msg.results) == 0 {
			m.errorMsg = "No results found"
			return m, nil
		}
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = imdbItem{result: r}
		}
		m.resultsList.SetItems(items)
		m.view = ResultsView
		m.errorMsg = ""
		return m, nil

	case streamsResultsMsg:
		m.loading = false
		m.streams = msg.results
		if len(msg.results) == 0 {
			m.errorMsg = "No streams found"
			// Go back to appropriate view
			if m.selectedEpisode != nil {
				m.view = EpisodesView
			} else {
				m.view = ResultsView
			}
			return m, nil
		}
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = streamItem{result: r}
		}
		m.allStreamItems = items // Store for filtering
		m.streamsList.SetItems(items)
		m.view = StreamsView
		m.isFiltering = false
		m.filterInput.SetValue("")
		m.errorMsg = ""
		return m, nil

	case seasonsResultsMsg:
		m.loading = false
		m.seasons = msg.results
		if len(msg.results) == 0 {
			// No seasons = treat as movie, fetch streams directly
			m.loading = true
			m.loadingMsg = "Fetching streams..."
			return m, tea.Batch(m.spinner.Tick, fetchStreams(m.selectedTitle.Id))
		}
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = seasonItem{result: r}
		}
		m.seasonsList.SetItems(items)
		m.view = SeasonsView
		m.errorMsg = ""
		return m, nil

	case episodesResultsMsg:
		m.loading = false
		m.episodes = msg.results
		if len(msg.results) == 0 {
			m.errorMsg = "No episodes found"
			m.view = SeasonsView
			return m, nil
		}
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = episodeItem{result: r}
		}
		m.allEpisodeItems = items // Store for filtering
		m.episodesList.SetItems(items)
		m.view = EpisodesView
		m.isFiltering = false
		m.filterInput.SetValue("")
		m.errorMsg = ""
		return m, nil

	case downloadProgressMsg:
		m.downloadProg = msg.progress
		return m, m.progress.SetPercent(msg.progress)

	case downloadCompleteMsg:
		m.downloading = false
		m.downloadProg = 1.0
		if msg.err != nil {
			m.downloadErr = msg.err
			m.errorMsg = "Download failed: " + msg.err.Error()
		} else {
			m.statusMsg = "Downloaded: " + msg.filename
		}
		m.view = StreamsView
		return m, nil

	case mpvLaunchedMsg:
		if msg.err != nil {
			m.errorMsg = "Failed to launch mpv: " + msg.err.Error()
		} else {
			m.statusMsg = "Playing in mpv..."
		}
		return m, nil

	case errorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateSearchView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil
		}
		m.loading = true
		m.loadingMsg = "Searching..."
		m.errorMsg = ""
		return m, tea.Batch(m.spinner.Tick, searchIMDB(query))
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) updateResultsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = SearchView
		m.searchInput.Focus()
		return m, nil
	case "enter":
		if item, ok := m.resultsList.SelectedItem().(imdbItem); ok {
			m.selectedTitle = &item.result
			m.selectedSeason = nil
			m.selectedEpisode = nil
			m.loading = true
			m.errorMsg = ""

			// Check if it's a series - fetch seasons first
			if item.result.Type == "tvSeries" || item.result.Type == "tvMiniSeries" {
				m.loadingMsg = "Fetching seasons..."
				return m, tea.Batch(m.spinner.Tick, fetchSeasons(item.result.Id))
			}

			// It's a movie - fetch streams directly
			m.loadingMsg = "Fetching streams..."
			return m, tea.Batch(m.spinner.Tick, fetchStreams(item.result.Id))
		}
	}

	var cmd tea.Cmd
	m.resultsList, cmd = m.resultsList.Update(msg)
	return m, cmd
}

func (m Model) updateSeasonsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = ResultsView
		return m, nil
	case "enter":
		if item, ok := m.seasonsList.SelectedItem().(seasonItem); ok {
			m.selectedSeason = &item.result
			m.loading = true
			m.loadingMsg = "Fetching episodes..."
			m.errorMsg = ""
			return m, tea.Batch(m.spinner.Tick, fetchEpisodes(m.selectedTitle.Id, item.result.Season))
		}
	}

	var cmd tea.Cmd
	m.seasonsList, cmd = m.seasonsList.Update(msg)
	return m, cmd
}

func (m Model) updateEpisodesView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.isFiltering {
		switch msg.String() {
		case "esc":
			// Cancel filtering, restore all items
			m.isFiltering = false
			m.filterInput.Blur()
			m.filterInput.SetValue("")
			m.episodesList.SetItems(m.allEpisodeItems)
			return m, nil
		case "enter":
			// Exit filter mode, keep filtered results
			m.isFiltering = false
			m.filterInput.Blur()
			return m, nil
		default:
			// Update filter input and filter items
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)

			// Apply filter
			filterText := strings.ToLower(m.filterInput.Value())
			if filterText == "" {
				m.episodesList.SetItems(m.allEpisodeItems)
			} else {
				var filtered []list.Item
				for _, item := range m.allEpisodeItems {
					if strings.Contains(strings.ToLower(item.FilterValue()), filterText) {
						filtered = append(filtered, item)
					}
				}
				m.episodesList.SetItems(filtered)
			}
			return m, cmd
		}
	}

	// Normal mode
	switch msg.String() {
	case "/":
		// Enter filter mode
		m.isFiltering = true
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, textinput.Blink
	case "esc":
		// Reset filter state for episodes view
		m.episodesList.SetItems(m.allEpisodeItems)
		m.view = SeasonsView
		return m, nil
	case "enter":
		if item, ok := m.episodesList.SelectedItem().(episodeItem); ok {
			m.selectedEpisode = &item.result
			m.loading = true
			m.loadingMsg = "Fetching streams..."
			m.errorMsg = ""
			// For episodes, we need to use the episode ID with season:episode format
			streamId := fmt.Sprintf("%s:%s:%d", m.selectedTitle.Id, m.selectedSeason.Season, item.result.EpisodeNumber)
			return m, tea.Batch(m.spinner.Tick, fetchStreams(streamId))
		}
	}

	var cmd tea.Cmd
	m.episodesList, cmd = m.episodesList.Update(msg)
	return m, cmd
}

func (m Model) updateStreamsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.isFiltering {
		switch msg.String() {
		case "esc":
			// Cancel filtering, restore all items
			m.isFiltering = false
			m.filterInput.Blur()
			m.filterInput.SetValue("")
			m.streamsList.SetItems(m.allStreamItems)
			return m, nil
		case "enter":
			// Exit filter mode, keep filtered results
			m.isFiltering = false
			m.filterInput.Blur()
			return m, nil
		default:
			// Update filter input and filter items
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)

			// Apply filter
			filterText := strings.ToLower(m.filterInput.Value())
			if filterText == "" {
				m.streamsList.SetItems(m.allStreamItems)
			} else {
				var filtered []list.Item
				for _, item := range m.allStreamItems {
					if strings.Contains(strings.ToLower(item.FilterValue()), filterText) {
						filtered = append(filtered, item)
					}
				}
				m.streamsList.SetItems(filtered)
			}
			return m, cmd
		}
	}

	// Normal mode
	switch msg.String() {
	case "/":
		// Enter filter mode
		m.isFiltering = true
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, textinput.Blink
	case "esc":
		// Go back to episodes if watching a series, otherwise results
		if m.selectedEpisode != nil {
			m.view = EpisodesView
		} else {
			m.view = ResultsView
		}
		// Reset filter state for streams view
		m.streamsList.SetItems(m.allStreamItems)
		m.statusMsg = ""
		m.errorMsg = ""
		return m, nil
	case "p", "enter":
		if item, ok := m.streamsList.SelectedItem().(streamItem); ok {
			m.selectedStream = &item.result
			m.statusMsg = ""
			m.errorMsg = ""
			return m, playStream(item.result.Url)
		}
	case "d":
		if item, ok := m.streamsList.SelectedItem().(streamItem); ok {
			m.selectedStream = &item.result
			m.downloading = true
			m.downloadProg = 0
			// Use filename from API if available, otherwise sanitize the name
			var filename string
			if item.result.BehaviorHints.Filename != "" {
				filename = item.result.BehaviorHints.Filename
			} else {
				filename = sanitizeFilename(item.result.Name)
			}
			m.downloadFile = "downloads/" + filename
			m.view = DownloadView
			m.statusMsg = ""
			m.errorMsg = ""
			return m, downloadStream(item.result.Url, m.downloadFile)
		}
	}

	var cmd tea.Cmd
	m.streamsList, cmd = m.streamsList.Update(msg)
	return m, cmd
}

func (m Model) updateDownloadView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// TODO: Cancel download
		m.downloading = false
		m.view = StreamsView
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case SearchView:
		return m.searchView()
	case ResultsView:
		return m.resultsView()
	case SeasonsView:
		return m.seasonsView()
	case EpisodesView:
		return m.episodesView()
	case StreamsView:
		return m.streamsView()
	case DownloadView:
		return m.downloadView()
	}
	return ""
}

func sanitizeFilename(name string) string {
	// Remove invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Trim spaces and limit length
	result = strings.TrimSpace(result)
	if len(result) > 100 {
		result = result[:100]
	}
	if result == "" {
		result = "download"
	}
	return result + ".mp4"
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
