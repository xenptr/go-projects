package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func formatDate(date time.Time) string {
	return date.Format("02 Jan 2006")
}

func inputDate(date time.Time) string {
	return date.Format("2006-01-02")
}

func parsePublishedDate(r *http.Request) (time.Time, error) {
	return time.Parse("2006-01-02", r.FormValue("published_at"))
}

func validateArticle(title, content string) error {
	switch {
	case title == "":
		return fmt.Errorf("title is required")
	case content == "":
		return fmt.Errorf("content is required")
	case len(title) > 200:
		return fmt.Errorf("title is too long")
	}
	return nil
}

// splitTags parses a comma-separated tag string into a deduplicated, trimmed slice.
func splitTags(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := map[string]struct{}{}
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		tags = append(tags, t)
	}
	return tags
}

// joinTags converts a tag slice to a comma-separated string for form pre-fill.
func joinTags(tags []string) string {
	return strings.Join(tags, ", ")
}

// commentIDFromPath parses the {cid} path value.
func commentIDFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("cid"))
}

// uniqueCategories returns a sorted, deduplicated list of categories from a slice of articles.
func uniqueCategories(articles []Article) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, a := range articles {
		if a.Category == "" {
			continue
		}
		if _, dup := seen[a.Category]; dup {
			continue
		}
		seen[a.Category] = struct{}{}
		result = append(result, a.Category)
	}
	sort.Strings(result)
	return result
}

// uniqueTags returns a sorted, deduplicated list of tags from a slice of articles.
func uniqueTags(articles []Article) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, a := range articles {
		for _, t := range a.Tags {
			if t == "" {
				continue
			}
			if _, dup := seen[t]; dup {
				continue
			}
			seen[t] = struct{}{}
			result = append(result, t)
		}
	}
	sort.Strings(result)
	return result
}

// contentPreview truncates s to at most maxChars characters, breaking on a
// word boundary and appending "…" when the text is longer.
func contentPreview(s string, maxChars int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxChars {
		return s
	}
	// Walk back from maxChars to the nearest space so we don't cut mid-word.
	cut := maxChars
	for cut > 0 && s[cut] != ' ' {
		cut--
	}
	if cut == 0 {
		cut = maxChars // no space found, hard-cut
	}
	return s[:cut] + "…"
}
