package main

import (
	"testing"
)

func TestParseMovieType(t *testing.T) {
	tests := []struct {
		input    string
		expected MovieType
		valid    bool
	}{
		{"playing", TypePlaying, true},
		{"now_playing", TypePlaying, true},
		{"now-playing", TypePlaying, true},
		{"nowplaying", TypePlaying, true},
		{"PLAYING", TypePlaying, true},
		{"popular", TypePopular, true},
		{"Popular", TypePopular, true},
		{"POPULAR", TypePopular, true},
		{"top", TypeTop, true},
		{"top_rated", TypeTop, true},
		{"top-rated", TypeTop, true},
		{"toprated", TypeTop, true},
		{"upcoming", TypeUpcoming, true},
		{"UPCOMING", TypeUpcoming, true},
		{"invalid", "", false},
		{"", "", false},
		{"trending", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual, ok := ParseMovieType(tt.input)
			if ok != tt.valid {
				t.Fatalf("ParseMovieType(%q) ok = %v; want %v", tt.input, ok, tt.valid)
			}
			if actual != tt.expected {
				t.Errorf("ParseMovieType(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestMovieTypeEndpoint(t *testing.T) {
	tests := []struct {
		movieType MovieType
		expected  string
	}{
		{TypePlaying, "/movie/now_playing"},
		{TypePopular, "/movie/popular"},
		{TypeTop, "/movie/top_rated"},
		{TypeUpcoming, "/movie/upcoming"},
		{MovieType("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.movieType), func(t *testing.T) {
			if actual := tt.movieType.Endpoint(); actual != tt.expected {
				t.Errorf("Endpoint() = %q; want %q", actual, tt.expected)
			}
		})
	}
}

func TestMovieTypeTitle(t *testing.T) {
	tests := []struct {
		movieType MovieType
		expected  string
	}{
		{TypePlaying, "Now Playing Movies"},
		{TypePopular, "Popular Movies"},
		{TypeTop, "Top Rated Movies"},
		{TypeUpcoming, "Upcoming Movies"},
		{MovieType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.movieType), func(t *testing.T) {
			if actual := tt.movieType.Title(); actual != tt.expected {
				t.Errorf("Title() = %q; want %q", actual, tt.expected)
			}
		})
	}
}

func TestGenreName(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{28, "Action"},
		{12, "Adventure"},
		{16, "Animation"},
		{35, "Comedy"},
		{80, "Crime"},
		{99, "Documentary"},
		{18, "Drama"},
		{10751, "Family"},
		{14, "Fantasy"},
		{36, "History"},
		{27, "Horror"},
		{10402, "Music"},
		{9648, "Mystery"},
		{10749, "Romance"},
		{878, "Sci-Fi"},
		{10770, "TV Movie"},
		{53, "Thriller"},
		{10752, "War"},
		{37, "Western"},
		{99999, "Genre 99999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if actual := GenreName(tt.id); actual != tt.expected {
				t.Errorf("GenreName(%d) = %q; want %q", tt.id, actual, tt.expected)
			}
		})
	}
}

func TestGenreNames(t *testing.T) {
	if actual := GenreNames([]int{28, 878, 12}); actual != "Action, Sci-Fi, Adventure" {
		t.Errorf("GenreNames() = %q; want 'Action, Sci-Fi, Adventure'", actual)
	}
	if actual := GenreNames(nil); actual != "N/A" {
		t.Errorf("GenreNames(nil) = %q; want 'N/A'", actual)
	}
	if actual := GenreNames([]int{}); actual != "N/A" {
		t.Errorf("GenreNames([]) = %q; want 'N/A'", actual)
	}
}
