# ARMSS Poster Studio — Requirements

A web application for ARMSS Employment Solutions that turns a short description of an
occasion into a finished, on-brand poster. A pipeline of LLM agents writes the brief,
copy and art direction; an image-generation provider produces the artwork; the backend
composites the ARMSS logo and contact footer so brand assets are always pixel-exact.

**Stack:** Go backend, React + TypeScript frontend, PostgreSQL, S3-compatible object storage.

---

## 1. Why this exists

Marketing staff currently brief a designer for every festival and announcement poster.
The output is good but slow, and the house style lives in one person's head. This tool
encodes the house style, lets non-designers produce a first-pass poster in minutes, and
guarantees the logo, tagline and phone numbers are never garbled.

**The critical constraint:** image models cannot render the ARMSS logo. A previous
attempt produced "OARMSS". The logo, tagline and phone numbers are therefore **never**
generated — the backend overlays them from vector/PNG source files after the artwork
returns. Any design that asks the image model to draw brand assets is wrong.

---

## 2. Users and roles

| Role | Can do |
|---|---|
| `staff` | Create posters, browse the template library, download results |
| `admin` | Everything above, plus manage the template library, brand assets, and view cost reports |

Single-tenant. Email + password auth with sessions is sufficient; no SSO required for v1.

---

## 3. The three ways a user starts

All three converge on the same pipeline. This is the core of the product and must be
built as specified.

### Flow A — Pick a template (the primary path)

The user opens a gallery of pre-uploaded ARMSS designs (past Independence Day, Diwali,
Pongal, hiring announcement, etc). They click one, type the occasion in a single text
box ("Diwali 2026"), and press Generate. The selected template becomes the style
reference; its extracted style profile is already cached, so the pipeline skips straight
to the brief.

### Flow B — Describe it, get recommendations

The user types a description with no template selected. Before running the pipeline the
system recommends the three closest templates from the library, with a one-line reason
each. The user picks one, or chooses "no reference" and continues.

### Flow C — Attach your own reference

The user attaches 1–4 of their own images. These are analysed on the fly, exactly like a
library template, and are optionally saved to the library by an admin.

The input area must offer all three simultaneously: a template gallery, a text box, and
a file attach button. A user can combine them — pick a template *and* attach an extra
reference.

---

## 4. Domain model

```
Template        a pre-uploaded reference design + its cached StyleProfile
StyleProfile    structured description of a house style (palette, type, layout, furniture)
BrandAsset      logo, tagline lockup, footer bar — versioned, admin-managed
Project         one poster job, owned by a user
Brief           structured requirements, produced by the intake agent
Concept         one of three creative directions
Copy            headline, subhead, CTA, alternates
ArtDirection    layout spec + the image-generation prompts
Render          one generated image + its composited final version
Review          QA scoring of a render against brief and style profile
```

A `Project` moves through states:
`draft → briefing → concepts_ready → developing → render_ready → composited → approved`

Every agent output is persisted. The user must be able to edit any stage and re-run only
the stages downstream of it. This is the main reason for splitting the pipeline into
agents at all — do not collapse it into one prompt.

---

## 5. The agent pipeline

Seven agents, run server-side. Each is a single Anthropic API call with a strict output
schema. Use **tool use with an input schema** to enforce shape — do not prompt for JSON
and parse the text, which fails on truncation.

Model: `claude-sonnet-5`. Config-driven so it can be changed without a deploy.

| # | Agent | Input | Output | Notes |
|---|---|---|---|---|
| 1 | Style Analyst | Reference images | `StyleProfile` | Vision call. Cached per template — never re-run for library templates. |
| 2 | Intake | User text + StyleProfile | `Brief` + open questions + assumptions | Must fill gaps with assumptions rather than blocking. |
| 3 | Creative Director | Brief + StyleProfile | 3 × `Concept` | Three genuinely different big ideas, not three colourways. |
| 4 | Copywriter | Brief + chosen Concept | `Copy` | Short, active voice, no exclamation marks. |
| 5 | Art Director | Brief + Concept + Copy + StyleProfile | `ArtDirection` | Produces both prompt variants — see §6. |
| 6 | Layout Proof | ArtDirection | SVG wireframe | Optional but cheap. Catches hollow compositions before spending image credits. |
| 7 | Reviewer | Everything + the rendered image | `Review` | Vision call. Scores against brief *and* StyleProfile, lists fixes. |

Token budgets: agents 1–5 and 7 need 3000–4000 max output tokens. Agent 6 needs 8000.
A 1000-token cap truncates output and is the single most likely source of bugs.

### Output schemas

```go
type StyleProfile struct {
    Summary        string       `json:"summary"`
    Palette        []PaletteHex `json:"palette"`        // hex + role
    Typography     string       `json:"typography"`
    LayoutPatterns string       `json:"layoutPatterns"`
    Imagery        string       `json:"imagery"`
    Finish         string       `json:"finish"`         // photoreal composite / 3D / painting
    Tone           string       `json:"tone"`
    LogoTreatment  string       `json:"logoTreatment"`
    FixedElements  []string     `json:"fixedElements"`  // furniture repeating on every poster
    Consistency    string       `json:"consistency"`
}

type Brief struct {
    Objective   string   `json:"objective"`
    Audience    string   `json:"audience"`
    KeyMessage  string   `json:"keyMessage"`
    Occasion    string   `json:"occasion"`
    EventDate   string   `json:"eventDate"`
    Tone        []string `json:"tone"`
    Format      string   `json:"format"`
    Dimensions  string   `json:"dimensions"`
    Palette     []string `json:"palette"`
    MustInclude []string `json:"mustInclude"`
    Avoid       []string `json:"avoid"`
    Assumptions []string `json:"assumptions"`
    Questions   []string `json:"questions"`
}

type Concept struct {
    Name           string `json:"name"`
    BigIdea        string `json:"bigIdea"`
    VisualMetaphor string `json:"visualMetaphor"`
    Mood           string `json:"mood"`
    Palette        string `json:"palette"`
}

type Copy struct {
    Headline   string   `json:"headline"`
    Subhead    string   `json:"subhead"`
    CTA        string   `json:"cta"`
    ValueWords []string `json:"valueWords"`   // e.g. LIGHT / PROSPERITY / JOY
    Alternates []string `json:"alternates"`
}

type ArtDirection struct {
    Layout           string   `json:"layout"`
    FocalPoint       string   `json:"focalPoint"`
    Typography       string   `json:"typography"`
    Colour           string   `json:"colour"`
    TextToRender     []string `json:"textToRender"`
    FullPosterPrompt string   `json:"fullPosterPrompt"`
    BackgroundPrompt string   `json:"backgroundPrompt"`
    NegativePrompt   string   `json:"negativePrompt"`
    AspectRatio      string   `json:"aspectRatio"`
    ReservedZones    []Zone   `json:"reservedZones"`   // see §6
    CompositeSteps   []string `json:"compositeSteps"`
}

type Review struct {
    Score     int      `json:"score"`
    Verdict   string   `json:"verdict"`
    HouseMatch string  `json:"houseMatch"`
    Working   []string `json:"working"`
    Risks     []string `json:"risks"`
    Fixes     []string `json:"fixes"`
}
```

---

## 6. Reserved zones and logo compositing

This is the feature that makes the output usable, so specify it precisely.

The Art Director must emit `reservedZones` — rectangles, in normalised 0–1 coordinates,
that the image prompt explicitly asks to keep clear:

```go
type Zone struct {
    ID     string  `json:"id"`      // "logo" | "tagline" | "footer"
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    W      float64 `json:"w"`
    H      float64 `json:"h"`
    Fill   string  `json:"fill"`    // hex the background should be here
    Note   string  `json:"note"`    // what gets placed, for the prompt text
}
```

The generated `fullPosterPrompt` must instruct the model to leave those zones clean and
uncluttered, and the `negativePrompt` must include `logo, brand marks, wordmark, watermark, text in the top band`.

After the image returns, the Go backend composites brand assets into those zones:

- Load the active `BrandAsset` files (logo PNG with alpha, tagline lockup, footer bar).
- Scale the logo to fit its zone preserving aspect ratio, centre it, apply a soft drop shadow.
- Draw the footer bar as a solid rect in the brand navy, then place the phone icon and the
  contact string from config using a bundled font.
- Write the composited result as a new `Render` row with `kind = "composited"`, keeping the
  raw generation for comparison.

Use `github.com/disintegration/imaging` or `golang.org/x/image/draw` plus
`github.com/golang/freetype` for text. Bundle the brand font in the repo; do not rely on
system fonts.

The user must be able to toggle between raw and composited output, and to nudge the logo
zone if the artwork crowds it.

---

## 7. Image generation providers

Anthropic's API does not generate images. The pipeline's thinking and the pixels come
from different vendors, so put the image step behind an interface.

```go
type ImageProvider interface {
    Name() string
    SupportsTextRendering() bool
    Generate(ctx context.Context, req GenerateRequest) ([]GeneratedImage, error)
    Edit(ctx context.Context, req EditRequest) ([]GeneratedImage, error) // optional
}
```

Implement at least one of Ideogram, OpenAI `gpt-image-1`, Google Gemini image, or Flux.
Selection is per-request with a server default. When
`provider.SupportsTextRendering() == false`, send `BackgroundPrompt` instead of
`FullPosterPrompt` and render the headline type in the compositing step.

Providers fail and rate-limit often. Wrap calls with retry-with-backoff, a hard timeout,
and a clear user-facing message naming the provider that failed. Never silently fall
back to a different provider — cost and style both change.

---

## 8. Template library and recommendations

Seed the library with ARMSS's existing designs. Each `Template` row carries:
image key, title, occasion tags, season, a cached `StyleProfile`, and `is_active`.

Admin upload flow: attach image → run Style Analyst → admin reviews and edits the
extracted profile → save. The profile is editable by hand because an admin's correction
beats a model's guess, and it's cached so it costs nothing at generation time.

**Recommendation** (Flow B): given the user's free text, rank templates and return the
top three with a reason. Implement in this order of preference:

1. Tag and keyword overlap on occasion, season and format. Cheap, deterministic, good enough.
2. If that returns fewer than three, one LLM call ranking template titles and summaries
   against the user's text.

Do not build a vector database for v1. The library will hold tens of items, not thousands.

---

## 9. HTTP API

REST, JSON, `/api/v1`. Long-running work streams over Server-Sent Events so the UI can
show each agent completing rather than one long spinner.

```
POST   /auth/login                      → session cookie
POST   /auth/logout

GET    /templates                       ?occasion=&q=      list active templates
GET    /templates/:id
POST   /templates                       admin: upload + analyse
PATCH  /templates/:id                   admin: edit profile / deactivate

GET    /brand-assets                    active logo, tagline, footer config
PUT    /brand-assets                    admin: replace

POST   /uploads                          multipart, returns an upload id + signed URL
POST   /recommendations                  { text } → 3 templates + reasons

POST   /projects                         { text, templateId?, uploadIds[] } → project id
GET    /projects                         list mine
GET    /projects/:id                     full state, all agent outputs
GET    /projects/:id/events              SSE: stage progress + payloads

POST   /projects/:id/brief/rerun         { extraDetail } re-runs intake and downstream
POST   /projects/:id/concepts/:n/select  chooses a direction, runs agents 4–6
PATCH  /projects/:id/copy                manual copy edits
PATCH  /projects/:id/art-direction       manual prompt edits
POST   /projects/:id/render              { provider? } calls the image provider
POST   /projects/:id/composite           { zoneOverrides? } overlays brand assets
POST   /projects/:id/review              runs the reviewer against a render
GET    /projects/:id/download            ?kind=composited  final file

GET    /admin/usage                      token + image spend by user and day
```

SSE event shape:

```json
{ "stage": "art_direction", "status": "done", "payload": { }, "tokensIn": 0, "tokensOut": 0 }
```

`status` is one of `running | done | error`. Send a heartbeat every 15s.

---

## 10. Database

PostgreSQL. Migrations with `golang-migrate`. Store agent payloads as `jsonb` — the
schemas will change and rigid columns will slow that down.

```sql
CREATE TABLE users (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email        text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  role         text NOT NULL DEFAULT 'staff',
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE templates (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title         text NOT NULL,
  object_key    text NOT NULL,
  thumb_key     text,
  occasion_tags text[] NOT NULL DEFAULT '{}',
  season        text,
  style_profile jsonb,
  is_active     boolean NOT NULL DEFAULT true,
  created_by    uuid REFERENCES users(id),
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE brand_assets (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  kind        text NOT NULL,          -- logo | tagline | footer
  object_key  text NOT NULL,
  meta        jsonb NOT NULL DEFAULT '{}',
  version     int NOT NULL,
  is_active   boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE projects (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES users(id),
  title        text,
  input_text   text NOT NULL,
  template_id  uuid REFERENCES templates(id),
  state        text NOT NULL DEFAULT 'draft',
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE project_references (
  project_id  uuid REFERENCES projects(id) ON DELETE CASCADE,
  object_key  text NOT NULL,
  style_profile jsonb,
  PRIMARY KEY (project_id, object_key)
);

CREATE TABLE stage_outputs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  stage       text NOT NULL,       -- style | brief | concepts | copy | art | proof | review
  payload     jsonb NOT NULL,
  model       text,
  tokens_in   int,
  tokens_out  int,
  superseded  boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON stage_outputs (project_id, stage, created_at DESC);

CREATE TABLE renders (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  kind        text NOT NULL,       -- raw | composited
  provider    text,
  object_key  text NOT NULL,
  prompt      text,
  cost_cents  int,
  created_at  timestamptz NOT NULL DEFAULT now()
);
```

Re-running a stage inserts a new row and marks the previous one `superseded`, so the
history is auditable and nothing is destroyed.

---

## 11. Go project layout

```
cmd/server/main.go
internal/
  config/          env loading, provider keys, model names
  http/            router, handlers, middleware, SSE
  auth/            sessions, password hashing (bcrypt)
  agents/          one file per agent, shared client, schemas
    client.go      Anthropic calls, tool-use schema enforcement, retries
    style.go intake.go director.go copy.go art.go proof.go review.go
  pipeline/        orchestration, state machine, stage re-run logic
  imagegen/        ImageProvider interface + one impl per vendor
  compositor/      zone overlay, logo scaling, footer text
  templates/       library CRUD + recommendation ranking
  store/           postgres repositories, migrations
  storage/         S3/MinIO wrapper, signed URLs, thumbnails
  observability/   structured logging, request ids, usage metering
web/               React app
```

Router: `chi`. Config: env vars via `envconfig`, with a `.env.example`. No global state;
pass dependencies explicitly. Every external call takes a `context.Context` with a timeout.

---

## 12. React frontend

Vite + TypeScript + Tailwind. React Router. TanStack Query for server state. No Redux.

### Screens

**Start** — the three input affordances together: a template gallery grid, a large text
area, and a file attach control. Recommendations appear inline under the text area once
the user has typed enough, without blocking them.

**Pipeline** — the working view. A vertical run of stages, each showing its agent's output
as it streams in over SSE. Every stage has an edit control and a re-run control; re-running
visibly invalidates the stages below it. A sticky panel on the right holds the layout proof,
then the render, then the composited final, with a raw/composited toggle.

**Review** — the QA score, the fix list, and the download button. Fixes should be
one-click actionable where possible: a fix about copy jumps to the copy stage.

**Admin** — template library management with inline style-profile editing, brand asset
upload, and a usage report.

### Requirements

- Optimistic state per stage; a failed stage shows an error and a retry, never a stuck spinner.
- The prompt text must be copyable in one click, since users will want to paste it into
  other tools.
- Mobile-usable down to 380px. Staff will check posters on phones.
- Visible keyboard focus, reduced-motion respected.
- Do not use browser storage for anything the server owns.

Visual direction: this is a print production tool, not a SaaS dashboard. Avoid a grid of
identical rounded cards. Take the palette from ARMSS's own brand — navy `#16265A`, red
`#D32027`, saffron `#F5A623`, warm cream `#FFF8EE` — and let the poster previews be the
only loud thing on the page.

---

## 13. Non-functional

**Uploads.** Images only: `image/png`, `image/jpeg`, `image/webp`. 10MB per file, 4 files
per project. Validate by sniffing content, not by extension. Downscale to 1500px on the
long edge before sending to the vision agents — full-resolution print files waste tokens
and add latency for no accuracy gain.

**Secrets.** All provider API keys server-side only. The frontend must never hold one.

**Cost.** Record tokens and image spend per call in `stage_outputs` and `renders`. Enforce
a configurable monthly cap per user and a global cap; return `429` with a clear message
when hit. Image generation is the expensive part — never auto-generate more than one
image per user action.

**Rate limits.** Per-user limit on project creation and render calls.

**Idempotency.** Accept an `Idempotency-Key` header on render and composite so a retried
request does not double-charge.

**Errors.** User-facing messages say what happened and what to do. Log the provider
error with a request id; never leak raw provider payloads to the client.

**Observability.** Structured JSON logs with request id and project id. `/healthz` and
`/readyz`. Metrics for stage latency, stage failure rate, and provider spend.

**Testing.** Table-driven unit tests for the compositor geometry and the recommendation
ranking. A recorded-fixture test per agent so schema changes surface immediately —
do not hit live APIs in CI.

---

## 14. Acceptance criteria

1. A staff user picks a library template, types "Diwali 2026", and within two minutes has
   a downloadable poster whose logo, tagline and phone numbers are pixel-identical to the
   source brand files.
2. Typing a description with no template selected returns three recommended templates with
   a reason each.
3. Attaching two of the user's own posters produces a style profile that names their
   palette and their fixed furniture.
4. Editing the headline and re-running only invalidates copy, art direction, proof and
   render — not the brief or concepts.
5. Every stage's output is visible, editable and re-runnable.
6. A truncated or malformed agent response surfaces as a retry, never as a crash or a
   permanent spinner.
7. An image provider outage produces a message naming the provider, and the project stays
   resumable.
8. Admin can add a template and correct its extracted style profile by hand.
9. Admin can see token and image spend per user per day.
10. The full flow is usable on a 380px viewport.

---

## 15. Build order

1. Skeleton: Go server, Postgres, migrations, auth, S3 wrapper, health endpoints.
2. Anthropic client with tool-use schema enforcement, retries, and usage metering. Fixture tests.
3. Agents 2–5 behind the API, no UI. Verify with curl against a hardcoded brief.
4. React shell: start screen, pipeline screen, SSE wiring.
5. Template library: model, admin upload, Style Analyst, gallery in the UI.
6. Recommendations.
7. Image provider interface plus one implementation.
8. Compositor and reserved zones — the logo problem is solved here.
9. Reviewer agent and the review screen.
10. Layout proof, cost caps, admin usage report, polish.

Ship after step 8. Steps 9 and 10 are valuable but nobody is blocked without them.

---

## 16. Out of scope for v1

Multi-tenancy. Social scheduling. Video or animated posters. Print-shop integration or
CMYK conversion. A vector template store. Real-time collaborative editing. Fine-tuning
or training on ARMSS assets. Automatic translation into Tamil — worth doing later, but it
changes the typography requirements enough to be its own project.

---

## 17. Notes for the implementer

- The logo is never generated. If you find yourself writing a prompt that describes the
  ARMSS mark, stop — that path has already failed once and produced "OARMSS".
- Enforce agent output shape with tool-use schemas, not prompt instructions plus
  `json.Unmarshal`. Prompt-and-parse breaks on truncation and you will spend days on it.
- Give each agent a real token budget. 3000–4000 for JSON stages, 8000 for the SVG proof.
- Cache template style profiles aggressively. Re-analysing a library image on every
  generation is the easiest large cost to avoid.
- The image provider is the only step that costs real money per click. Gate it behind an
  explicit user action, always.
