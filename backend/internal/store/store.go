package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"jobscout/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

// Store is a concurrency-safe JSON-file repository. Multi-user, zero external
// services. The interface is small so SQLite/Postgres is a later drop-in swap.
type Store struct {
	mu       sync.RWMutex
	dir      string
	users    map[string]models.User        // id -> user
	byEmail  map[string]string             // email -> id
	profiles map[string]models.Profile     // userID -> profile
	postings map[string]models.Posting     // id -> posting (global)
	userJobs map[string]models.UserJob     // "userID|postingID"
	apps     map[string]models.Application // "userID|postingID"
	usage    map[string]int                // "userID|YYYY-MM-DD" -> llm calls
}

func New(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, "uploads"), 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		dir:      dir,
		users:    map[string]models.User{},
		byEmail:  map[string]string{},
		profiles: map[string]models.Profile{},
		postings: map[string]models.Posting{},
		userJobs: map[string]models.UserJob{},
		apps:     map[string]models.Application{},
		usage:    map[string]int{},
	}
	return s, s.load()
}

func (s *Store) load() error {
	_ = readJSON(s.path("users.json"), &s.users)
	for id, u := range s.users {
		s.byEmail[u.Email] = id
	}
	_ = readJSON(s.path("profiles.json"), &s.profiles)
	_ = readJSON(s.path("postings.json"), &s.postings)
	_ = readJSON(s.path("userjobs.json"), &s.userJobs)
	_ = readJSON(s.path("applications.json"), &s.apps)
	_ = readJSON(s.path("usage.json"), &s.usage)
	if len(s.postings) == 0 {
		for _, p := range seedPostings() {
			s.postings[p.ID] = p
		}
		_ = writeJSON(s.path("postings.json"), s.postings)
	}
	return nil
}

// ---- users ----

func (s *Store) CreateUser(u models.User, p models.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byEmail[u.Email]; ok {
		return ErrExists
	}
	s.users[u.ID] = u
	s.byEmail[u.Email] = u.ID
	s.profiles[u.ID] = p
	if err := writeJSON(s.path("users.json"), s.users); err != nil {
		return err
	}
	return writeJSON(s.path("profiles.json"), s.profiles)
}

func (s *Store) UserByEmail(email string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[email]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return s.users[id], nil
}

// ---- profiles ----

func (s *Store) GetProfile(userID string) (models.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[userID]
	if !ok {
		return models.Profile{}, ErrNotFound
	}
	return p, nil
}

func (s *Store) SaveProfile(p models.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.profiles[p.UserID] = p
	return writeJSON(s.path("profiles.json"), s.profiles)
}

// ---- postings (global) ----

func (s *Store) AllPostings() []models.Posting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Posting, 0, len(s.postings))
	for _, p := range s.postings {
		out = append(out, p)
	}
	return out
}

func (s *Store) UpsertPostings(in []models.Posting) (added int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range in {
		if _, ok := s.postings[p.ID]; !ok {
			added++
		}
		s.postings[p.ID] = p
	}
	return added, writeJSON(s.path("postings.json"), s.postings)
}

// ---- user jobs (per-user scoring + triage) ----

type JobFilter struct {
	Status      string
	Source      string
	MinScore    int
	HideUSOnly  bool
	YC          bool
	Funded      bool
	NewlyFunded bool
	Internship  bool
}

func ujKey(userID, postingID string) string { return userID + "|" + postingID }

// UpsertUserScores writes per-user scores, preserving existing triage status.
func (s *Store) UpsertUserScores(userID string, scores map[string]struct {
	Score  int
	Reason string
}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	for postingID, sc := range scores {
		k := ujKey(userID, postingID)
		uj, ok := s.userJobs[k]
		if !ok {
			uj = models.UserJob{UserID: userID, PostingID: postingID, Status: models.StatusNew, DiscoveredAt: now}
		}
		uj.Score = sc.Score
		uj.Reason = sc.Reason
		s.userJobs[k] = uj
	}
	return writeJSON(s.path("userjobs.json"), s.userJobs)
}

func (s *Store) ListJobs(userID string, f JobFilter) []models.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Job, 0)
	for _, p := range s.postings {
		uj, ok := s.userJobs[ujKey(userID, p.ID)]
		if !ok {
			continue // not yet scored for this user
		}
		if f.Status != "" && f.Status != "all" && string(uj.Status) != f.Status {
			continue
		}
		if f.Source != "" && f.Source != "all" && p.Source != f.Source {
			continue
		}
		if uj.Score < f.MinScore {
			continue
		}
		if f.HideUSOnly && p.Eligibility == models.EligUSOnly {
			continue
		}
		if f.YC && !p.YC {
			continue
		}
		if f.Funded && !p.Funded {
			continue
		}
		if f.NewlyFunded && !p.NewlyFunded {
			continue
		}
		if f.Internship && !p.Internship {
			continue
		}
		out = append(out, models.Job{Posting: p, Score: uj.Score, Reason: uj.Reason, Status: uj.Status, DiscoveredAt: uj.DiscoveredAt})
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Score > out[k].Score })
	return out
}

func (s *Store) GetJob(userID, postingID string) (models.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.postings[postingID]
	if !ok {
		return models.Job{}, ErrNotFound
	}
	uj := s.userJobs[ujKey(userID, postingID)]
	return models.Job{Posting: p, Score: uj.Score, Reason: uj.Reason, Status: uj.Status, DiscoveredAt: uj.DiscoveredAt}, nil
}

func (s *Store) SetJobStatus(userID, postingID string, status models.TriageStatus) (models.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.postings[postingID]
	if !ok {
		return models.Job{}, ErrNotFound
	}
	k := ujKey(userID, postingID)
	uj, ok := s.userJobs[k]
	if !ok {
		uj = models.UserJob{UserID: userID, PostingID: postingID, DiscoveredAt: time.Now().UTC().Format(time.RFC3339)}
	}
	uj.Status = status
	s.userJobs[k] = uj
	if err := writeJSON(s.path("userjobs.json"), s.userJobs); err != nil {
		return models.Job{}, err
	}
	return models.Job{Posting: p, Score: uj.Score, Reason: uj.Reason, Status: uj.Status, DiscoveredAt: uj.DiscoveredAt}, nil
}

// ---- applications ----

func (s *Store) ListApplications(userID string) []models.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Application, 0)
	for _, a := range s.apps {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].UpdatedAt > out[k].UpdatedAt })
	return out
}

func (s *Store) GetApplication(userID, postingID string) (models.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.apps[ujKey(userID, postingID)]
	if !ok {
		return models.Application{}, ErrNotFound
	}
	return a, nil
}

func (s *Store) SaveApplication(a models.Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apps[ujKey(a.UserID, a.PostingID)] = a
	return writeJSON(s.path("applications.json"), s.apps)
}

// ---- llm usage cap ----

// BumpUsage increments today's count and reports whether the user is still
// under the cap (returns true if the call is allowed).
func (s *Store) BumpUsage(userID string, cap int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := userID + "|" + time.Now().UTC().Format("2006-01-02")
	if s.usage[k] >= cap {
		return false
	}
	s.usage[k]++
	_ = writeJSON(s.path("usage.json"), s.usage)
	return true
}

// ---- helpers ----

func (s *Store) path(name string) string { return filepath.Join(s.dir, name) }

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) || len(b) == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
