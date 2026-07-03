// Package yc loads Y Combinator company data from the public yc-oss/api
// dataset (https://github.com/yc-oss/api) via raw.githubusercontent.com.
// We only pull RECENT batches, so every match counts as "newly funded".
package yc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RecentBatches are the YC cohorts treated as "newly funded". Add new slugs
// here as fresh batches land (file names live under yc-oss/api/batches/).
var RecentBatches = []string{
	"summer-2025", "spring-2025", "winter-2025",
	"fall-2024", "summer-2024", "winter-2024",
}

const rawBase = "https://raw.githubusercontent.com/yc-oss/api/main/batches/"

// Company is the subset of yc-oss fields we use.
type Company struct {
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Batch        string   `json:"batch"`
	OneLiner     string   `json:"one_liner"`
	LongDesc     string   `json:"long_description"`
	URL          string   `json:"url"`
	Website      string   `json:"website"`
	IsHiring     bool     `json:"isHiring"`
	Industries   []string `json:"industries"`
	Tags         []string `json:"tags"`
	Regions      []string `json:"regions"`
	AllLocations string   `json:"all_locations"`
}

var client = &http.Client{Timeout: 25 * time.Second}

// LoadHiring returns hiring companies across the recent batches.
func LoadHiring(ctx context.Context) ([]Company, error) {
	return loadFiltered(ctx, func(c Company) bool { return c.IsHiring })
}

// LoadRecent returns ALL companies from the recent batches (hiring or not).
// Intended for cold-outreach flows where you email founders directly.
func LoadRecent(ctx context.Context) ([]Company, error) {
	return loadFiltered(ctx, func(Company) bool { return true })
}

func loadFiltered(ctx context.Context, keep func(Company) bool) ([]Company, error) {
	var all []Company
	for _, b := range RecentBatches {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, rawBase+b+".json", nil)
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("yc batch", b, ":", err)
			continue
		}
		var cos []Company
		if res.StatusCode == 200 {
			_ = json.NewDecoder(res.Body).Decode(&cos)
		}
		res.Body.Close()
		for _, c := range cos {
			if keep(c) {
				all = append(all, c)
			}
		}
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no YC companies loaded")
	}
	return all, nil
}

// Remote reports whether the company looks remote-friendly.
func (c Company) Remote() bool {
	blob := strings.ToLower(c.AllLocations + " " + strings.Join(c.Regions, " "))
	return strings.Contains(blob, "remote")
}

// Index maps normalized company name -> batch, for labeling jobs from other
// sources whose company happens to be a (recent) YC company.
func Index(cos []Company) map[string]string {
	m := make(map[string]string, len(cos))
	for _, c := range cos {
		m[normalize(c.Name)] = c.Batch
	}
	return m
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// drop common suffixes / YC tags so "Whip (YC W24)" matches "Whip"
	for _, cut := range []string{" (yc", ", inc", " inc.", " inc", " ltd", " llc", ".ai", ".com"} {
		if i := strings.Index(s, cut); i > 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

// Normalize is exported for callers that label jobs against the Index.
func Normalize(s string) string { return normalize(s) }
