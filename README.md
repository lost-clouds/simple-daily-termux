# simple-daily-termux — Personal Daily Management Tool

[简体中文](README_ZH.md) | [English](README.md)

A lightweight personal time management application with Go backend and vanilla JavaScript SPA frontend. Integrates **TODO list**, **Pomodoro timer**, **Diary+Ledger**, **Calendar**, and **Countdown** into a single binary. Designed to work alongside [Blog-termux](https://github.com/lost-clouds/Blog-termux) via nginx reverse proxy, adding a 9th dashboard card that shows daily summary data.

> Built for Termux on Android. Pure Go (zero CGO), single binary deployment, <20MB footprint.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Directory Structure](#directory-structure)
- [Architecture](#architecture)
- [Module Reference](#module-reference)
- [Deployment Guide](#deployment-guide)
  - [1. Requirements](#1-requirements)
  - [2. Build](#2-build)
  - [3. Configure](#3-configure)
  - [4. Run](#4-run)
  - [5. Integration with Blog-termux](#5-integration-with-blog-termux)
- [API Reference](#api-reference)
- [Usage](#usage)
- [FAQ](#faq)

---

## Quick Start

```bash
# 1. Clone the project
git clone https://github.com/lost-clouds/simple-daily-termux.git ~/simple-daily-termux
cd ~/simple-daily-termux

# 2. Build (requires Go 1.22+)
go build -o simple-daily-termux .

# 3. Create config
cp config.example.json config.json
# Edit config.json if needed (default: SQLite at ./data/daily.db, port 8090)

# 4. Run
./simple-daily-termux config.json
# Server starts at http://127.0.0.1:8090
```

Run the smoke test to verify all endpoints:

```bash
bash scripts/smoke.sh
# All 12 checks should PASS
```

---

## Directory Structure

```
simple-daily-termux/
├── main.go                          # Entry point: config → store → services → handlers → server
├── go.mod / go.sum                  # Go module (2 deps: sqlite + mysql drivers)
├── config.json / config.example.json
├── internal/
│   ├── config/config.go             # Config struct + Load() + Validate()
│   ├── idgen/idgen.go               # App-level ID generation (crypto/rand)
│   ├── httputil/response.go         # Unified JSON response envelope
│   ├── store/sqlstore/
│   │   ├── store.go                 # Store interface + all Repository implementations
│   │   ├── sqlite.go                # SQLite init (WAL mode, pure Go)
│   │   ├── mysql.go                 # MySQL init
│   │   └── migrations.go            # DDL (SQLite + MySQL dual syntax)
│   ├── todo/        {model, service, handler}.go
│   ├── countdown/   {model, service, handler}.go
│   ├── pomodoro/    {model, service, handler}.go
│   ├── diary/       {model, service, handler, ledgerparser}.go
│   ├── ledger/      {model, service, handler}.go
│   ├── calendar/    {model, service, handler}.go
│   └── summary/     {service, handler}.go
├── web/
│   ├── index.html                    # Standalone SPA (5 tabs)
│   ├── blog-termux-index.html        # Blog-termux integration index (9th dashboard card)
│   ├── css/
│   │   ├── build.sh                  # CSS build script (cat merge)
│   │   ├── style.css                 # Built output (go:embed)
│   │   └── src/                      # CSS source (modular)
│   │       ├── variables.css         #   CSS custom properties (same palette as Blog-termux)
│   │       ├── base.css / layout.css / responsive.css
│   │       ├── themes/dark.css
│   │       └── components/*.css      #   8 component stylesheets
│   ├── js/                           # ES Modules (zero bundler)
│   │   ├── main.js → app.js          #   Entry → Main controller
│   │   ├── constants.js              #   API path constants
│   │   ├── utils.js                  #   Shared utilities (esc, etc.)
│   │   ├── theme.js                  #   Theme manager
│   │   ├── calendar.js / todo.js / countdown.js
│   │   ├── pomodoro.js / diary.js / ledger.js
│   │   └── update-widget.js          #   Blog-termux integration card
│   └── lib/marked.min.js             # Markdown parser
├── scripts/
│   ├── start.sh / stop.sh            # Process management (PID file)
│   └── smoke.sh                      # curl health check
└── plan.md                           # Design plan
```

---

## Architecture

### Overall Layout

```
simple-daily-termux (Go binary)
  │
  ├── main.go ── dependency injection tree
  │   ├── config.Load("config.json")
  │   ├── sqlstore.NewSQLite()          → Store
  │   ├── ledger.NewService()           → Ledger + Settings repo
  │   ├── countdown.NewService()        → Countdown repo
  │   ├── todo.NewService(_, countSvc)  → Todo repo + countdown linkage
  │   ├── pomodoro.NewService()         → Pomodoro repo
  │   ├── diary.NewService(_, ledgerSvc)→ Diary repo + ledger linkage
  │   ├── calendar.NewService(...)      → 4 repos aggregated
  │   └── summary.NewService(...)       → Ledger + Countdown + Pomodoro
  │
  ├── http.ServeMux (Go 1.22 pattern routing)
  │   ├── /api/todos, /api/countdown, /api/pomodoro, ...
  │   ├── /api/summary                  → Blog-termux integration endpoint
  │   └── /                             → embedded SPA (go:embed)
  │
  └── graceful shutdown (SIGINT/SIGTERM → 5s timeout)
```

### Frontend Tab Structure

```
index.html (SPA)
  │
  ├─ header ─── brand title + theme toggle (🌓)
  │
  ├─ tab-bar ── [📅Calendar] [✅TODO] [⏱️Pomodoro] [📝Diary] [⏳Countdown]
  │              PC/tablet top | mobile bottom-fixed
  │
  └─ content area (5 sections, 1 visible at a time)
      ├── #sec-calendar     Month grid + events + todo deadlines + countdown targets
      ├── #sec-todo         Task list with filters, priority, deadline linkage
      ├── #sec-pomodoro     Countdown timer + today's focus minutes
      ├── #sec-diary        Markdown editor + ledger block embedding + monthly summary
      └── #sec-countdown    Countdown list (manual + todo-derived)
```

### Module Dependency Graph

```
main.go
  ├── config.Load()
  ├── sqlstore.NewSQLite() / NewMySQL()
  ├── ledger.NewService(st.Ledgers(), st.Settings())
  ├── countdown.NewService(st.Countdowns())
  ├── todo.NewService(st.Todos(), countSvc)          ← injects countdown for deadline sync
  ├── pomodoro.NewService(st.Pomodoros())
  ├── diary.NewService(st.Diaries(), ledgerSvc)       ← injects ledger for block parsing
  ├── calendar.NewService(st.Calendars(), st.Todos(), st.Countdowns(), st.Diaries())
  ├── summary.NewService(ledgerSvc, countSvc, pomoSvc, timezone)
  └── http.ServeMux (route registration + middleware)
```

**Key design**: All dependencies are explicitly injected via constructor parameters in `main.go`. No package-level mutable state. No global service locators. Go's exported/unexported visibility enforces module boundaries at compile time.

### Data Flow

```
Client (browser)
  │  GET /api/todos, POST /api/ledger, PUT /api/diary/2026-06-21, ...
  ↓
Go HTTP handlers → Service layer (business logic) → Repository interface → SQLStore
  │                                                    ↑
  │  Todo ↔ Countdown linkage                          │
  │  Diary ↔ Ledger linkage (markdown code block)      │
  │  Calendar aggregation (4 data sources)              │
  │  Summary aggregation (for Blog-termux card)         │
  ↓
JSON response: {"ok": true, "data": {...}}
```

---

## Module Reference

### todo — TODO List with Deadline Linkage

| | |
|---|---|
| Package | `internal/todo/` |
| Repository | `todo.Repository` (defined in model.go) |
| Service | `TodoService{repo, countSvc}` — injects `*countdown.Service` |
| API | `GET/POST /api/todos`, `GET/PUT/DELETE /api/todos/{id}` |

**Linkage logic**: When a todo is created/updated with a `deadline_at`, a countdown event is auto-created. When the deadline is cleared or the todo is deleted/completed, the countdown event is removed. The countdown list shows both manual and todo-derived events interleaved.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Task name |
| `status` | pending / doing / done | Current state |
| `priority` | 0-3 | Priority level |
| `deadline_at` | RFC3339 / null | Drives countdown linkage |

---

### countdown — Countdown Events

| | |
|---|---|
| Package | `internal/countdown/` |
| Repository | `countdown.Repository` |
| Service | `CountdownService{repo}` |
| API | `GET/POST /api/countdown`, `DELETE /api/countdown/{id}` |

`days_left` is computed at read time (UTC day boundaries). Supports two sources: `manual` (user-created) and `todo` (auto-derived from todo deadlines). Both sources appear in the same list.

---

### pomodoro — Focus Timer

| | |
|---|---|
| Package | `internal/pomodoro/` |
| Repository | `pomodoro.Repository` |
| Service | `PomodoroService{repo}` |
| API | `POST /api/pomodoro/start`, `POST /api/pomodoro/{id}/finish`, `GET /api/pomodoro/today`, `GET /api/pomodoro` |

Session status: `running` → `completed` / `aborted`. `actual_minutes` is computed from `ended_at - started_at`. Today's focus time respects the configured timezone.

| Field | Description |
|-------|-------------|
| `planned_minutes` | Target duration (default 25) |
| `actual_minutes` | Real elapsed time |
| `linked_todo_id` | Optional task association |

---

### diary — Markdown Journal

| | |
|---|---|
| Package | `internal/diary/` |
| Repository | `diary.Repository` |
| Service | `DiaryService{repo, ledgerSvc}` — injects `*ledger.Service` |
| API | `GET /api/diary/{date}`, `PUT /api/diary/{date}`, `GET /api/diary?month=YYYY-MM` |

**Ledger linkage**: On save, scans `content_md` for `` ```ledger `` code blocks, deletes old diary-sourced ledger entries, and inserts new ones. The process is idempotent — re-saving the same diary doesn't duplicate ledger entries. If all ledger blocks are removed, associated entries are cleaned up.

````markdown
```ledger
type: expense
amount: 35.5
category: 餐饮
note: 午饭
```
````

Markdown rendering uses `marked.js` (with a pure-JS fallback). Ledger blocks are rendered as styled transaction cards instead of code blocks.

---

### ledger — Personal Accounting

| | |
|---|---|
| Package | `internal/ledger/` |
| Repository | `ledger.Repository` + `SettingsRepository` |
| Service | `LedgerService{repo, settings}` |
| API | `GET/POST /api/ledger?month=`, `GET /api/ledger/summary?month=`, `DELETE /api/ledger/{id}` |

**Money stored as integer cents** to avoid floating-point errors.

| Metric | Formula |
|--------|---------|
| Monthly expense | `SUM(amount_cents) WHERE type='expense'` |
| Monthly income | `SUM(amount_cents) WHERE type='income'` |
| Monthly balance | income − expense |
| Savings | `opening_savings_cents` + cumulative sum of all historical monthly balances |

`opening_savings_cents` is set once via the settings KV store, then savings roll forward automatically.

---

### calendar — Aggregated Month View

| | |
|---|---|
| Package | `internal/calendar/` |
| Repository | `calendar.Repository` + 3 others (todo, countdown, diary) |
| Service | `CalendarService{calRepo, todoRepo, countRepo, diaryRepo}` |
| API | `GET /api/calendar?month=YYYY-MM`, `POST/PUT/DELETE /api/calendar/events`, `GET/PUT /api/calendar/events/{id}` |

Aggregates 4 data sources in a single response:
- **Calendar events** — standalone scheduled events
- **Todo deadlines** — todos with `deadline_at` in the target month
- **Countdown targets** — countdown events targeting the month
- **Diary dates** — days with diary entries (for calendar grid dots)

---

### summary — Blog-termux Integration Card

| | |
|---|---|
| Package | `internal/summary/` |
| Service | `SummaryService{ledgerSvc, countSvc, pomoSvc, timezone}` |
| API | `GET /api/summary` |

Lightweight endpoint designed for the 9th dashboard card in Blog-termux:

```json
{
  "calendar": {"month": "2026-06", "today": "2026-06-21"},
  "ledger": {"expense": 1234.50, "income": 6000.00, "balance": 4765.50, "savings": 88234.00},
  "countdown": [{"title": "Project deadline", "days_left": 193}],
  "focus_today_minutes": 95
}
```

---

## Deployment Guide

### 1. Requirements

| Component | Purpose | Notes |
|-----------|---------|-------|
| Go 1.22+ | Build | Required for `http.ServeMux` pattern routing |
| SQLite | Default database | Built-in via `modernc.org/sqlite` (pure Go, no CGO) |
| MySQL (optional) | Alternative database | Only when `config.json` `driver=mysql` |
| Nginx (optional) | Reverse proxy | Only for Blog-termux integration |

> **NOT required**: Node.js, Python, PHP, Docker, C/C++ toolchain, GPU.

### 2. Build

```bash
cd ~/simple-daily-termux

# Download Go dependencies (requires network)
GOPROXY=https://goproxy.cn,direct go mod tidy

# Build
go build -o simple-daily-termux .

# Verify
./simple-daily-termux --help  # (accepts config path as argument)
```

Cross-compilation (no CGO):

```bash
GOOS=linux GOARCH=arm64 go build -o simple-daily-termux .
GOOS=linux GOARCH=amd64 go build -o simple-daily-termux .
```

### 3. Configure

Copy and edit `config.json`:

```json
{
  "server": {
    "addr": "127.0.0.1:8090",
    "cors": false
  },
  "database": {
    "driver": "sqlite",
    "sqlite": { "path": "./data/daily.db" },
    "timezone": "Local"
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `server.addr` | `127.0.0.1:8090` | Listen address |
| `server.cors` | `false` | Enable CORS headers (for dev) |
| `database.driver` | `sqlite` | `sqlite` or `mysql` |
| `database.sqlite.path` | `./data/daily.db` | SQLite database file |
| `database.mysql.dsn` | — | MySQL DSN (when `driver=mysql`) |
| `database.timezone` | `Local` | Timezone for pomodoro "today" calculations |

### 4. Run

**Manual start:**

```bash
./simple-daily-termux config.json
# Listening on 127.0.0.1:8090
```

**Daemon (via scripts):**

```bash
bash scripts/start.sh
bash scripts/stop.sh
```

**Smoke test:**

```bash
bash scripts/smoke.sh
# All 12 checks should PASS
```

Add to crontab for auto-restart:

```bash
# crontab -e
# */5 * * * * cd ~/simple-daily-termux && bash scripts/start.sh
```

### 5. Integration with Blog-termux

**Step 1 — Add nginx proxy rules**

Add to Blog-termux's nginx config:

```nginx
location /update/ {
    proxy_pass http://127.0.0.1:8090/;
    proxy_set_header Host $host;
}
location /api/summary {
    proxy_pass http://127.0.0.1:8090/api/summary;
    add_header Cache-Control "no-store";
}
```

**Step 2 — Deploy the rewritten index.html**

```bash
cp web/blog-termux-index.html /path/to/Blog-termux/index.html
cp web/js/update-widget.js   /path/to/Blog-termux/js/update-widget.js
```

**Step 3 — Reload nginx**

```bash
nginx -s reload
```

The Blog-termux dashboard now shows a 9th card ("Daily Overview") with live data from simple-daily-termux. Clicking the card navigates to the full SPA.

---

## API Reference

All responses use the unified envelope format: `{"ok": true, "data": ...}` or `{"ok": false, "error": "..."}`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/todos` | List todos (`?status=&has_deadline=`) |
| POST | `/api/todos` | Create todo |
| GET | `/api/todos/{id}` | Get todo by ID |
| PUT | `/api/todos/{id}` | Update todo |
| DELETE | `/api/todos/{id}` | Delete todo |
| GET | `/api/countdown` | List all countdown events |
| POST | `/api/countdown` | Create manual countdown |
| DELETE | `/api/countdown/{id}` | Delete countdown |
| POST | `/api/pomodoro/start` | Start pomodoro session |
| POST | `/api/pomodoro/{id}/finish` | Finish session (`{"status":"completed\|aborted"}`) |
| GET | `/api/pomodoro/today` | Today's total focus minutes |
| GET | `/api/pomodoro` | List sessions (`?from=&to=`) |
| GET | `/api/diary/{date}` | Get diary entry (`date=YYYY-MM-DD`) |
| PUT | `/api/diary/{date}` | Save diary entry (triggers ledger sync) |
| GET | `/api/diary` | List month entries (`?month=YYYY-MM`) |
| GET | `/api/ledger` | List ledger entries (`?month=YYYY-MM`) |
| GET | `/api/ledger/summary` | Monthly summary (`?month=YYYY-MM`) |
| POST | `/api/ledger` | Create ledger entry |
| DELETE | `/api/ledger/{id}` | Delete ledger entry |
| GET | `/api/calendar` | Aggregated month view (`?month=YYYY-MM`) |
| POST | `/api/calendar/events` | Create calendar event |
| PUT | `/api/calendar/events/{id}` | Update calendar event |
| DELETE | `/api/calendar/events/{id}` | Delete calendar event |
| GET | `/api/summary` | Integration card data |
| GET | `/api/health` | Health check |

---

## Usage

| Action | How |
|--------|-----|
| **Switch tab** | PC/tablet: click top tab bar. Mobile: tap bottom nav bar |
| **Dark mode** | Click 🌓 button, preference auto-saved |
| **Create todo** | TODO tab → "+ New Task" → fill form → Save |
| **Set deadline** | Add deadline in todo form → auto-appears in Countdown |
| **Start pomodoro** | Pomodoro tab → choose duration → Start |
| **Write diary** | Diary tab → pick date → write Markdown → "Insert Ledger" for accounting |
| **View ledger** | Diary tab → monthly summary + entry list at bottom |
| **Browse calendar** | Calendar tab → month grid with diary dots + events + deadlines |
| **Manage countdowns** | Countdown tab → manual entries + auto-generated from todos |
| **Set opening savings** | Ledger → need to set `opening_savings_cents` via settings (API) |

---

## FAQ

### Q: How to switch from SQLite to MySQL?

Edit `config.json`:

```json
{
  "database": {
    "driver": "mysql",
    "mysql": { "dsn": "user:pass@tcp(127.0.0.1:3306)/daily?parseTime=true" }
  }
}
```

Restart the server. Tables are auto-created on startup. Existing SQLite data must be migrated manually.

### Q: Where is the database file?

Default: `./data/daily.db`. Change via `config.json` → `database.sqlite.path`.

### Q: How to set the opening savings balance?

The `opening_savings_cents` setting is stored in the `settings` table. Set it via the API:

```bash
# Set opening savings to ¥50,000.00 (5,000,000 cents)
# This requires a settings endpoint — currently set via direct DB or future settings UI
```

### Q: Pomodoro timer inaccurate after switching tabs?

The frontend timer uses `setInterval` which browsers throttle in background tabs. The timer pauses automatically when switching away from the Pomodoro tab. If you need exact timing, use the server-side `started_at`/`ended_at` values.

### Q: Multiple ledger entries after re-saving a diary?

No. The diary save logic is idempotent — old diary-sourced entries are deleted before new ones are inserted.

### Q: Calendar shows empty?

Make sure you have data in at least one of: calendar events, todo deadlines, countdown events, or diary entries. The calendar module aggregates from all four sources.

### Q: Blog-termux dashboard card shows "--"?

Check:
1. Is simple-daily-termux running? `curl http://127.0.0.1:8090/api/health`
2. Is nginx proxy configured? `curl http://127.0.0.1:7443/api/summary`
3. Is `update-widget.js` deployed to Blog-termux?

---

## Technical Highlights

| Feature | Implementation |
|---------|---------------|
| Zero framework | Go `net/http` standard library + `http.ServeMux` pattern routing |
| Pure Go SQLite | `modernc.org/sqlite` — zero CGO, cross-compile anywhere |
| Single binary | Frontend embedded via `go:embed`, single file deployment |
| Modular architecture | handler → service → repository, each package self-contained |
| Variable isolation | Go: zero package-level mutable state. JS: ES Module scope + `tu-` CSS prefix |
| Money precision | Integer cents (int64), `math.Round(f * 100)` conversion |
| Idempotent sync | Diary-ledger: delete old + insert new on every save |
| Unified envelope | `{"ok": true/false, "data": ...}` across all endpoints |
| Theme | CSS variables + `body.dark` toggle (same palette as Blog-termux) |
| Responsive | 4 breakpoints (1024/640/400px), bottom nav on mobile |
| Square corners | `border-radius: 0` — visual differentiation from Blog-termux |
| Markdown rendering | marked.js with ledger code block extraction |
| No build tools | ES Modules (zero bundler), CSS via `cat` merge |
| Graceful shutdown | SIGINT/SIGTERM → 5s context timeout → server.Shutdown |
| Timezone-aware | Configurable timezone for pomodoro today calculations |

---

## Links

[Blog-termux](https://github.com/lost-clouds/Blog-termux) — the companion dashboard project
