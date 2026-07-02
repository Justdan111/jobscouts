package sources

import (
	"strings"
	"time"

	"jobscout/internal/yc"
)

// YCToRawJobs turns hiring YC companies into feed entries that link to each
// company's YC jobs page. These are company-level "is hiring" signals — the
// most direct way to surface newly funded startups with open roles.
func YCToRawJobs(cos []yc.Company) []RawJob {
	now := time.Now().UTC().Format(time.RFC3339)
	out := make([]RawJob, 0, len(cos))
	for _, c := range cos {
		loc := "Remote-friendly"
		if !c.Remote() {
			if c.AllLocations != "" {
				loc = c.AllLocations
			} else {
				loc = "See company"
			}
		}
		desc := c.OneLiner
		if c.LongDesc != "" {
			desc = c.OneLiner + " — " + c.LongDesc
		}
		tags := append([]string{"startup", "yc", strings.ToLower(c.Batch)}, lower(c.Industries)...)
		out = append(out, RawJob{
			ID:          "yc-" + c.Slug,
			Source:      "yc",
			Title:       c.Name + " — open roles (" + c.Batch + ")",
			Company:     c.Name,
			Location:    loc,
			URL:         strings.TrimRight(c.URL, "/") + "/jobs",
			Description: truncate(desc, 1200),
			Tags:        tags,
			PostedAt:    now,
		})
	}
	return out
}
