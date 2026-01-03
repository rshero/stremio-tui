package tui

import (
	"fmt"
	"strings"
	"time"

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
	BatchInputView
	BatchSelectView
)

type Tab int

const (
	MainTab Tab = iota
	DownloadsTab
)

type DownloadStatus int

const (
	DownloadPending DownloadStatus = iota
	DownloadInProgress
	DownloadComplete
	DownloadFailed
	DownloadCancelled
)

type Download struct {
	ID         int
	Name       string
	Filename   string
	URL        string
	Progress   float64
	Status     DownloadStatus
	Error      error
	CancelChan chan struct{}
}

// BatchStream represents a stream in batch download selection
type BatchStream struct {
	Episode  apiutils.Episode
	Stream   apiutils.AlcSearchResult
	Selected bool
}

// BatchFailure tracks episodes that failed to fetch streams
type BatchFailure struct {
	Episode apiutils.Episode
	Reason  string
}

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
	currentTab  Tab
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

	// Downloads tracking
	downloads           []Download
	nextDownloadID      int
	selectedDownloadIdx int

	// Batch download state
	batchInput       textinput.Model
	batchStreams     []BatchStream
	batchFailed      []BatchFailure // tracks failed episode fetches
	batchSelectedIdx int
	batchFetching    int // tracks how many episodes are still being fetched

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

	// Batch download input
	bi := textinput.New()
	bi.Placeholder = "e.g. 1080p.BluRay"
	bi.Width = 40
	bi.Prompt = "Release name: "
	bi.PromptStyle = SelectedStyle
	bi.TextStyle = NormalStyle
	bi.PlaceholderStyle = DimStyle

	return Model{
		view:         SearchView,
		currentTab:   MainTab,
		searchInput:  ti,
		filterInput:  fi,
		batchInput:   bi,
		resultsList:  resultsList,
		seasonsList:  seasonsList,
		episodesList: episodesList,
		streamsList:  streamsList,
		spinner:      sp,
		progress:     prog,
		downloads:    []Download{},
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
		case "tab":
			// Cycle between tabs (don't switch if filtering or loading)
			if !m.isFiltering && !m.loading {
				if m.currentTab == MainTab {
					m.currentTab = DownloadsTab
				} else {
					m.currentTab = MainTab
				}
				return m, nil
			}
		case "q":
			if m.currentTab == DownloadsTab {
				m.currentTab = MainTab
				return m, nil
			}
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

		// Handle tab-specific keys
		if m.currentTab == DownloadsTab {
			return m.updateDownloadsTab(msg)
		}

		// Only process view-specific keys on Main tab
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
		case BatchInputView:
			return m.updateBatchInputView(msg)
		case BatchSelectView:
			return m.updateBatchSelectView(msg)
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		// Update progress model if there are active downloads
		if m.activeDownloadCount() > 0 {
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
		// Update progress for specific download
		for i := range m.downloads {
			if m.downloads[i].ID == msg.id {
				m.downloads[i].Progress = msg.progress
				m.downloads[i].Status = DownloadInProgress
				break
			}
		}
		return m, nil

	case downloadCompleteMsg:
		// Update status for specific download
		for i := range m.downloads {
			if m.downloads[i].ID == msg.id {
				if msg.err != nil {
					m.downloads[i].Status = DownloadFailed
					m.downloads[i].Error = msg.err
				} else {
					m.downloads[i].Progress = 1.0
					m.downloads[i].Status = DownloadComplete
					m.downloads[i].Filename = msg.filename
				}
				break
			}
		}
		return m, nil

	case mpvLaunchedMsg:
		if msg.err != nil {
			m.errorMsg = "Failed to launch mpv: " + msg.err.Error()
		} else {
			m.statusMsg = "Playing in mpv..."
		}
		return m, nil

	case batchStreamResultMsg:
		m.batchFetching--

		// Check for fetch error
		if msg.err != nil {
			m.batchFailed = append(m.batchFailed, BatchFailure{
				Episode: msg.episode,
				Reason:  msg.err.Error(),
			})
		} else {
			// Get the release name filter
			releaseName := strings.ToLower(m.batchInput.Value())

			// Find streams matching the release name and add them
			found := false
			for _, stream := range msg.streams {
				if strings.Contains(strings.ToLower(stream.Name), releaseName) ||
					strings.Contains(strings.ToLower(stream.BehaviorHints.Filename), releaseName) {
					m.batchStreams = append(m.batchStreams, BatchStream{
						Episode:  msg.episode,
						Stream:   stream,
						Selected: true, // Pre-select matching streams
					})
					found = true
					break // Only add first matching stream per episode
				}
			}

			// Track as failure if no matching stream found
			if !found && len(msg.streams) > 0 {
				m.batchFailed = append(m.batchFailed, BatchFailure{
					Episode: msg.episode,
					Reason:  "no match for '" + m.batchInput.Value() + "'",
				})
			} else if !found {
				m.batchFailed = append(m.batchFailed, BatchFailure{
					Episode: msg.episode,
					Reason:  "no streams available",
				})
			}
		}

		// When all episodes are fetched, transition to select view
		if m.batchFetching <= 0 {
			m.loading = false
			if len(m.batchStreams) == 0 && len(m.batchFailed) == 0 {
				m.errorMsg = "No streams found matching '" + m.batchInput.Value() + "'"
				m.view = EpisodesView
			} else {
				m.view = BatchSelectView
				m.batchSelectedIdx = 0
			}
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
	case "b":
		// Start batch download - enter release name
		m.view = BatchInputView
		m.batchInput.SetValue("")
		m.batchInput.Focus()
		m.batchStreams = []BatchStream{}
		m.batchSelectedIdx = 0
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

func (m Model) updateBatchInputView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = EpisodesView
		m.batchInput.Blur()
		return m, nil
	case "enter":
		releaseName := strings.TrimSpace(m.batchInput.Value())
		if releaseName == "" {
			m.errorMsg = "Please enter a release name"
			return m, nil
		}
		m.batchInput.Blur()
		m.loading = true
		m.loadingMsg = "Fetching streams for all episodes..."
		m.batchStreams = []BatchStream{}
		m.batchFailed = []BatchFailure{}
		m.batchFetching = len(m.episodes)

		// Start fetching streams for all episodes with staggered delays
		var cmds []tea.Cmd
		cmds = append(cmds, m.spinner.Tick)
		for i, ep := range m.episodes {
			// Stagger requests by 5s each to avoid rate limiting
			delay := time.Duration(i) * 5 * time.Second
			cmds = append(cmds, fetchBatchStreams(m.selectedTitle.Id, m.selectedSeason.Season, ep, delay))
		}
		return m, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	m.batchInput, cmd = m.batchInput.Update(msg)
	return m, cmd
}

func (m Model) updateBatchSelectView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = EpisodesView
		return m, nil
	case "j", "down":
		if len(m.batchStreams) > 0 && m.batchSelectedIdx < len(m.batchStreams)-1 {
			m.batchSelectedIdx++
		}
		return m, nil
	case "k", "up":
		if m.batchSelectedIdx > 0 {
			m.batchSelectedIdx--
		}
		return m, nil
	case " ":
		// Toggle selection
		if len(m.batchStreams) > 0 && m.batchSelectedIdx < len(m.batchStreams) {
			m.batchStreams[m.batchSelectedIdx].Selected = !m.batchStreams[m.batchSelectedIdx].Selected
		}
		return m, nil
	case "a":
		// Select all
		for i := range m.batchStreams {
			m.batchStreams[i].Selected = true
		}
		return m, nil
	case "n":
		// Select none
		for i := range m.batchStreams {
			m.batchStreams[i].Selected = false
		}
		return m, nil
	case "enter":
		// Start downloading all selected
		var cmds []tea.Cmd
		for _, bs := range m.batchStreams {
			if !bs.Selected {
				continue
			}
			// Create filename
			var filename string
			if bs.Stream.BehaviorHints.Filename != "" {
				filename = bs.Stream.BehaviorHints.Filename
			} else {
				filename = sanitizeFilename(fmt.Sprintf("S%sE%02d_%s", m.selectedSeason.Season, bs.Episode.EpisodeNumber, bs.Stream.Name))
			}
			filepath := "downloads/" + filename

			// Add to downloads list
			cancelChan := make(chan struct{})
			download := Download{
				ID:         m.nextDownloadID,
				Name:       fmt.Sprintf("S%sE%02d: %s", m.selectedSeason.Season, bs.Episode.EpisodeNumber, bs.Stream.Name),
				Filename:   filepath,
				URL:        bs.Stream.Url,
				Progress:   0,
				Status:     DownloadPending,
				CancelChan: cancelChan,
			}
			m.downloads = append(m.downloads, download)
			cmds = append(cmds, downloadStreamWithProgress(download.ID, bs.Stream.Url, filepath, cancelChan))
			m.nextDownloadID++
		}

		if len(cmds) > 0 {
			m.statusMsg = fmt.Sprintf("Started %d downloads - press Tab to view", len(cmds))
		} else {
			m.statusMsg = "No streams selected"
		}
		m.view = EpisodesView
		return m, tea.Batch(cmds...)
	}
	return m, nil
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
			// Use filename from API if available, otherwise sanitize the name
			var filename string
			if item.result.BehaviorHints.Filename != "" {
				filename = item.result.BehaviorHints.Filename
			} else {
				filename = sanitizeFilename(item.result.Name)
			}
			filepath := "downloads/" + filename

			// Add to downloads list
			cancelChan := make(chan struct{})
			download := Download{
				ID:         m.nextDownloadID,
				Name:       item.result.Name,
				Filename:   filepath,
				URL:        item.result.Url,
				Progress:   0,
				Status:     DownloadPending,
				CancelChan: cancelChan,
			}
			m.downloads = append(m.downloads, download)
			m.nextDownloadID++

			m.statusMsg = "Download started - press Tab to view progress"
			m.errorMsg = ""

			// Start download in background with progress reporting
			return m, downloadStreamWithProgress(download.ID, item.result.Url, filepath, cancelChan)
		}
	}

	var cmd tea.Cmd
	m.streamsList, cmd = m.streamsList.Update(msg)
	return m, cmd
}

func (m Model) updateDownloadsTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentTab = MainTab
		return m, nil
	case "j", "down":
		if len(m.downloads) > 0 && m.selectedDownloadIdx < len(m.downloads)-1 {
			m.selectedDownloadIdx++
		}
		return m, nil
	case "k", "up":
		if m.selectedDownloadIdx > 0 {
			m.selectedDownloadIdx--
		}
		return m, nil
	case "x":
		// Cancel selected download
		if len(m.downloads) > 0 && m.selectedDownloadIdx < len(m.downloads) {
			d := &m.downloads[m.selectedDownloadIdx]
			if d.Status == DownloadPending || d.Status == DownloadInProgress {
				d.Status = DownloadCancelled
				if d.CancelChan != nil {
					close(d.CancelChan)
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	var content string

	// Show downloads tab or main content
	if m.currentTab == DownloadsTab {
		content = m.downloadsPageView()
	} else {
		switch m.view {
		case SearchView:
			content = m.searchView()
		case ResultsView:
			content = m.resultsView()
		case SeasonsView:
			content = m.seasonsView()
		case EpisodesView:
			content = m.episodesView()
		case StreamsView:
			content = m.streamsView()
		case BatchInputView:
			content = m.batchInputView()
		case BatchSelectView:
			content = m.batchSelectView()
		}
	}

	// Add tab bar at the bottom
	tabBar := m.renderTabBar()

	return content + "\n" + tabBar
}

// Helper to count active downloads
func (m Model) activeDownloadCount() int {
	count := 0
	for _, d := range m.downloads {
		if d.Status == DownloadPending || d.Status == DownloadInProgress {
			count++
		}
	}
	return count
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
