package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

)

// RawJob is a posting before scoring.
type RawJob struct {
	ID, Source, Title, Company, Location, URL, Description, PostedAt string
	Tags                                                             []string
}

var client = &http.Client{Timeout: 20 * time.Second}

// FetchAll gathers postings from every source, tolerating individual failures.
func FetchAll(ctx context.Context) []RawJob {
	var out []RawJob
	if jobs, err := fetchRemoteOK(ctx); err != nil {
		fmt.Println("remoteok:", err)
	} else {
		out = append(out, jobs...)
	}
	if jobs, err := fetchHackerNews(ctx); err != nil {
		fmt.Println("hackernews:", err)
	} else {
		out = append(out, jobs...)
	}
	if jobs, err := fetchRemotive(ctx); err != nil {
		fmt.Println("remotive:", err)
	} else {
		out = append(out, jobs...)
	}
	return out
}

// ---- RemoteOK (free JSON feed; first element is a legal notice) ----

func fetchRemoteOK(ctx context.Context) ([]RawJob, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://remoteok.com/api", nil)
	req.Header.Set("User-Agent", "jobscout/0.2 (personal job agent)")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}
	var rows []struct {
		ID, Slug, Position, Company, Location, Description, URL, Date string
		Tags                                                          []string
	}
	if err := json.NewDecoder(res.Body).Decode(&rows); err != nil {
		return nil, err
	}
	var out []RawJob
	for _, r := range rows {
		if r.Position == "" || r.ID == "" {
			continue
		}
		url := r.URL
		if url == "" {
			url = "https://remoteok.com/remote-jobs/" + r.Slug
		}
		out = append(out, RawJob{
			ID: "remoteok-" + r.ID, Source: "remoteok",
			Title: strings.TrimSpace(r.Position), Company: orDefault(r.Company, "Unknown"),
			Location: orDefault(r.Location, "Remote"), URL: url,
			Description: stripHTML(r.Description), Tags: lower(r.Tags),
			PostedAt: orDefault(r.Date, time.Now().UTC().Format(time.RFC3339)),
		})
	}
	return out, nil
}

// ---- Hacker News "Who is hiring" via free Algolia API ----

func fetchHackerNews(ctx context.Context) ([]RawJob, error) {
	type hit struct {
		ObjectID, Title string
	}
	var search struct{ Hits []hit }
	if err := getJSON(ctx,
		"https://hn.algolia.com/api/v1/search_by_date?tags=story,author_whoishiring&query=who%20is%20hiring&hitsPerPage=3",
		&search); err != nil {
		return nil, err
	}
	var threadID string
	for _, h := range search.Hits {
		if strings.Contains(strings.ToLower(h.Title), "who is hiring") {
			threadID = h.ObjectID
			break
		}
	}
	if threadID == "" {
		return nil, nil
	}
	var item struct {
		Children []struct {
			ID     int    `json:"id"`
			Text   string `json:"text"`
			Author string `json:"author"`
		} `json:"children"`
	}
	if err := getJSON(ctx, "https://hn.algolia.com/api/v1/items/"+threadID, &item); err != nil {
		return nil, err
	}
	var out []RawJob
	for _, c := range item.Children {
		if c.Text == "" {
			continue
		}
		text := decodeHTML(c.Text)
		headline := firstLine(text)
		if len(headline) < 8 {
			continue
		}
		loc := "See post"
		if strings.Contains(strings.ToLower(text), "remote") {
			loc = "Remote (see post)"
		}
		out = append(out, RawJob{
			ID: fmt.Sprintf("hn-%d", c.ID), Source: "hackernews",
			Title: truncate(headline, 120), Company: guessCompany(headline),
			Location: loc, URL: fmt.Sprintf("https://news.ycombinator.com/item?id=%d", c.ID),
			Description: truncate(text, 1200), Tags: extractTags(text),
			PostedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

// ---- helpers ----

func getJSON(ctx context.Context, url string, v any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(v)
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&#x2F;", "/").Replace(s)
	return truncate(strings.Join(strings.Fields(s), " "), 1200)
}

func decodeHTML(s string) string {
	s = strings.ReplaceAll(s, "<p>", "\n")
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.NewReplacer(
		"&#x2F;", "/", "&#x27;", "'", "&quot;", "\"", "&amp;", "&", "&gt;", ">", "&lt;", "<",
	).Replace(s)
	return strings.TrimSpace(s)
}

func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

func guessCompany(headline string) string {
	if i := strings.Index(headline, "|"); i > 0 && i < 60 {
		return strings.TrimSpace(headline[:i])
	}
	return "See post"
}

func extractTags(text string) []string {
	known := []string{"react native", "react", "next.js", "typescript", "javascript",
		"go", "golang", "node", "frontend", "mobile", "full stack", "remote"}
	low := strings.ToLower(text)
	var out []string
	for _, k := range known {
		if strings.Contains(low, k) {
			out = append(out, k)
		}
	}
	return out
}

func lower(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
