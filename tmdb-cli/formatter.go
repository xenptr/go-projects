package main

import (
	"fmt"
	"strconv"
	"strings"
)

const maxOverviewWidth = 72

// formatNumber formats an integer with commas for readability (e.g. 1,234,567).
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		b.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// wrapText wraps a text string to a specific width, prefixing wrapped lines with indent.
func wrapText(text string, width int, indent string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "No overview available."
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return "No overview available."
	}

	var b strings.Builder
	currentLineLen := 0

	for _, word := range words {
		wordLen := len(word)
		if currentLineLen == 0 {
			b.WriteString(word)
			currentLineLen = wordLen
		} else if currentLineLen+1+wordLen <= width {
			b.WriteByte(' ')
			b.WriteString(word)
			currentLineLen += 1 + wordLen
		} else {
			b.WriteByte('\n')
			b.WriteString(indent)
			b.WriteString(word)
			currentLineLen = wordLen
		}
	}

	return b.String()
}

// formatMovie formats a single movie item for terminal display.
func formatMovie(m Movie, index int, detailed bool) string {
	var b strings.Builder

	year := ""
	if len(m.ReleaseDate) >= 4 {
		year = " (" + m.ReleaseDate[:4] + ")"
	}

	fmt.Fprintf(&b, "%2d. %s%s\n", index, m.Title, year)

	if detailed && m.OriginalTitle != "" && m.OriginalTitle != m.Title {
		fmt.Fprintf(&b, "    Original Title: %s [%s]\n", m.OriginalTitle, m.OriginalLanguage)
	}

	fmt.Fprintf(&b, "    ⭐ Rating: %.1f/10 (%s votes)  |  🔥 Popularity: %.1f\n",
		m.VoteAverage, formatNumber(m.VoteCount), m.Popularity)

	genres := GenreNames(m.GenreIDs)
	fmt.Fprintf(&b, "    🎭 Genres: %s\n", genres)

	relDate := m.ReleaseDate
	if relDate == "" {
		relDate = "Unknown"
	}

	if detailed {
		adultStr := "No"
		if m.Adult {
			adultStr = "Yes (18+)"
		}
		fmt.Fprintf(&b, "    📅 Release: %s  |  Adult: %s  |  TMDB ID: %d\n", relDate, adultStr, m.ID)
	} else {
		fmt.Fprintf(&b, "    📅 Release: %s\n", relDate)
	}

	indent := "       "
	wrappedOverview := wrapText(m.Overview, maxOverviewWidth, indent)
	fmt.Fprintf(&b, "    📝 Overview: %s\n", wrappedOverview)

	return b.String()
}

// formatResponse formats the complete MovieResponse into a readable terminal output.
func formatResponse(resp *MovieResponse, movieType MovieType, detailed bool, limit int) string {
	if resp == nil {
		return "No response received.\n"
	}

	var b strings.Builder
	line := strings.Repeat("─", 78)

	b.WriteString(line + "\n")
	fmt.Fprintf(&b, "🎬 %s  (Page %s of %s · %s total movies)\n",
		movieType.Title(),
		formatNumber(resp.Page),
		formatNumber(resp.TotalPages),
		formatNumber(resp.TotalResults),
	)

	if resp.Dates != nil && resp.Dates.Minimum != "" && resp.Dates.Maximum != "" {
		fmt.Fprintf(&b, "   Release Window: %s to %s\n", resp.Dates.Minimum, resp.Dates.Maximum)
	}
	b.WriteString(line + "\n\n")

	movies := resp.Results
	if limit > 0 && limit < len(movies) {
		movies = movies[:limit]
	}

	if len(movies) == 0 {
		b.WriteString("No movies found for this category.\n")
		return b.String()
	}

	for i, movie := range movies {
		b.WriteString(formatMovie(movie, i+1, detailed))
		if i < len(movies)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + line + "\n")
	if limit > 0 && limit < len(resp.Results) {
		fmt.Fprintf(&b, "Showing %d of %d movies on this page (Page %d/%d)\n",
			len(movies), len(resp.Results), resp.Page, resp.TotalPages)
	} else {
		fmt.Fprintf(&b, "Showing %d movies (Page %d of %d)\n",
			len(movies), resp.Page, resp.TotalPages)
	}

	return b.String()
}
