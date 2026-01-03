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

func (m Model) downloadView() string {
	var b strings.Builder

	title := TitleStyle.Render("Downloading...")
	b.WriteString(title + "\n\n")

	if m.downloadFile != "" {
		b.WriteString(DimStyle.Render("File: "+m.downloadFile) + "\n\n")
	}

	// Progress bar
	prog := m.progress.View()
	b.WriteString(prog + "\n\n")

	percent := fmt.Sprintf("%.1f%%", m.downloadProg*100)
	b.WriteString(StatusStyle.Render(percent) + "\n\n")

	if m.downloadErr != nil {
		b.WriteString(ErrorStyle.Render("Error: "+m.downloadErr.Error()) + "\n")
	}

	help := HelpStyle.Render("esc: cancel")
	b.WriteString(help)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		b.String(),
	)
}
