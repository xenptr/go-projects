package main

import (
	"fmt"
	"strings"
)

// MovieType represents the category of movies to fetch.
type MovieType string

const (
	TypePlaying  MovieType = "playing"
	TypePopular  MovieType = "popular"
	TypeTop      MovieType = "top"
	TypeUpcoming MovieType = "upcoming"
)

// ParseMovieType parses and normalizes a string into a MovieType.
func ParseMovieType(s string) (MovieType, bool) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	normalized = strings.ReplaceAll(normalized, "_", "-")

	switch normalized {
	case "playing", "now-playing", "nowplaying":
		return TypePlaying, true
	case "popular":
		return TypePopular, true
	case "top", "top-rated", "toprated":
		return TypeTop, true
	case "upcoming":
		return TypeUpcoming, true
	default:
		return "", false
	}
}

// Endpoint returns the TMDB API path for the movie type.
func (t MovieType) Endpoint() string {
	switch t {
	case TypePlaying:
		return "/movie/now_playing"
	case TypePopular:
		return "/movie/popular"
	case TypeTop:
		return "/movie/top_rated"
	case TypeUpcoming:
		return "/movie/upcoming"
	default:
		return ""
	}
}

// Title returns a human-readable title for the movie type.
func (t MovieType) Title() string {
	switch t {
	case TypePlaying:
		return "Now Playing Movies"
	case TypePopular:
		return "Popular Movies"
	case TypeTop:
		return "Top Rated Movies"
	case TypeUpcoming:
		return "Upcoming Movies"
	default:
		return string(t)
	}
}

// Movie represents a single movie item returned from TMDB API.
type Movie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	GenreIDs         []int   `json:"genre_ids"`
	Adult            bool    `json:"adult"`
	OriginalLanguage string  `json:"original_language"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
}

// DateRange represents the release date range for now_playing and upcoming endpoints.
type DateRange struct {
	Maximum string `json:"maximum"`
	Minimum string `json:"minimum"`
}

// MovieResponse represents the paginated response from TMDB movie list endpoints.
type MovieResponse struct {
	Page         int        `json:"page"`
	Results      []Movie    `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
	Dates        *DateRange `json:"dates,omitempty"`
}

// APIErrorResponse represents an error response from TMDB API.
type APIErrorResponse struct {
	StatusCode    int    `json:"status_code"`
	StatusMessage string `json:"status_message"`
	Success       bool   `json:"success"`
}

// GenreName maps standard TMDB genre IDs to their names.
func GenreName(id int) string {
	switch id {
	case 28:
		return "Action"
	case 12:
		return "Adventure"
	case 16:
		return "Animation"
	case 35:
		return "Comedy"
	case 80:
		return "Crime"
	case 99:
		return "Documentary"
	case 18:
		return "Drama"
	case 10751:
		return "Family"
	case 14:
		return "Fantasy"
	case 36:
		return "History"
	case 27:
		return "Horror"
	case 10402:
		return "Music"
	case 9648:
		return "Mystery"
	case 10749:
		return "Romance"
	case 878:
		return "Sci-Fi"
	case 10770:
		return "TV Movie"
	case 53:
		return "Thriller"
	case 10752:
		return "War"
	case 37:
		return "Western"
	default:
		return fmt.Sprintf("Genre %d", id)
	}
}

// GenreNames converts a slice of genre IDs to comma-separated names.
func GenreNames(ids []int) string {
	if len(ids) == 0 {
		return "N/A"
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = GenreName(id)
	}
	return strings.Join(names, ", ")
}
