package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jobscout/internal/apply"
	"jobscout/internal/auth"
	"jobscout/internal/config"
	"jobscout/internal/llm"
	"jobscout/internal/mail"
	"jobscout/internal/models"
	"jobscout/internal/pipeline"
	"jobscout/internal/profile"
	"jobscout/internal/store"
	"jobscout/internal/yc"
)

type Server struct {
	cfg config.Config
	st  *store.Store
	llm *llm.Client
}

func NewServer(cfg config.Config, st *store.Store, c *llm.Client) http.Handler {
	s := &Server{cfg: cfg, st: st, llm: c}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.health)
	// auth (public)
	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	// profile (auth)
	mux.HandleFunc("GET /api/profile", s.requireAuth(s.getProfile))
	mux.HandleFunc("PUT /api/profile", s.requireAuth(s.putProfile))
	mux.HandleFunc("POST /api/profile/resume", s.requireAuth(s.uploadResume))
	// jobs (auth)
	mux.HandleFunc("GET /api/jobs", s.requireAuth(s.listJobs))
	mux.HandleFunc("PATCH /api/jobs/{id}", s.requireAuth(s.patchJob))
	mux.HandleFunc("POST /api/jobs/{id}/draft", s.requireAuth(s.draft))
	mux.HandleFunc("POST /api/refresh", s.requireAuth(s.refresh))
	// startups for cold-outreach (auth)
	mux.HandleFunc("GET /api/companies", s.requireAuth(s.listCompanies))
	// applications (auth)
	mux.HandleFunc("GET /api/applications", s.requireAuth(s.listApplications))
	mux.HandleFunc("PUT /api/applications/{id}", s.requireAuth(s.putApplication))
	mux.HandleFunc("POST /api/digest", s.requireAuth(s.digest))

	return logging(cors(s.cfg.AllowedOrigin, mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "llm": s.llm.Enabled()})
}

// ---- auth ----

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var b struct{ Email, Password, Invite string }
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
	if b.Email == "" || len(b.Password) < 8 {
		writeErr(w, 400, "email and an 8+ char password are required")
		return
	}
	if len(s.cfg.InviteCodes) > 0 && !s.cfg.InviteCodes[b.Invite] {
		writeErr(w, 403, "a valid invite code is required")
		return
	}
	hash, err := auth.HashPassword(b.Password)
	if err != nil {
		writeErr(w, 500, "could not hash password")
		return
	}
	id := newID()
	user := models.User{ID: id, Email: b.Email, PassHash: hash, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := s.st.CreateUser(user, profile.Default(id, b.Email)); err != nil {
		if errors.Is(err, store.ErrExists) {
			writeErr(w, 409, "an account with that email already exists")
			return
		}
		writeErr(w, 500, err.Error())
		return
	}
	s.issue(w, user)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var b struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	user, err := s.st.UserByEmail(strings.ToLower(strings.TrimSpace(b.Email)))
	if err != nil || !auth.VerifyPassword(b.Password, user.PassHash) {
		writeErr(w, 401, "wrong email or password")
		return
	}
	s.issue(w, user)
}

func (s *Server) issue(w http.ResponseWriter, user models.User) {
	tok, err := auth.IssueToken(s.cfg.JWTSecret, user.ID, 14*24*time.Hour)
	if err != nil {
		writeErr(w, 500, "could not issue token")
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "user": map[string]string{"id": user.ID, "email": user.Email}})
}

// ---- profile ----

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	p, err := s.st.GetProfile(uid(r))
	if err != nil {
		writeErr(w, 404, "profile not found")
		return
	}
	writeJSON(w, 200, map[string]any{"profile": p})
}

func (s *Server) putProfile(w http.ResponseWriter, r *http.Request) {
	var p models.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	current, err := s.st.GetProfile(uid(r))
	if err != nil {
		writeErr(w, 404, "profile not found")
		return
	}
	p.UserID = uid(r)
	p.ResumeFile = current.ResumeFile // file is managed by the upload endpoint
	if err := s.st.SaveProfile(p); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"profile": p})
}

func (s *Server) uploadResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(8 << 20); err != nil { // 8MB cap
		writeErr(w, 400, "file too large or invalid (8MB max)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, 400, "no file provided")
		return
	}
	defer file.Close()

	dir := filepath.Join(s.cfg.DataDir, "uploads", uid(r))
	_ = os.MkdirAll(dir, 0o755)
	name := filepath.Base(header.Filename)
	dst := filepath.Join(dir, name)
	out, err := os.Create(dst)
	if err != nil {
		writeErr(w, 500, "could not save file")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		writeErr(w, 500, "could not write file")
		return
	}

	p, _ := s.st.GetProfile(uid(r))
	p.ResumeFile = dst
	_ = s.st.SaveProfile(p)
	writeJSON(w, 200, map[string]any{"resumeFile": name})
}

// ---- jobs ----

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	// keyword-score new postings for this user (cheap, no LLM), then list.
	_, _ = pipeline.ScoreForUser(r.Context(), s.st, s.llm, uid(r), false)

	q := r.URL.Query()
	minScore, _ := strconv.Atoi(q.Get("minScore"))
	jobs := s.st.ListJobs(uid(r), store.JobFilter{
		Status: q.Get("status"), Source: q.Get("source"), MinScore: minScore,
		HideUSOnly: q.Get("hideUsOnly") == "1", YC: q.Get("yc") == "1",
		Funded: q.Get("funded") == "1", NewlyFunded: q.Get("newlyFunded") == "1",
		Internship: q.Get("internship") == "1",
	})
	writeJSON(w, 200, map[string]any{"jobs": jobs})
}

func (s *Server) patchJob(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Status models.TriageStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.Status == "" {
		writeErr(w, 400, "status is required")
		return
	}
	job, err := s.st.SetJobStatus(uid(r), r.PathValue("id"), b.Status)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, 404, "job not found")
		return
	}
	writeJSON(w, 200, map[string]any{"job": job})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	fetched, added, err := pipeline.RefreshPostings(r.Context(), s.st)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	scored, err := pipeline.ScoreForUser(r.Context(), s.st, s.llm, uid(r), true)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"fetched": fetched, "added": added, "scored": scored})
}

func (s *Server) draft(w http.ResponseWriter, r *http.Request) {
	if s.llm.Enabled() && !s.st.BumpUsage(uid(r), s.cfg.DailyLLMCap) {
		writeErr(w, 429, "daily AI limit reached — try again tomorrow")
		return
	}
	id := r.PathValue("id")
	job, err := s.st.GetJob(uid(r), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, 404, "job not found")
		return
	}
	prof, _ := s.st.GetProfile(uid(r))
	d, err := apply.Generate(r.Context(), s.llm, prof, job)
	if err != nil {
		writeErr(w, 422, err.Error())
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	app := models.Application{
		UserID: uid(r), PostingID: id, Stage: models.StageDrafted,
		ResumeHighlights: d.ResumeHighlights, CoverEmail: d.CoverEmail, Pitch: d.Pitch,
		CreatedAt: now, UpdatedAt: now,
	}
	if ex, err := s.st.GetApplication(uid(r), id); err == nil {
		app.CreatedAt = ex.CreatedAt
		app.Notes = ex.Notes
		if ex.Stage != "" && ex.Stage != models.StageDrafted {
			app.Stage = ex.Stage
		}
	}
	_ = s.st.SaveApplication(app)
	writeJSON(w, 200, map[string]any{"application": app})
}

// ---- startups (cold-outreach) ----

// listCompanies returns recent-batch YC companies (hiring or not) — the raw
// list for founder outreach. Optional ?hiringOnly=1 narrows to c.IsHiring.
func (s *Server) listCompanies(w http.ResponseWriter, r *http.Request) {
	cos, err := yc.LoadRecent(r.Context())
	if err != nil {
		writeErr(w, 502, "could not load YC data")
		return
	}
	hiringOnly := r.URL.Query().Get("hiringOnly") == "1"
	type outCompany struct {
		Name       string   `json:"name"`
		Slug       string   `json:"slug"`
		Batch      string   `json:"batch"`
		OneLiner   string   `json:"oneLiner"`
		LongDesc   string   `json:"longDesc"`
		Website    string   `json:"website"`
		YCURL      string   `json:"ycUrl"`
		Location   string   `json:"location"`
		Industries []string `json:"industries"`
		IsHiring   bool     `json:"isHiring"`
		Remote     bool     `json:"remote"`
	}
	out := make([]outCompany, 0, len(cos))
	for _, c := range cos {
		if hiringOnly && !c.IsHiring {
			continue
		}
		loc := c.AllLocations
		if loc == "" {
			loc = "See company"
		}
		out = append(out, outCompany{
			Name: c.Name, Slug: c.Slug, Batch: c.Batch, OneLiner: c.OneLiner,
			LongDesc: c.LongDesc, Website: c.Website, YCURL: c.URL,
			Location: loc, Industries: c.Industries, IsHiring: c.IsHiring,
			Remote: c.Remote(),
		})
	}
	writeJSON(w, 200, map[string]any{"companies": out})
}

// ---- applications ----

func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"applications": s.st.ListApplications(uid(r))})
}

func (s *Server) putApplication(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var b struct {
		Stage            models.AppStage `json:"stage"`
		Notes            string          `json:"notes"`
		ResumeHighlights string          `json:"resumeHighlights"`
		CoverEmail       string          `json:"coverEmail"`
		Pitch            string          `json:"pitch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	app, err := s.st.GetApplication(uid(r), id)
	if errors.Is(err, store.ErrNotFound) {
		app = models.Application{UserID: uid(r), PostingID: id, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	}
	if b.Stage != "" {
		app.Stage = b.Stage
	}
	if b.ResumeHighlights != "" {
		app.ResumeHighlights = b.ResumeHighlights
	}
	if b.CoverEmail != "" {
		app.CoverEmail = b.CoverEmail
	}
	if b.Pitch != "" {
		app.Pitch = b.Pitch
	}
	app.Notes = b.Notes
	app.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_ = s.st.SaveApplication(app)
	writeJSON(w, 200, map[string]any{"application": app})
}

func (s *Server) digest(w http.ResponseWriter, r *http.Request) {
	jobs := s.st.ListJobs(uid(r), store.JobFilter{})
	sent, err := mail.SendDigest(r.Context(), jobs)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"sent": sent})
}
