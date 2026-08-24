package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestParseFlags_ValidTypes(t *testing.T) {
	tests := []struct {
		args        []string
		expectedTyp MovieType
	}{
		{[]string{"--type", "playing"}, TypePlaying},
		{[]string{"--type", "popular"}, TypePopular},
		{[]string{"--type", "top"}, TypeTop},
		{[]string{"--type", "upcoming"}, TypeUpcoming},
		{[]string{"--type=playing"}, TypePlaying},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cfg, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.MovieType != tt.expectedTyp {
				t.Errorf("MovieType = %q; want %q", cfg.MovieType, tt.expectedTyp)
			}
		})
	}
}

func TestParseFlags_AllOptions(t *testing.T) {
	args := []string{
		"--type", "popular",
		"--api-key", "my-test-key",
		"--page", "3",
		"--limit", "10",
		"--json",
		"--detailed",
	}

	cfg, err := parseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.MovieType != TypePopular {
		t.Errorf("MovieType = %q; want %q", cfg.MovieType, TypePopular)
	}
	if cfg.APIKey != "my-test-key" {
		t.Errorf("APIKey = %q; want 'my-test-key'", cfg.APIKey)
	}
	if cfg.Page != 3 {
		t.Errorf("Page = %d; want 3", cfg.Page)
	}
	if cfg.Limit != 10 {
		t.Errorf("Limit = %d; want 10", cfg.Limit)
	}
	if !cfg.JSON {
		t.Errorf("JSON = %v; want true", cfg.JSON)
	}
	if !cfg.Detailed {
		t.Errorf("Detailed = %v; want true", cfg.Detailed)
	}
}

func TestParseFlags_Version(t *testing.T) {
	for _, flagName := range []string{"--version", "-v"} {
		t.Run(flagName, func(t *testing.T) {
			cfg, err := parseFlags([]string{flagName})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cfg.ShowVersion {
				t.Errorf("ShowVersion = %v; want true", cfg.ShowVersion)
			}
		})
	}
}

func TestParseFlags_Errors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{"missing type", []string{}, "missing required flag: --type"},
		{"empty type", []string{"--type", ""}, "missing required flag: --type"},
		{"invalid type", []string{"--type", "invalid"}, "invalid movie type"},
		{"invalid page zero", []string{"--type", "popular", "--page", "0"}, "invalid page number"},
		{"invalid page negative", []string{"--type", "popular", "--page", "-1"}, "invalid page number"},
		{"invalid limit negative", []string{"--type", "popular", "--limit", "-5"}, "invalid limit"},
		{"unknown flag", []string{"--foo"}, "flag provided but not defined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFlags(tt.args)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestParseFlags_APIKeyEnvFallback(t *testing.T) {
	// Set environment variables and test fallback
	t.Run("TMDB_API_KEY", func(t *testing.T) {
		t.Setenv("TMDB_API_KEY", "env-api-key-123")
		cfg, err := parseFlags([]string{"--type", "playing"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "env-api-key-123" {
			t.Errorf("APIKey = %q; want 'env-api-key-123'", cfg.APIKey)
		}
	})

	t.Run("Flag overrides env", func(t *testing.T) {
		t.Setenv("TMDB_API_KEY", "env-api-key-123")
		cfg, err := parseFlags([]string{"--type", "playing", "--api-key", "override-key"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "override-key" {
			t.Errorf("APIKey = %q; want 'override-key'", cfg.APIKey)
		}
	})

	t.Run("TMDB_BEARER_TOKEN fallback", func(t *testing.T) {
		os.Unsetenv("TMDB_API_KEY")
		os.Unsetenv("TMDB_TOKEN")
		t.Setenv("TMDB_BEARER_TOKEN", "bearer-token-xyz")
		cfg, err := parseFlags([]string{"--type", "top"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "bearer-token-xyz" {
			t.Errorf("APIKey = %q; want 'bearer-token-xyz'", cfg.APIKey)
		}
	})
}

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	output := buf.String()

	if !strings.Contains(output, "TMDB CLI Tool") {
		t.Errorf("usage output missing title")
	}
	if !strings.Contains(output, "--type \"playing\"") {
		t.Errorf("usage output missing examples")
	}
}
