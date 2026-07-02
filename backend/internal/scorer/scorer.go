package scorer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jobscout/internal/llm"
	"jobscout/internal/models"
	"jobscout/internal/profile"
)

// Eligibility is a property of the posting (user-independent). Computed once at
// classification time.
func Eligibility(text string) models.Eligibility {
	t := strings.ToLower(text)
	for _, s := range profile.USOnlySignals {
		if strings.Contains(t, s) {
			return models.EligUSOnly
		}
	}
	for _, s := range profile.GlobalSignals {
		if strings.Contains(t, s) {
			return models.EligGlobal
		}
	}
	return models.EligUnknown
}

// Keyword scores a posting against ONE user's profile. Returns 0-100 + a reason.
func Keyword(p models.Posting, prof models.Profile) (int, string) {
	title := strings.ToLower(p.Title)
	blob := strings.ToLower(strings.Join([]string{
		p.Title, p.Company, p.Location, strings.Join(p.Tags, " "), p.Description,
	}, " "))

	score := 0
	var why []string

	if hit := firstHit(title, lower(prof.Roles)); hit != "" {
		score += 35
		why = append(why, "role: "+hit)
	}
	if hits := allHits(blob, lower(prof.Stack)); len(hits) > 0 {
		score += min(30, len(hits)*12)
		why = append(why, "stack: "+strings.Join(firstN(hits, 3), ", "))
	}
	if hit := firstHit(blob, lower(prof.SeniorityWant)); hit != "" {
		score += 15
		why = append(why, "level: "+hit)
	}
	if hit := firstHit(blob, profile.HardSeniority); hit != "" {
		score -= 20
		why = append(why, "stretch: "+hit)
	}

	switch p.Eligibility {
	case models.EligGlobal:
		score += 15
		why = append(why, "global-friendly")
	case models.EligUSOnly:
		score -= 30
		why = append(why, "US-only")
	}

	if strings.Contains(blob, "react native") || strings.Contains(blob, "expo") {
		if contains(lower(prof.Stack), "react native") || contains(lower(prof.Stack), "expo") {
			score += 15
			why = append(why, "react native")
		}
	}
	if p.Internship && contains(lower(prof.JobTypes), "internship") {
		score += 10
		why = append(why, "internship match")
	}

	score = clamp(score, 0, 100)
	reason := "weak match"
	if len(why) > 0 {
		reason = strings.Join(why, " · ")
	}
	return score, reason
}

// Refine asks Claude to re-judge a user's strongest candidates. No-op without a key.
type verdict struct {
	ID     string `json:"id"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

func Refine(ctx context.Context, c *llm.Client, prof models.Profile, jobs []models.Job) map[string]struct {
	Score  int
	Reason string
} {
	out := map[string]struct {
		Score  int
		Reason string
	}{}
	if !c.Enabled() || len(jobs) == 0 {
		return out
	}
	type item struct{ ID, Title, Company, Location, Snippet string }
	list := make([]item, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, item{j.ID, j.Title, j.Company, j.Location, truncate(j.Description, 500)})
	}
	payload, _ := json.Marshal(list)

	prof.Name = orStr(prof.Name, "the candidate")
	prompt := fmt.Sprintf(`Score job posts for this candidate. Return ONLY a JSON array, no prose.

Candidate: %s, based in %s. Wants: %s. Stack: %s. Levels: %s.
Eligibility note: %s

For each job return {"id","score" (0-100 fit),"reason" (<=12 words)}.
Penalize roles that don't fit the wanted roles/levels.

Jobs:
%s`, prof.Name, orStr(prof.BasedIn, "unspecified"),
		strings.Join(prof.Roles, ", "), strings.Join(prof.Stack, ", "),
		strings.Join(prof.SeniorityWant, ", "), orStr(prof.EligibleNote, "n/a"), string(payload))

	text, err := c.Complete(ctx, prompt, 1500)
	if err != nil {
		fmt.Println("llm refine:", err)
		return out
	}
	start, end := strings.Index(text, "["), strings.LastIndex(text, "]")
	if start < 0 || end <= start {
		return out
	}
	var vs []verdict
	if err := json.Unmarshal([]byte(text[start:end+1]), &vs); err != nil {
		return out
	}
	for _, v := range vs {
		out[v.ID] = struct {
			Score  int
			Reason string
		}{v.Score, v.Reason}
	}
	return out
}

func firstHit(hay string, needles []string) string {
	for _, n := range needles {
		if n != "" && strings.Contains(hay, n) {
			return n
		}
	}
	return ""
}
func allHits(hay string, needles []string) []string {
	var out []string
	for _, n := range needles {
		if n != "" && strings.Contains(hay, n) {
			out = append(out, n)
		}
	}
	return out
}
func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
func lower(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(strings.TrimSpace(s))
	}
	return out
}
func firstN(s []string, n int) []string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
func orStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
