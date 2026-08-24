package main

import (
	"strings"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{99, "99"},
		{100, "100"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{-1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if actual := formatNumber(tt.input); actual != tt.expected {
				t.Errorf("formatNumber(%d) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	text := "This is a short text that should be wrapped across multiple lines if it exceeds the maximum width specified."
	wrapped := wrapText(text, 30, "   ")

	lines := strings.Split(wrapped, "\n")
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}

	for i, l := range lines {
		if i > 0 && !strings.HasPrefix(l, "   ") {
			t.Errorf("line %d does not have indent: %q", i, l)
		}
	}

	// Empty text
	if wrapText("", 30, " ") != "No overview available." {
		t.Errorf("expected default overview for empty string")
	}
}

func TestFormatMovie(t *testing.T) {
	movie := Movie{
		ID:               550,
		Title:            "Fight Club",
		OriginalTitle:    "Fight Club",
		Overview:         "An insomniac office worker and a devil-may-care soap maker form an underground fight club.",
		ReleaseDate:      "1999-10-15",
		VoteAverage:      8.4,
		VoteCount:        27000,
		Popularity:       85.4,
		GenreIDs:         []int{18, 53},
		Adult:            false,
		OriginalLanguage: "en",
	}

	formatted := formatMovie(movie, 1, false)
	if !strings.Contains(formatted, "1. Fight Club (1999)") {
		t.Errorf("missing title and year in output: %s", formatted)
	}
	if !strings.Contains(formatted, "8.4/10") {
		t.Errorf("missing rating in output: %s", formatted)
	}
	if !strings.Contains(formatted, "Drama, Thriller") {
		t.Errorf("missing genres in output: %s", formatted)
	}

	// Detailed view
	detailedFormatted := formatMovie(movie, 1, true)
	if !strings.Contains(detailedFormatted, "TMDB ID: 550") {
		t.Errorf("missing TMDB ID in detailed output: %s", detailedFormatted)
	}
}

func TestFormatResponse(t *testing.T) {
	resp := sampleMovieResponse()

	output := formatResponse(&resp, TypePlaying, false, 0)
	if !strings.Contains(output, "Now Playing Movies") {
		t.Errorf("missing section title")
	}
	if !strings.Contains(output, "Inception (2010)") {
		t.Errorf("missing movie 1")
	}
	if !strings.Contains(output, "The Dark Knight (2008)") {
		t.Errorf("missing movie 2")
	}
	if !strings.Contains(output, "Release Window: 2010-07-01 to 2010-07-31") {
		t.Errorf("missing release window")
	}

	// Test with limit
	limitedOutput := formatResponse(&resp, TypePopular, false, 1)
	if !strings.Contains(limitedOutput, "Inception") {
		t.Errorf("missing movie 1 with limit")
	}
	if strings.Contains(limitedOutput, "The Dark Knight") {
		t.Errorf("movie 2 should not be in limited output")
	}

	// Nil response
	if !strings.Contains(formatResponse(nil, TypeTop, false, 0), "No response") {
		t.Errorf("expected no response text for nil")
	}

	// Empty results
	emptyResp := MovieResponse{Page: 1, TotalPages: 0, TotalResults: 0, Results: []Movie{}}
	if !strings.Contains(formatResponse(&emptyResp, TypeUpcoming, false, 0), "No movies found") {
		t.Errorf("expected no movies found message")
	}
}
