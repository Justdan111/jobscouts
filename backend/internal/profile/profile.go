// Package profile holds UNIVERSAL classification signals (used to label postings
// for everyone) and a default profile for new accounts. Per-user matching rules
// now live on each user's models.Profile, filled in via the Settings UI.
package profile

import (
	"time"

	"jobscout/internal/models"
)

// Universal eligibility signals — properties of a posting, same for all users.
var (
	USOnlySignals = []string{
		"us only", "u.s. only", "united states only", "must be based in the us",
		"must be located in the united states", "must reside in the us",
		"authorized to work in the us", "us work authorization", "us citizen",
		"based in the us", "us-based", "usa only",
	}
	GlobalSignals = []string{
		"worldwide", "global remote", "remote anywhere", "anywhere in the world",
		"work from anywhere", "fully remote", "africa", "nigeria", "emea",
		"any timezone", "global",
	}
	// Universal "stretch" seniority — penalized for everyone.
	HardSeniority = []string{
		"senior", "staff", "principal", "lead", "head of", "director", "vp ",
	}
)

// Default gives a new account broadly-useful starter preferences so the feed
// isn't empty before they finish filling things in.
func Default(userID, email string) models.Profile {
	now := time.Now().UTC().Format(time.RFC3339)
	return models.Profile{
		UserID:        userID,
		Name:          "",
		Headline:      "",
		BasedIn:       "",
		Timezone:      "",
		Links:         []models.Link{},
		Roles:         []string{"frontend", "mobile", "software engineer", "full stack"},
		Stack:         []string{"react", "typescript", "javascript"},
		SeniorityWant: []string{"junior", "entry", "intern", "graduate", "associate"},
		JobTypes:      []string{"full-time", "internship"},
		RemoteOnly:    true,
		EligibleNote:  "",
		UpdatedAt:     now,
	}
}
