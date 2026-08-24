package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.themoviedb.org/3"

// Client is a client for the TMDB API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new TMDB API client.
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: defaultBaseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FetchMovies retrieves movies of the specified type and page from the TMDB API.
func (c *Client) FetchMovies(ctx context.Context, movieType MovieType, page int) (*MovieResponse, error) {
	if c.APIKey == "" {
		return nil, errors.New("TMDB API key is not configured.\nPlease set the TMDB_API_KEY environment variable or provide --api-key.\nGet an API key from: https://www.themoviedb.org/settings/api")
	}

	endpoint := movieType.Endpoint()
	if endpoint == "" {
		return nil, fmt.Errorf("unsupported movie type: %s", movieType)
	}

	reqURL, err := url.Parse(strings.TrimRight(c.BaseURL, "/") + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	q := reqURL.Query()
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}

	// Detect if API key is a JWT/Bearer token (TMDB API Read Access Token) or standard v3 key (32-char hex)
	isBearerToken := strings.Contains(c.APIKey, ".") || len(c.APIKey) > 40
	if !isBearerToken {
		q.Set("api_key", c.APIKey)
	}
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if isBearerToken {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, ctx.Err()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errors.New("request timed out while contacting TMDB API; please check your network connection")
		}
		return nil, fmt.Errorf("network error contacting TMDB API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result MovieResponse
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse TMDB response: %w", err)
		}
		return &result, nil

	case http.StatusUnauthorized:
		var apiErr APIErrorResponse
		if err := json.Unmarshal(bodyBytes, &apiErr); err == nil && apiErr.StatusMessage != "" {
			return nil, fmt.Errorf("unauthorized: %s (please verify your TMDB_API_KEY)", apiErr.StatusMessage)
		}
		return nil, errors.New("unauthorized: invalid or expired TMDB API key (please verify your TMDB_API_KEY)")

	case http.StatusNotFound:
		return nil, fmt.Errorf("TMDB endpoint %s not found (HTTP 404)", endpoint)

	case http.StatusTooManyRequests:
		return nil, errors.New("TMDB API rate limit exceeded; please wait a moment and try again")

	default:
		var apiErr APIErrorResponse
		if err := json.Unmarshal(bodyBytes, &apiErr); err == nil && apiErr.StatusMessage != "" {
			return nil, fmt.Errorf("TMDB API error (HTTP %d): %s", resp.StatusCode, apiErr.StatusMessage)
		}
		return nil, fmt.Errorf("TMDB API returned unexpected HTTP status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
}
