package apply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"jobscout/internal/llm"
	"jobscout/internal/models"
)

type Draft struct {
	ResumeHighlights string `json:"resumeHighlights"`
	CoverEmail       string `json:"coverEmail"`
	Pitch            string `json:"pitch"`
}

// Generate writes tailored material grounded in THIS user's profile (resume text
// + their answers). Requires an Anthropic key and a non-empty resume.
func Generate(ctx context.Context, c *llm.Client, prof models.Profile, job models.Job) (Draft, error) {
	if !c.Enabled() {
		return Draft{}, errors.New("application drafting is disabled (no Anthropic key configured)")
	}
	if strings.TrimSpace(prof.ResumeText) == "" {
		return Draft{}, errors.New("add your résumé text in Settings first — that's what we tailor from")
	}

	links := make([]string, 0, len(prof.Links))
	for _, l := range prof.Links {
		links = append(links, l.Label+": "+l.URL)
	}

	prompt := fmt.Sprintf(`Help this candidate apply for a job. Honest, specific, non-generic. No clichés, no invented achievements — use ONLY facts from the candidate's material. Reorder and emphasize; never fabricate.

=== CANDIDATE ===
Name: %s
Headline: %s
Based in: %s  (%s)
Links: %s
Résumé:
%s

What they're looking for: %s
Their strengths: %s
Targets: %s
Constraints: %s
Extra context: %s

=== JOB ===
Title: %s
Company: %s
Location: %s
Details: %s

=== TASK ===
Return ONLY a JSON object, no prose, no markdown:
{
  "resumeHighlights": "A tailored résumé header for THIS job: a 2-sentence summary aimed at this role, then 4-5 bullet lines (each starting with '• ') reordering/emphasizing the candidate's most relevant real experience. Facts only.",
  "coverEmail": "A complete cover email (first line 'Subject: ...', then body). 130-180 words. Specific to this company/role, connecting 2-3 real strengths to what the job needs. Sign off with the candidate's name and links.",
  "pitch": "A 2-3 sentence message for a quick application box or DM."
}`,
		orStr(prof.Name, "(unnamed)"), orStr(prof.Headline, ""), orStr(prof.BasedIn, ""), orStr(prof.EligibleNote, ""),
		strings.Join(links, " · "), trunc(prof.ResumeText, 2500),
		orStr(prof.LookingFor, "—"), orStr(prof.Strengths, "—"), orStr(prof.Targets, "—"),
		orStr(prof.Constraints, "—"), orStr(prof.ExtraContext, "—"),
		job.Title, job.Company, job.Location, trunc(job.Description, 900))

	text, err := c.Complete(ctx, prompt, 1600)
	if err != nil {
		return Draft{}, err
	}
	a, b := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if a < 0 || b <= a {
		return Draft{}, fmt.Errorf("unexpected model output")
	}
	var d Draft
	if err := json.Unmarshal([]byte(text[a:b+1]), &d); err != nil {
		return Draft{}, err
	}
	return d, nil
}

func trunc(s string, n int) string {
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
