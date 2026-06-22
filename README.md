# simple-daily-termux — Personal Daily Management Tool

[简体中文](README_ZH.md) | [English](README.md)

[![CI](https://github.com/lost-clouds/simple-daily-termux/actions/workflows/ci.yml/badge.svg)](https://github.com/lost-clouds/simple-daily-termux/actions/workflows/ci.yml)

A lightweight personal time management application — Go backend + vanilla JavaScript SPA frontend, compiled into a single binary. Provides **TODO list**, **Pomodoro timer**, **Diary + Ledger**, **Calendar aggregation**, and **Countdown** modules. Integrates with [Blog-termux](https://github.com/lost-clouds/Blog-termux) via nginx reverse proxy, adding a daily summary card to the dashboard.

> Built for Termux on Android. Pure Go (zero CGO), single binary deployment, ~15MB footprint.

![Screenshot](example/example.png)

---

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Module Reference](#module-reference)
- [Deployment Guide](#deployment-guide)
  - [Standalone Deployment](#standalone-deployment)
  - [Integration with Blog-termux](#integration-with-blog-termux)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Usage](#usage)
- [Development](#development)
- [FAQ](#faq)

---

## Quick Start

**Option A — Download pre-built binary (recommended):**

```bash
# linux-amd64 / linux-arm64 / linux-armv7 available
curl -sSLO https://github.com/lost-clouds/simple-daily-termux/releases/latest/download/simple-daily-termux-linux-arm64.tar.gz
tar -xzf simple-daily-termux-linux-arm64.tar.gz
```

**Option B — Build from source (Go 1.22+):**

```bash
git clone https://github.com/lost-clouds/simple-daily-termux.git
cd simple-daily-termux
bash web/css/build.sh    # Build CSS
go build -o simple-daily-termux .
```

**Run:**

```bash
cp config.example.json config.json
# Edit config.json if needed
./simple-daily-termux config.json
# → http://127.0.0.1:8090
```

**Verify:**

```bash
bash scripts/smoke.sh          # 12 API endpoint checks
# All PASS → ready to use
```

> Current release: [v0.0.2](https://github.com/lost-clouds/simple-daily-termux/releases/tag/v0.0.2)

---

## Architecture

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| HTTP | `net/http` (zero framework, `http.ServeMux` pattern routing) |
| Database | SQLite (`modernc.org/sqlite`, pure Go, zero CGO) / MySQL optional |
| Frontend | Vanilla ES Modules (no bundler), CSS custom properties (cat merge), `marked.js` |
| Assets | `go:embed` — frontend compiled into binary |
| Process | PID file + `start.sh` / `stop.sh` (no systemd required) |

### Directory Structure

```
simple-daily-termux/
├── main.go                     # Entry point — DI tree, route registration, graceful shutdown
├── config.json / config.example.json
├── internal/
│   ├── config/config.go        # Config loading + validation
│   ├── idgen/idgen.go          # crypto/rand ID generation
│   ├── httputil/response.go    # Unified JSON envelope
│   ├── store/sqlstore/         # SQLite/MySQL implementation (Store interface + migrations)
│   ├── todo/       {model, service, handler}.go
│   ├── countdown/  {model, service, handler}.go
│   ├── pomodoro/   {model, service, handler}.go
│   ├── diary/      {model, service, handler, ledgerparser}.go
│   ├── ledger/     {model, service, handler}.go
│   ├── calendar/   {model, service, handler}.go
│   └── summary/    {service, handler}.go
├── web/                         # go:embed → compiled into binary
│   ├── index.html               # SPA (6 tabs: home, calendar, todo, pomodoro, diary, countdown)
│   ├── blog-termux-index.html   # Blog-termux integration HTML (9th dashboard card)
│   ├── css/  src/*.css + build.sh + style.css
│   ├── js/   app.js + 7 modules + theme.js + utils.js
│   └── lib/  marked.min.js
├── example/                     # Nginx config templates
│   ├── standalone.conf          # Independent deployment
│   └── integration.conf         # Blog-termux integration snippet
├── scripts/
│   ├── start.sh / stop.sh       # Process management
│   └── smoke.sh                 # curl health check
└── .github/workflows/
    ├── ci.yml                   # Build + smoke test on push/PR
    └── release.yml              # Cross-compile + GitHub Release on tag
```

### Dependency Graph

```
main.go
  ├── config.Load()
  ├── sqlstore.NewSQLite() / NewMySQL()          → Store
  ├── ledger.NewService(st.Ledgers(), st.Settings())
  ├── countdown.NewService(st.Countdowns())
  ├── todo.NewService(st.Todos(), countSvc)       ← injects countdown for deadline sync
  ├── pomodoro.NewService(st.Pomodoros())
  ├── diary.NewService(st.Diaries(), ledgerSvc)   ← injects ledger for code block parsing
  ├── calendar.NewService(st.Calendars(), st.Todos(), st.Countdowns(), st.Diaries())
  ├── summary.NewService(ledgerSvc, countSvc, pomoSvc, timezone)
  └── http.ServeMux → all handlers register their own routes
```

---

## Module Reference

### Home Page

Three-panel layout: left column (today's TODOs + diary preview), right column (calendar grid with countdown event previews). Clicking a calendar date loads that day's TODOs and diary. Navigation bar under calendar.

- Today's focus/rest minutes shown below TODO panel
- Today's income/expense shown below diary panel
- Calendar cells display countdown event titles (truncated with `…`)
- Daily/long-term tasks auto-instantiated via `POST /api/todos/ensure-daily`

### TODO

| Field | Description |
|-------|-------------|
| `task_type` | `daily` (auto-create daily) / `long_term` (auto-create until deadline) / `one_time` (manual) |
| `priority` | Auto-calculated: I (red, <14d) / II (orange, 14-30d) / III (yellow, 30-60d) / IV (blue, ≥60d) |
| `deadline_at` | Drives countdown linkage — setting/clearing auto-syncs countdown events |

### Pomodoro

- Large countdown timer (5rem desktop, responsive scaling)
- Start focus session → countdown → completion overlay (frosted blue card) → click to start rest timer
- Separate focus/rest time accumulation
- Background timer continues when switching tabs (no session abort)

### Diary + Ledger

- PC: left-right layout (Markdown editor | rendered preview)
- Markdown via `marked.js` with custom ledger block rendering
- Ledger blocks in diary (` ```ledger ... ``` `) auto-sync to ledger on save (idempotent)
- Standalone ledger entry: category dropdown (salary/dining/transport/shopping/utilities/side-income/other)
- Monthly summary: expense, income, balance, cumulative savings

### Calendar

- Month grid with countdown events and todo deadlines displayed per cell
- Diary dates marked with blue dots
- Full-page calendar view with events list

### Countdown

- Manual countdown events + auto-derived from todo deadlines
- `days_left` computed at read time
- Urgent styling (red) for ≤7 days

---

## Deployment Guide

### Standalone Deployment

**1. Build:**

```bash
git clone https://github.com/lost-clouds/simple-daily-termux.git
cd simple-daily-termux
bash web/css/build.sh
go build -o simple-daily-termux .
```

**2. Configure:**

```bash
cp config.example.json config.json
```

Default: SQLite at `./data/daily.db`, listen `127.0.0.1:8090`.

**3. Run:**

```bash
# Manual
./simple-daily-termux config.json

# Daemon
bash scripts/start.sh
bash scripts/stop.sh
```

**4. Smoke test:**

```bash
bash scripts/smoke.sh
```

**5. (Optional) Reverse proxy with nginx:**

See [example/standalone.conf](example/standalone.conf).

### Integration with Blog-termux

**Step 1 — Deploy simple-daily-termux** (see standalone deployment above).

**Step 2 — Add nginx proxy rules:**

Copy the content of [example/integration.conf](example/integration.conf) into Blog-termux's nginx `server {}` block. This adds two locations:

- `/simpledaily/` → proxies to the Go SPA
- `/api/summary` → data endpoint for the dashboard card

**Step 3 — Deploy the integration file:**

```bash
cp web/blog-termux-index.html /path/to/Blog-termux/index.html
```
The summary card widget is inlined — no separate JS file to deploy.

**Step 4 — Reload nginx:**

```bash
nginx -s reload
```

The Blog-termux dashboard now shows a "Daily Overview" card as the first card in the grid. Clicking it opens the full SPA in a new tab.

---

## API Reference

All responses: `{"ok": true, "data": ...}` or `{"ok": false, "error": "..."}`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/todos` | List (`?status=&task_type=&entry_date=`) |
| POST | `/api/todos` | Create |
| GET | `/api/todos/{id}` | Get by ID |
| PUT | `/api/todos/{id}` | Update |
| DELETE | `/api/todos/{id}` | Delete |
| POST | `/api/todos/ensure-daily` | Auto-instantiate daily/long-term tasks for a date |
| GET | `/api/countdown` | List all |
| POST | `/api/countdown` | Create manual |
| DELETE | `/api/countdown/{id}` | Delete |
| POST | `/api/pomodoro/start` | Start focus session |
| POST | `/api/pomodoro/start-rest` | Start rest session |
| POST | `/api/pomodoro/{id}/finish` | Finish (`{"status":"completed\|aborted"}`) |
| GET | `/api/pomodoro/today` | Today's focus + rest minutes |
| GET | `/api/diary/{date}` | Get entry (date=YYYY-MM-DD) |
| PUT | `/api/diary/{date}` | Save (triggers ledger sync) |
| GET | `/api/diary` | List month (`?month=YYYY-MM`) |
| GET | `/api/ledger` | List (`?month=YYYY-MM`) |
| GET | `/api/ledger/{id}` | Get by ID |
| GET | `/api/ledger/summary` | Monthly summary (`?month=YYYY-MM`) |
| POST | `/api/ledger` | Create entry |
| DELETE | `/api/ledger/{id}` | Delete |
| PUT | `/api/settings/{key}` | Set KV (`{"value":"..."}`) |
| GET | `/api/calendar` | Aggregated month view (`?month=YYYY-MM`) |
| POST | `/api/calendar/events` | Create event |
| PUT | `/api/calendar/events/{id}` | Update event |
| DELETE | `/api/calendar/events/{id}` | Delete event |
| GET | `/api/summary` | Integration card data |
| GET | `/api/health` | Health check |

---

## Configuration

```json
{
  "server": {
    "addr": "127.0.0.1:8090",
    "cors": false
  },
  "database": {
    "driver": "sqlite",
    "sqlite": { "path": "./data/daily.db" },
    "mysql": { "dsn": "user:pass@tcp(host:3306)/daily?parseTime=true" },
    "timezone": "Local"
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `server.addr` | `127.0.0.1:8090` | Listen address |
| `server.cors` | `false` | Enable CORS headers (dev only) |
| `database.driver` | `sqlite` | `sqlite` or `mysql` |
| `database.sqlite.path` | `./data/daily.db` | SQLite file path |
| `database.timezone` | `Local` | Timezone for pomodoro "today" calculations |

---

## Usage

| Action | How |
|--------|-----|
| **Home** | Calendar + today's TODOs + diary preview. Click calendar date → loads that day |
| **Create daily task** | TODO tab → type "daily" → auto-appears every day |
| **Create long-term task** | TODO tab → type "long_term" + deadline → daily instances until deadline |
| **Start pomodoro** | Pomodoro tab → choose duration → Start. Completion overlay → click to start rest |
| **Write diary** | Diary tab → pick date → write Markdown. Ledger blocks auto-sync |
| **Add ledger entry** | Diary tab → "+ 记账" → select category from dropdown → enter amount → Save |
| **Set opening savings** | `curl -X PUT http://127.0.0.1:8090/api/settings/opening_savings_cents -d '{"value":"5000000"}'` |
| **Dark mode** | Click 🌓 button |

---

## Development

```bash
# Build CSS (after editing src/*.css)
bash web/css/build.sh

# Build binary
go build -o simple-daily-termux .

# Cross-compile (zero CGO)
GOOS=linux GOARCH=arm64 go build -o simple-daily-termux .

# Run tests
bash scripts/smoke.sh

# CI (on push/PR)
# See .github/workflows/ci.yml
```

### Design Principles

- **Zero package-level mutable state** in Go — all deps injected via constructors, exported/unexported visibility enforces boundaries
- **ES Module scope isolation** — every JS file is a separate module, no global namespace pollution
- **CSS tu- prefix** — all custom classes prefixed to avoid conflicts when embedded in Blog-termux
- **Money in integer cents** — `amount_cents int64`, converted on the frontend boundary

---

## FAQ

**Q: How to set the opening savings balance?**

```bash
curl -X PUT http://127.0.0.1:8090/api/settings/opening_savings_cents \
  -H 'Content-Type: application/json' \
  -d '{"value":"5000000"}'
# 5,000,000 cents = ¥50,000.00
```

**Q: Daily tasks don't appear?**

Daily/long-term tasks are instantiated when the home page loads (via `POST /api/todos/ensure-daily`). If they don't appear, check the browser console for errors.

**Q: Switch from SQLite to MySQL?**

Edit `config.json`: set `driver` to `mysql` and provide `mysql.dsn`. Restart. Tables are auto-created. Existing SQLite data must be migrated manually.

**Q: Blog-termux dashboard card shows "--"?**

1. Is simple-daily-termux running? `curl http://127.0.0.1:8090/api/health`
2. Are the nginx proxy rules added? `curl http://127.0.0.1:7443/api/summary`
3. Is `blog-termux-index.html` deployed?

**Q: CSS/JS not loading through nginx?**

The SPA uses relative paths. Ensure nginx has `location /simpledaily/ { proxy_pass http://127.0.0.1:8090/; }`.

---

## Links

- [Blog-termux](https://github.com/lost-clouds/Blog-termux) — companion dashboard
- [Linux.do](https://linux.do)
