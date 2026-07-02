# JobScout — Friends Edition (multi-user, zero paid services)

A self-hosted job agent your friends can log into. Each person fills in their own
profile — résumé, portfolio links, and tailoring questions — and gets jobs scored
for *them*, plus applications drafted from *their* résumé. Invite-only.

**No paid services.** Go API + Next.js UI, a JSON file store on disk, résumé files
on disk, and auth (passwords + sessions) implemented with the Go standard library.
No database service, no auth vendor, no object store. It runs on one machine with
a filesystem — your own always-on box, a Raspberry Pi, or a free-forever VM
(e.g. Oracle Cloud Always Free).

## What each friend gets

- **A profile they fill in** (Profile tab): name, headline, location, **portfolio
  links**, **résumé** (paste text + optional file upload), **preferences** (roles,
  stack, seniority, job types, remote-only, eligibility note), and **questions**
  (what they want, strengths, targets, constraints, extra context).
- **Discovery scored for them** — the newly-funded-startup feed (YC + others),
  ranked against *their* profile.
- **Apply now** — tailored résumé highlights + cover email + pitch, grounded in
  *their* résumé and answers. Tracked in their own pipeline.
- Their data is **isolated** — every request is scoped to their account.

## Run it

Backend (Go 1.22+):
```
cd backend
cp .env.example .env       # set JWT_SECRET and INVITE_CODES; ANTHROPIC_API_KEY optional
go run ./cmd/server        # http://localhost:8080
```

Frontend (Node 18+):
```
cd frontend
cp .env.example .env.local # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install && npm run dev # http://localhost:3000
```

Open the app -> Sign up with an invite code (from INVITE_CODES) -> fill in your
Profile -> Run scan on Discover -> Apply now on any job.

## Inviting friends

Set INVITE_CODES=code1,code2 in the backend .env and share a code. Anyone without
a valid code is rejected at signup. Leave INVITE_CODES empty to go open later.
Discovery and scoring work with no AI key; Apply now needs ANTHROPIC_API_KEY on
the server (see Cost).

## Cost control (shared AI key)

If you set one ANTHROPIC_API_KEY, all friends share it. DAILY_LLM_CAP caps how many
AI drafts each user can run per day (default 25), so no one runs it away. Discovery
and keyword scoring are always free and uncapped.

Upgrade seam: to make each user bring their own key later, add a key field in the
profile and prefer it over the shared key.

## Security notes (small, but real)

- Every request is scoped to the JWT's user id; users only see their own data.
  Passwords are hashed with PBKDF2-HMAC-SHA256 (stdlib). Sessions are signed JWTs
  (HS256, stdlib).
- Set a strong JWT_SECRET and serve over HTTPS in production.
- Appropriate for a few trusted friends. Before going public, see the upgrade docs
  (bcrypt/argon2, a real database, rate limiting, a privacy policy).

## Backend layout

```
cmd/server/main.go      entrypoint
internal/
  auth/                 PBKDF2 password hashing + HS256 JWTs (stdlib only)
  config/               env (JWT secret, invite codes, AI cap)
  models/               User, Profile, Posting, UserJob, Job, Application
  profile/              universal signals + default profile for new accounts
  store/                multi-user JSON store -> swap for SQLite/Postgres later
  sources/ yc/          RemoteOK, Remotive, HN, YC startups
  scorer/               per-user keyword scoring + eligibility + LLM refine
  apply/                tailored resume/cover/pitch from the user's profile
  pipeline/             RefreshPostings (global) + ScoreForUser (per-user)
  mail/                 optional Resend digest (off unless a key is set)
  httpx/                auth middleware, scoped handlers, upload, router
```

## Upgrading later

The store sits behind a small interface, so the JSON files swap for SQLite or
Postgres without touching the rest. See JobScout-v3-MULTIUSER-SPEC.md and
JobScout-FRIENDS-EDITION.md for the full upgrade path.

## Honest scope

This is the friends build: invite-only, file-based storage, shared-key-with-cap,
and stdlib auth. Deliberately not hardened for strangers — that's the documented
next step, and the architecture is built to absorb it.
# jobscouts
