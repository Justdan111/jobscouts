package pipeline

import (
	"strings"

	"jobscout/internal/models"
	"jobscout/internal/yc"
)

var vcSignals = []string{
	"pre-seed", "seed round", "series a", "series b", "series c",
	"backed by", "raised", "venture", "vc-backed", "y combinator",
}

// classify sets the user-independent label fields on a posting.
func classify(p *models.Posting, ycIndex map[string]string) {
	blob := strings.ToLower(strings.Join([]string{
		p.Title, p.Company, strings.Join(p.Tags, " "), p.Description,
	}, " "))

	p.Internship = hasTag(p.Tags, "internship") ||
		strings.Contains(strings.ToLower(p.Title), "intern")

	batch, inIndex := ycIndex[yc.Normalize(p.Company)]
	if p.Source == "yc" || inIndex {
		p.YC = true
		p.YCBatch = batch
	}
	p.NewlyFunded = p.YC

	p.Funded = p.YC
	if !p.Funded {
		for _, s := range vcSignals {
			if strings.Contains(blob, s) {
				p.Funded = true
				break
			}
		}
	}
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}
