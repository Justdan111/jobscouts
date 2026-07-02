package pipeline

import (
	"context"
	"time"

	"jobscout/internal/llm"
	"jobscout/internal/models"
	"jobscout/internal/scorer"
	"jobscout/internal/sources"
	"jobscout/internal/store"
	"jobscout/internal/yc"
)

// RefreshPostings gathers from all sources (incl. newly funded YC startups),
// classifies them (user-independent), and upserts the global postings table.
func RefreshPostings(ctx context.Context, st *store.Store) (fetched, added int, err error) {
	raw := sources.FetchAll(ctx)

	ycIndex := map[string]string{}
	if cos, e := yc.LoadHiring(ctx); e == nil {
		raw = append(raw, sources.YCToRawJobs(cos)...)
		ycIndex = yc.Index(cos)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	postings := make([]models.Posting, 0, len(raw))
	for _, r := range raw {
		blob := r.Title + " " + r.Company + " " + r.Location + " " + r.Description
		p := models.Posting{
			ID: r.ID, Source: r.Source, Title: r.Title, Company: r.Company,
			Location: r.Location, URL: r.URL, Description: r.Description, Tags: r.Tags,
			PostedAt: r.PostedAt, Eligibility: scorer.Eligibility(blob), FirstSeen: now,
		}
		classify(&p, ycIndex)
		postings = append(postings, p)
	}
	added, err = st.UpsertPostings(postings)
	return len(raw), added, err
}

// ScoreForUser computes this user's scores for all postings. Cheap keyword pass
// always; optional LLM refinement of the top candidates when useLLM is true.
func ScoreForUser(ctx context.Context, st *store.Store, c *llm.Client, userID string, useLLM bool) (int, error) {
	prof, err := st.GetProfile(userID)
	if err != nil {
		return 0, err
	}
	postings := st.AllPostings()

	type sc = struct {
		Score  int
		Reason string
	}
	scores := map[string]sc{}
	candidates := make([]models.Job, 0)
	for _, p := range postings {
		score, reason := scorer.Keyword(p, prof)
		if score < 20 && !p.NewlyFunded {
			continue // drop weak, non-startup noise
		}
		scores[p.ID] = sc{Score: score, Reason: reason}
		if score >= 30 {
			candidates = append(candidates, models.Job{Posting: p, Score: score, Reason: reason})
		}
	}

	if useLLM && c.Enabled() && len(candidates) > 0 {
		if len(candidates) > 40 {
			candidates = candidates[:40]
		}
		for id, v := range scorer.Refine(ctx, c, prof, candidates) {
			scores[id] = sc{Score: v.Score, Reason: v.Reason}
		}
	}

	if err := st.UpsertUserScores(userID, scores); err != nil {
		return 0, err
	}
	return len(scores), nil
}
