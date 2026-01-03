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
		help = HelpStyle.Render("enter: select • b: batch download • /: filter • esc: back")
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

	return lipgloss.JoinHorizontal(lipgloss.Top, mainTab, " ", downloadsTab, HelpStyle.Render("  tab: switch"))
}

func (m Model) downloadsPageView() string {
	var b strings.Builder

	title := TitleStyle.Render("Downloads")
	b.WriteString(title + "\n\n")

	if len(m.downloads) == 0 {
		b.WriteString(DimStyle.Render("No downloads yet. Press 'd' on a stream to start downloading.") + "\n")
	} else {
		for i, d := range m.downloads {
			isSelected := i == m.selectedDownloadIdx
			b.WriteString(m.renderDownloadItem(d, isSelected))
		}
	}

	b.WriteString("\n")
	help := HelpStyle.Render("j/k: navigate • x: cancel download • esc/q: back to main")
	b.WriteString(help)

	// Use consistent height with other views
	content := b.String()
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 4). // Reserve space for tab bar
		Render(content)
}

func (m Model) batchInputView() string {
	var b strings.Builder

	if m.loading {
		loading := m.spinner.View() + " " + m.loadingMsg
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			loading,
		)
	}

	title := TitleStyle.Render("Batch Download")
	b.WriteString(title + "\n\n")

	// Show selected title and season
	if m.selectedTitle != nil && m.selectedSeason != nil {
		titleInfo := DimStyle.Render(fmt.Sprintf("%s - Season %s (%d episodes)", m.selectedTitle.PrimaryTitle, m.selectedSeason.Season, len(m.episodes)))
		b.WriteString(titleInfo + "\n\n")
	}

	b.WriteString(SubtitleStyle.Render("Enter release name to filter streams:") + "\n\n")
	b.WriteString(InputStyle.Render(m.batchInput.View()) + "\n\n")

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n\n")
	}

	help := HelpStyle.Render("enter: search streams • esc: cancel")
	b.WriteString(help)

	return lipgloss.Place(
		m.width, m.height-4,
		lipgloss.Center, lipgloss.Center,
		b.String(),
	)
}

func (m Model) batchSelectView() string {
	var b strings.Builder

	title := TitleStyle.Render("Select Streams to Download")
	b.WriteString(title + "\n\n")

	// Show selected title and season
	if m.selectedTitle != nil && m.selectedSeason != nil {
		titleInfo := DimStyle.Render(fmt.Sprintf("%s - Season %s", m.selectedTitle.PrimaryTitle, m.selectedSeason.Season))
		b.WriteString(titleInfo + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("Filter: %s", m.batchInput.Value())) + "\n\n")
	}

	// Count selected
	selectedCount := 0
	for _, bs := range m.batchStreams {
		if bs.Selected {
			selectedCount++
		}
	}
	b.WriteString(StatusStyle.Render(fmt.Sprintf("%d of %d selected", selectedCount, len(m.batchStreams))) + "\n\n")

	// Show streams with checkboxes
	for i, bs := range m.batchStreams {
		isSelected := i == m.batchSelectedIdx
		b.WriteString(m.renderBatchStreamItem(bs, isSelected))
	}

	// Show failed episodes if any
	if len(m.batchFailed) > 0 {
		b.WriteString("\n" + ErrorStyle.Render(fmt.Sprintf("Failed to fetch %d episode(s):", len(m.batchFailed))) + "\n")
		for _, failure := range m.batchFailed {
			failInfo := fmt.Sprintf("  E%02d: %s", failure.Episode.EpisodeNumber, failure.Reason)
			b.WriteString(DimStyle.Render(failInfo) + "\n")
		}
	}

	if m.errorMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errorMsg) + "\n")
	}

	help := HelpStyle.Render("space: toggle • a: all • n: none • enter: start downloads • esc: cancel")
	b.WriteString("\n" + help)

	// Use consistent height
	content := b.String()
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 4).
		Render(content)
}

func (m Model) renderBatchStreamItem(bs BatchStream, isSelected bool) string {
	checkbox := "[ ]"
	if bs.Selected {
		checkbox = "[✓]"
	}

	selector := "  "
	nameStyle := NormalStyle
	if isSelected {
		selector = "› "
		nameStyle = SelectedStyle
	}

	// Truncate name if too long
	name := bs.Stream.Name
	maxNameLen := m.width - 40
	if maxNameLen < 20 {
		maxNameLen = 20
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	episodeInfo := fmt.Sprintf("E%02d", bs.Episode.EpisodeNumber)

	return fmt.Sprintf("%s%s %s %s\n", selector, checkbox, DimStyle.Render(episodeInfo), nameStyle.Render(name))
}

func (m Model) renderDownloadItem(d Download, isSelected bool) string {
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
	case DownloadCancelled:
		style = DownloadFailedStyle
		statusIcon = "⊘"
		statusText = "Cancelled"
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

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "› "
		style = style.BorderForeground(lipgloss.Color("#7C3AED")) // primary color for selected
	}

	content := fmt.Sprintf("%s%s %s\n   %s%s", selector, statusIcon, name, DimStyle.Render(statusText), progressBar)

	return style.Width(m.width - 4).Render(content) + "\n"
}
