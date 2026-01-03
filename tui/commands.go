package tui

import (
	"os"
	"os/exec"
	"time"

	"github.com/cavaliergopher/grab/v3"
	tea "github.com/charmbracelet/bubbletea"

	apiutils "github.com/rshero/stremio-tui/utils"
)

// Package-level program reference for sending progress updates
var programRef *tea.Program

func SetProgramRef(p *tea.Program) {
	programRef = p
}

// Message types
type searchResultsMsg struct {
	results []apiutils.ImdbSearchResult
}

type streamsResultsMsg struct {
	results []apiutils.AlcSearchResult
}

type seasonsResultsMsg struct {
	results []apiutils.Season
}

type episodesResultsMsg struct {
	results []apiutils.Episode
}

type downloadProgressMsg struct {
	id       int
	progress float64
}

type downloadCompleteMsg struct {
	id       int
	filename string
	err      error
}

type downloadStartedMsg struct {
	id int
}

type mpvLaunchedMsg struct {
	err error
}

type errorMsg string

// Commands
func searchIMDB(query string) tea.Cmd {
	return func() tea.Msg {
		results := apiutils.ImdbSearch(query, 20)
		return searchResultsMsg{results: results}
	}
}

func fetchStreams(id string) tea.Cmd {
	return func() tea.Msg {
		results := apiutils.AlcStream(id)
		return streamsResultsMsg{results: results}
	}
}

func fetchSeasons(id string) tea.Cmd {
	return func() tea.Msg {
		results := apiutils.FetchSeasons(id)
		return seasonsResultsMsg{results: results}
	}
}

func fetchEpisodes(id string, season string) tea.Cmd {
	return func() tea.Msg {
		results := apiutils.FetchEpisodes(id, season)
		return episodesResultsMsg{results: results}
	}
}

func playStream(url string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("mpv", url)
		err := cmd.Start()
		return mpvLaunchedMsg{err: err}
	}
}

func downloadStream(url, filename string) tea.Cmd {
	return func() tea.Msg {
		// Ensure downloads directory exists
		if err := os.MkdirAll("downloads", 0755); err != nil {
			return downloadCompleteMsg{filename: filename, err: err}
		}

		client := grab.NewClient()
		req, err := grab.NewRequest(filename, url)
		if err != nil {
			return downloadCompleteMsg{filename: filename, err: err}
		}

		resp := client.Do(req)

		// Monitor progress
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Progress updates are handled by the main loop
				// This is simplified - in a real app you'd use channels
			case <-resp.Done:
				if err := resp.Err(); err != nil {
					return downloadCompleteMsg{filename: filename, err: err}
				}
				return downloadCompleteMsg{filename: resp.Filename, err: nil}
			}
		}
	}
}

// Download with progress reporting via package-level program reference
func downloadStreamWithProgress(id int, url, filename string) tea.Cmd {
	return func() tea.Msg {
		// Ensure downloads directory exists
		if err := os.MkdirAll("downloads", 0755); err != nil {
			return downloadCompleteMsg{id: id, filename: filename, err: err}
		}

		client := grab.NewClient()
		req, err := grab.NewRequest(filename, url)
		if err != nil {
			return downloadCompleteMsg{id: id, filename: filename, err: err}
		}

		resp := client.Do(req)

		// Monitor progress in background
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progress := resp.Progress()
				if programRef != nil {
					programRef.Send(downloadProgressMsg{id: id, progress: progress})
				}
			case <-resp.Done:
				if err := resp.Err(); err != nil {
					return downloadCompleteMsg{id: id, filename: filename, err: err}
				}
				return downloadCompleteMsg{id: id, filename: resp.Filename, err: nil}
			}
		}
	}
}
