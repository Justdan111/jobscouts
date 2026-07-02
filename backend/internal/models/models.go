package models

type Eligibility string

const (
	EligGlobal  Eligibility = "global"
	EligUSOnly  Eligibility = "us-only"
	EligUnknown Eligibility = "unknown"
)

type TriageStatus string

const (
	StatusNew       TriageStatus = "new"
	StatusSaved     TriageStatus = "saved"
	StatusDismissed TriageStatus = "dismissed"
)

type AppStage string

const (
	StageDrafted   AppStage = "drafted"
	StageApplied   AppStage = "applied"
	StageInterview AppStage = "interviewing"
	StageOffer     AppStage = "offer"
	StageRejected  AppStage = "rejected"
)

// ---- accounts ----

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	PassHash  string `json:"-"` // never serialized to clients
	CreatedAt string `json:"createdAt"`
}

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Profile is everything a friend fills in to get tailored jobs. It drives both
// scoring (the preference lists) and drafting (resume + answers).
type Profile struct {
	UserID   string `json:"userId"`
	Name     string `json:"name"`
	Headline string `json:"headline"`
	BasedIn  string `json:"basedIn"`
	Timezone string `json:"timezone"`
	Links    []Link `json:"links"`

	// Resume: pasted text is what we tailor from; file is the uploaded PDF.
	ResumeText string `json:"resumeText"`
	ResumeFile string `json:"resumeFile"` // path on disk, "" if none

	// Preferences drive the scorer.
	Roles         []string `json:"roles"`
	Stack         []string `json:"stack"`
	SeniorityWant []string `json:"seniorityWant"`
	JobTypes      []string `json:"jobTypes"` // internship, full-time, contract
	RemoteOnly    bool     `json:"remoteOnly"`
	EligibleNote  string   `json:"eligibleNote"` // "Can work from Nigeria; global/EMEA remote"

	// Free-text answers used to tailor applications.
	LookingFor   string `json:"lookingFor"`
	Strengths    string `json:"strengths"`
	Targets      string `json:"targets"`
	Constraints  string `json:"constraints"`
	ExtraContext string `json:"extraContext"`

	UpdatedAt string `json:"updatedAt"`
}

// ---- discovery ----

// Posting holds user-independent facts about a job (shared across all users).
type Posting struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	Title       string      `json:"title"`
	Company     string      `json:"company"`
	Location    string      `json:"location"`
	URL         string      `json:"url"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	PostedAt    string      `json:"postedAt"`
	Eligibility Eligibility `json:"eligibility"`
	YC          bool        `json:"yc"`
	YCBatch     string      `json:"ycBatch"`
	Funded      bool        `json:"funded"`
	NewlyFunded bool        `json:"newlyFunded"`
	Internship  bool        `json:"internship"`
	FirstSeen   string      `json:"firstSeen"`
}

// UserJob is one user's scoring + triage of a posting.
type UserJob struct {
	UserID       string       `json:"userId"`
	PostingID    string       `json:"postingId"`
	Score        int          `json:"score"`
	Reason       string       `json:"reason"`
	Status       TriageStatus `json:"status"`
	DiscoveredAt string       `json:"discoveredAt"`
}

// Job is the composed view returned by the API (posting facts + this user's state).
type Job struct {
	Posting
	Score        int          `json:"score"`
	Reason       string       `json:"reason"`
	Status       TriageStatus `json:"status"`
	DiscoveredAt string       `json:"discoveredAt"`
}

// ---- applications ----

type Application struct {
	UserID           string   `json:"userId"`
	PostingID        string   `json:"postingId"`
	Stage            AppStage `json:"stage"`
	ResumeHighlights string   `json:"resumeHighlights"`
	CoverEmail       string   `json:"coverEmail"`
	Pitch            string   `json:"pitch"`
	Notes            string   `json:"notes"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
}
