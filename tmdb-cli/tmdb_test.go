package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sampleMovieResponse() MovieResponse {
	return MovieResponse{
		Page:         1,
		TotalPages:   10,
		TotalResults: 200,
		Results: []Movie{
			{
				ID:               101,
				Title:            "Inception",
				OriginalTitle:    "Inception",
				Overview:         "A thief who steals corporate secrets through the use of dream-sharing technology.",
				ReleaseDate:      "2010-07-16",
				VoteAverage:      8.8,
				VoteCount:        35000,
				Popularity:       120.5,
				GenreIDs:         []int{28, 878, 12},
				Adult:            false,
				OriginalLanguage: "en",
			},
			{
				ID:               102,
				Title:            "The Dark Knight",
				OriginalTitle:    "The Dark Knight",
				Overview:         "Batman raises the stakes in his war on crime.",
				ReleaseDate:      "2008-07-18",
				VoteAverage:      9.0,
				VoteCount:        40000,
				Popularity:       150.0,
				GenreIDs:         []int{28, 80, 18},
				Adult:            false,
				OriginalLanguage: "en",
			},
		},
		Dates: &DateRange{
			Minimum: "2010-07-01",
			Maximum: "2010-07-31",
		},
	}
}

func TestFetchMovies_Success(t *testing.T) {
	expectedSample := sampleMovieResponse()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected query and headers
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept application/json header, got %s", r.Header.Get("Accept"))
		}

		switch r.URL.Path {
		case "/movie/now_playing", "/movie/popular", "/movie/top_rated", "/movie/upcoming":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedSample)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient("valid-api-key-1234567890abcdef")
	client.BaseURL = server.URL

	types := []MovieType{TypePlaying, TypePopular, TypeTop, TypeUpcoming}
	for _, mType := range types {
		t.Run(string(mType), func(t *testing.T) {
			resp, err := client.FetchMovies(context.Background(), mType, 1)
			if err != nil {
				t.Fatalf("FetchMovies(%s) returned unexpected error: %v", mType, err)
			}
			if resp.Page != 1 {
				t.Errorf("Page = %d; want 1", resp.Page)
			}
			if len(resp.Results) != 2 {
				t.Fatalf("Results count = %d; want 2", len(resp.Results))
			}
			if resp.Results[0].Title != "Inception" {
				t.Errorf("Results[0].Title = %q; want 'Inception'", resp.Results[0].Title)
			}
		})
	}
}

func TestFetchMovies_BearerAuth(t *testing.T) {
	jwtToken := "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJleGFtcGxlIn0.signature"
	receivedBearer := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer "+jwtToken {
			receivedBearer = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleMovieResponse())
	}))
	defer server.Close()

	client := NewClient(jwtToken)
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !receivedBearer {
		t.Errorf("expected Authorization Bearer header, but it was not received")
	}
}

func TestFetchMovies_MissingAPIKey(t *testing.T) {
	client := NewClient("")
	_, err := client.FetchMovies(context.Background(), TypePlaying, 1)
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
	if !strings.Contains(err.Error(), "TMDB API key is not configured") {
		t.Errorf("expected missing key message, got %q", err.Error())
	}
}

func TestFetchMovies_401Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(APIErrorResponse{
			StatusCode:    7,
			StatusMessage: "Invalid API key: You must be granted a valid key.",
			Success:       false,
		})
	}))
	defer server.Close()

	client := NewClient("invalid-key")
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err == nil {
		t.Fatal("expected error for 401 unauthorized, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected unauthorized message, got %q", err.Error())
	}
}

func TestFetchMovies_404NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("some-key")
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got %q", err.Error())
	}
}

func TestFetchMovies_429RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("some-key")
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected rate limit error, got %q", err.Error())
	}
}

func TestFetchMovies_500InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient("some-key")
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 status in error, got %q", err.Error())
	}
}

func TestFetchMovies_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient("some-key")
	client.BaseURL = server.URL

	_, err := client.FetchMovies(context.Background(), TypePopular, 1)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse TMDB response") {
		t.Errorf("expected parse error, got %q", err.Error())
	}
}

func TestFetchMovies_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("some-key")
	client.BaseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.FetchMovies(ctx, TypePopular, 1)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
