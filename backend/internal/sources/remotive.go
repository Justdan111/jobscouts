package sources

import (
	"context"
	"strings"
	"time"
)

// Remotive publishes a free remote-jobs API. It gives us a clean job_type
// (so we can flag internships) and a candidate_required_location string that
// maps neatly onto eligibility.
func fetchRemotive(ctx context.Context) ([]RawJob, error) {
	var resp struct {
		Jobs []struct {
			ID                        int      `json:"id"`
			Title                     string   `json:"title"`
			CompanyName               string   `json:"company_name"`
			CandidateRequiredLocation string   `json:"candidate_required_location"`
			JobType                   string   `json:"job_type"`
			URL                       string   `json:"url"`
			Tags                      []string `json:"tags"`
			Description               string   `json:"description"`
			PublicationDate           string   `json:"publication_date"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, "https://remotive.com/api/remote-jobs?category=software-dev&limit=100", &resp); err != nil {
		return nil, err
	}
	out := make([]RawJob, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		tags := lower(j.Tags)
		// Surface internship as a tag so downstream classification is trivial.
		if strings.Contains(strings.ToLower(j.JobType), "intern") {
			tags = append(tags, "internship")
		}
		out = append(out, RawJob{
			ID: "remotive-" + itoa(j.ID), Source: "remotive",
			Title: strings.TrimSpace(j.Title), Company: orDefault(j.CompanyName, "Unknown"),
			Location: orDefault(j.CandidateRequiredLocation, "Remote"), URL: j.URL,
			Description: stripHTML(j.Description), Tags: tags,
			PostedAt: orDefault(j.PublicationDate, time.Now().UTC().Format(time.RFC3339)),
		})
	}
	return out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
