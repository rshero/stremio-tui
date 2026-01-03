package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) searchView() string {
	var b strings.Builder

	title := TitleStyle.Render("Stremio TUI")
	b.WriteString(title + "\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " " + m.loadingMsg + "\n\n")
	} else {
		b.WriteString(SubtitleStyle.Render("Search for movies and TV shows") + "\n\n")
	}

	b.WriteString(InputStyle.Render(m.searchInput.View()) + "\n\n")

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n\n")
	}

	help := HelpStyle.Render("enter: search • ctrl+c: quit")
	b.WriteString(help)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		b.String(),
	)
}

func (m Model) resultsView() string {
	var b strings.Builder

	if m.loading {
		loading := m.spinner.View() + " " + m.loadingMsg
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loading,
		)
	}

	b.WriteString(m.resultsList.View())
	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n")
	}

	help := HelpStyle.Render("enter: select • esc: back • j/k: navigate • q: quit")
	b.WriteString(help)

	return b.String()
}

func (m Model) seasonsView() string {
	var b strings.Builder

	if m.loading {
		loading := m.spinner.View() + " " + m.loadingMsg
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loading,
		)
	}

	// Show selected title
	if m.selectedTitle != nil {
		titleInfo := DimStyle.Render(m.selectedTitle.PrimaryTitle)
		b.WriteString(titleInfo + "\n")
	}

	b.WriteString(m.seasonsList.View())
	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n")
	}

	help := HelpStyle.Render("enter: select • esc: back • j/k: navigate • q: quit")
	b.WriteString(help)

	return b.String()
}

func (m Model) episodesView() string {
	var b strings.Builder

	if m.loading {
		loading := m.spinner.View() + " " + m.loadingMsg
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loading,
		)
	}

	// Show selected title and season
	if m.selectedTitle != nil && m.selectedSeason != nil {
		titleInfo := DimStyle.Render(fmt.Sprintf("%s - Season %s", m.selectedTitle.PrimaryTitle, m.selectedSeason.Season))
		b.WriteString(titleInfo + "\n")
	}

	// Show filter input when filtering
	if m.isFiltering {
		b.WriteString(InputStyle.Render(m.filterInput.View()) + "\n")
	}

	b.WriteString(m.episodesList.View())
	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n")
	}

	var help string
	if m.isFiltering {
		help = HelpStyle.Render("enter: apply filter • esc: cancel filter")
	} else {
		help = HelpStyle.Render("enter: select • /: filter • esc: back • j/k: navigate • q: quit")
	}
	b.WriteString(help)

	return b.String()
}

func (m Model) streamsView() string {
	var b strings.Builder

	if m.loading {
		loading := m.spinner.View() + " " + m.loadingMsg
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loading,
		)
	}

	// Show selected title (and episode if applicable)
	if m.selectedTitle != nil {
		var titleInfo string
		if m.selectedEpisode != nil {
			titleInfo = fmt.Sprintf("%s - S%sE%d: %s",
				m.selectedTitle.PrimaryTitle,
				m.selectedSeason.Season,
				m.selectedEpisode.EpisodeNumber,
				m.selectedEpisode.Title)
		} else {
			titleInfo = m.selectedTitle.PrimaryTitle
		}
		b.WriteString(DimStyle.Render(titleInfo) + "\n")
	}

	// Show filter input when filtering
	if m.isFiltering {
		b.WriteString(InputStyle.Render(m.filterInput.View()) + "\n")
	}

	b.WriteString(m.streamsList.View())
	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString(SuccessStyle.Render(m.statusMsg) + "\n")
	}

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n")
	}

	var help string
	if m.isFiltering {
		help = HelpStyle.Render("enter: apply filter • esc: cancel filter")
	} else {
		help = HelpStyle.Render("p/enter: play • d: download • /: filter • esc: back • q: quit")
	}
	b.WriteString(help)

	return b.String()
}

func (m Model) renderTabBar() string {
	var mainTab, downloadsTab string

	// Count active downloads for badge
	activeCount := m.activeDownloadCount()
	downloadsLabel := "Downloads"
	if activeCount > 0 {
		downloadsLabel = fmt.Sprintf("Downloads (%d)", activeCount)
	} else if len(m.downloads) > 0 {
		downloadsLabel = fmt.Sprintf("Downloads (%d)", len(m.downloads))
	}

	if m.currentTab == MainTab {
		mainTab = TabActiveStyle.Render("Main")
		downloadsTab = TabInactiveStyle.Render(downloadsLabel)
	} else {
		mainTab = TabInactiveStyle.Render("Main")
		downloadsTab = TabActiveStyle.Render(downloadsLabel)
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, mainTab, " ", downloadsTab)
	help := HelpStyle.Render("  tab: switch • q: back/quit • ctrl+c: quit")

	return lipgloss.JoinHorizontal(lipgloss.Top, tabBar, help)
}

func (m Model) downloadsPageView() string {
	var b strings.Builder

	title := TitleStyle.Render("Downloads")
	b.WriteString(title + "\n\n")

	if len(m.downloads) == 0 {
		b.WriteString(DimStyle.Render("No downloads yet. Press 'd' on a stream to start downloading.") + "\n")
	} else {
		for _, d := range m.downloads {
			b.WriteString(m.renderDownloadItem(d))
		}
	}

	return b.String()
}

func (m Model) renderDownloadItem(d Download) string {
	var style lipgloss.Style
	var statusIcon, statusText string

	switch d.Status {
	case DownloadPending:
		style = DownloadItemStyle
		statusIcon = "○"
		statusText = "Pending"
	case DownloadInProgress:
		style = DownloadActiveStyle
		statusIcon = "●"
		statusText = fmt.Sprintf("%.1f%%", d.Progress*100)
	case DownloadComplete:
		style = DownloadCompleteStyle
		statusIcon = "✓"
		statusText = "Complete"
	case DownloadFailed:
		style = DownloadFailedStyle
		statusIcon = "✗"
		if d.Error != nil {
			statusText = "Failed: " + d.Error.Error()
		} else {
			statusText = "Failed"
		}
	}

	// Truncate name if too long
	name := d.Name
	maxNameLen := m.width - 30
	if maxNameLen < 20 {
		maxNameLen = 20
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	// Build progress bar for active downloads
	var progressBar string
	if d.Status == DownloadInProgress {
		barWidth := 30
		filled := int(d.Progress * float64(barWidth))
		empty := barWidth - filled
		progressBar = "\n  " + SelectedStyle.Render(strings.Repeat("█", filled)) + DimStyle.Render(strings.Repeat("░", empty))
	}

	content := fmt.Sprintf("%s %s\n  %s%s", statusIcon, name, DimStyle.Render(statusText), progressBar)

	return style.Width(m.width - 4).Render(content) + "\n"
}
