# simple-daily-termux 实现计划

## Context

在 Blog-termux（纯静态 SPA，无后端 runtime）基础上新建 Go 后端 + WebUI 项目，提供 TODO 列表、番茄钟、日记记账、日历、倒计时等个人时间管理功能。项目交付后通过重写 Blog-termux 的 `index.html`（新增第 9 张联动卡片）和 nginx 反代实现同源集成。

设计硬约束：轻量低依赖、全模块化、变量不污染、三端适配、与 Blog-termux 风格统一但直角化。

---

## 1. 技术选型

### 后端
| 项 | 选型 | 理由 |
|---|---|---|
| 语言 | Go 1.22+ | `http.ServeMux` 原生支持 `"GET /todos/{id}"` 路由，省掉第三方路由库 |
| HTTP | `net/http` 标准库 | 零框架 |
| SQLite | `modernc.org/sqlite` | 纯 Go，无 CGO，Termux 下可直接编译 |
| MySQL | `github.com/go-sql-driver/mysql` | 纯 Go，无 CGO |
| CouchDB | `net/http` + `encoding/json` | CouchDB 本身就是 HTTP+JSON，零额外依赖 |
| 配置 | `config.json` + `encoding/json` | 与 Blog-termux 风格一致 |
| 静态资源 | `embed` 标准库 | 前端打包进二进制，交付物=1 个可执行文件 |
| ID 生成 | `crypto/rand` + 时间戳 | 手写 `internal/idgen`，不引入 uuid 库 |

**第三方依赖：最多 2 个**（sqlite 驱动 + mysql 驱动，CouchDB 零依赖）。

### 前端
与 Blog-termux 完全同路线：原生 ES Modules、CSS 自定义属性、无框架、`cat` 合并 CSS。

**直接复用** Blog-termux 以下文件（复制，不改动）：
- `theme.js` — 主题切换
- `utils.js` — escapeHtml / getSafeUrl / parseAutoindex
- `lightbox.js` — 图片灯箱
- `md-viewer.js` + `sanitizer.js` + `footnotes.js` — Markdown 渲染（扩展 ledger 代码块渲染）
- `lib/marked.min.js` — Markdown 解析

### 配色复用

与 Blog-termux 同名同值的 CSS 变量（已验证 [variables.css](css/src/variables.css) 和 [dark.css](css/src/themes/dark.css)）：

**浅色主题 (`:root`)：**
- `--accent: #007aff`, `--accent-hover: #005fc1`, `--accent-soft: rgba(0,122,255,0.1)`
- `--bg-primary: #e8e8ed`, `--bg-secondary: #f5f5f7`
- `--text-primary: #1c1c1e`, `--text-secondary: #636366`, `--text-tertiary: #aeaeb2`
- `--glass-bg: rgba(255,255,255,0.72)`, `--glass-bg-strong: rgba(255,255,255,0.88)`
- `--glass-border: rgba(0,0,0,0.06)`, `--glass-border-strong: rgba(0,0,0,0.10)`
- `--shadow-sm/md/lg`, `--radius-sm: 6px / md: 12px / lg: 18px / xl: 24px / full: 980px`

**深色主题 (`body.dark`)：**
- `--bg-primary: #0d0d0f`, `--bg-secondary: #1c1c1e`
- `--text-primary: #f5f5f7`, `--text-secondary: #aeaeb2`, `--text-tertiary: #636366`
- `--glass-bg: rgba(28,28,30,0.78)`, `--glass-border: rgba(255,255,255,0.08)`
- 背景：`radial-gradient(circle at 10% 20%, #1c1c1e, #0d0d0f)`

**唯一差异：** 所有 `border-radius` 变量统一改为 `0`（直角风格），即 `--radius-sm: 0; --radius-md: 0; --radius-lg: 0; --radius-xl: 0; --radius-full: 0;`

### 响应式断点（与 Blog-termux 完全一致）

| 断点 | 布局 |
|---|---|
| `>= 1024px` | 桌面：dashboard 4 列 |
| `640px - 1023px` | 平板：dashboard 2 列 |
| `< 640px` | 手机：单列，底部导航替代顶部 Tab |
| `< 400px` | 小屏：更紧凑的内边距 |

---

## 2. 目录结构

```
simple-daily-termux/
├── go.mod
├── main.go                          # 入口：config → store → services → handlers → server
├── config.example.json
├── internal/
│   ├── config/config.go             # Config struct + Load() + Validate()
│   ├── idgen/idgen.go               # NewID() = crypto/rand hex + unixnano hex
│   ├── httputil/response.go         # JSON() / Error() 统一响应封装
│   ├── store/
│   │   ├── store.go                 # 所有 Repository 接口 + Store 聚合 + New() 工厂
│   │   ├── sqlstore/
│   │   │   ├── sqlite.go            # NewSQLite(path)
│   │   │   ├── mysql.go             # NewMySQL(dsn)
│   │   │   ├── store.go             # SQL 通用实现（database/sql）
│   │   │   └── migrations.go        # DDL CREATE TABLE 语句 + 建索引
│   │   └── couchstore/
│   │       └── store.go             # CouchDB REST 实现
│   ├── todo/        {model, service, handler}.go
│   ├── countdown/   {model, service, handler}.go
│   ├── pomodoro/    {model, service, handler}.go
│   ├── diary/       {model, service, handler, ledgerparser}.go
│   ├── ledger/      {model, service, handler}.go
│   ├── calendar/    {model, service, handler}.go
│   └── summary/     {service, handler}.go
├── web/                              # go:embed 打包进二进制
│   ├── index.html                    # 独立的 full SPA（5 tab）
│   ├── blog-termux-index.html        # 重写的 Blog-termux index.html（含第9卡片）
│   ├── css/
│   │   ├── build.sh                  # cat 合并脚本
│   │   ├── style.css                 # 合并产物
│   │   └── src/
│   │       ├── _header.css
│   │       ├── variables.css         # 复用 Blog-termux 变量 + 全部 radius=0
│   │       ├── base.css
│   │       ├── layout.css
│   │       ├── components/
│   │       │   ├── header.css
│   │       │   ├── tabs.css
│   │       │   ├── calendar.css
│   │       │   ├── todo.css
│   │       │   ├── countdown.css
│   │       │   ├── pomodoro.css
│   │       │   ├── diary.css
│   │       │   ├── ledger.css
│   │       │   ├── summary-card.css  # Blog-termux 集成卡片样式
│   │       │   └── bottom-nav.css
│   │       ├── themes/dark.css
│   │       └── responsive.css
│   ├── js/
│   │   ├── main.js                   # 单行 import './app.js'
│   │   ├── app.js                    # 主控：初始化→Tab 切换→可见性管理
│   │   ├── constants.js              # API / LIBS 路径常量
│   │   ├── calendar.js
│   │   ├── todo.js
│   │   ├── countdown.js
│   │   ├── pomodoro.js
│   │   ├── diary.js
│   │   ├── ledger.js
│   │   ├── update-widget.js          # Blog-termux 集成卡片 JS
│   │   └── pure/                     # 纯逻辑函数，无 DOM 依赖，可 node --test
│   │       ├── date-utils.js
│   │       ├── ledger-parse.js
│   │       └── countdown-calc.js
│   └── lib/                          # 复用 Blog-termux 的第三方库
│       └── marked.min.js
├── scripts/
│   ├── start.sh / stop.sh / restart.sh
│   └── smoke.sh                      # curl 冒烟测试
└── README.md
```

---

## 3. 数据模型

所有 ID 用字符串（应用层生成），金额用「分」存整数。

```sql
-- todos
id TEXT PK, title TEXT NOT NULL, notes TEXT DEFAULT '',
status TEXT NOT NULL DEFAULT 'pending',  -- pending | doing | done
priority INTEGER NOT NULL DEFAULT 0,     -- 0-3
deadline_at TEXT,                        -- ISO 8601, nullable, 驱动倒计时联动
created_at TEXT NOT NULL, updated_at TEXT NOT NULL, completed_at TEXT

-- countdown_events
id TEXT PK, title TEXT NOT NULL, target_at TEXT NOT NULL,
source TEXT NOT NULL DEFAULT 'manual',   -- manual | todo
ref_id TEXT,                             -- source=todo 时指向 todos.id
note TEXT DEFAULT '', created_at TEXT NOT NULL

-- pomodoro_sessions
id TEXT PK, started_at TEXT NOT NULL, ended_at TEXT,
planned_minutes INTEGER NOT NULL DEFAULT 25,
actual_minutes INTEGER NOT NULL DEFAULT 0,
status TEXT NOT NULL DEFAULT 'completed', -- completed | aborted
linked_todo_id TEXT

-- diary_entries
id TEXT PK, entry_date TEXT NOT NULL UNIQUE,  -- YYYY-MM-DD, 一天一篇
content_md TEXT NOT NULL DEFAULT '',           -- 含内嵌 ```ledger 代码块
mood TEXT DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL

-- ledger_entries
id TEXT PK, entry_date TEXT NOT NULL,          -- YYYY-MM-DD
type TEXT NOT NULL,                             -- income | expense
amount_cents INTEGER NOT NULL,
category TEXT NOT NULL, note TEXT DEFAULT '',
source TEXT NOT NULL DEFAULT 'manual',          -- manual | diary
source_diary_id TEXT, created_at TEXT NOT NULL

-- calendar_events
id TEXT PK, title TEXT NOT NULL,
start_at TEXT NOT NULL, end_at TEXT,
all_day INTEGER NOT NULL DEFAULT 0,
note TEXT DEFAULT '', created_at TEXT NOT NULL

-- settings (KV)
key TEXT PK, value TEXT NOT NULL   -- e.g. opening_savings_cents
```

---

## 4. 存储抽象层

```go
// internal/store/store.go
type TodoRepository interface {
    Create(ctx, *todo.Todo) error
    Get(ctx, id string) (*todo.Todo, error)
    Update(ctx, *todo.Todo) error
    Delete(ctx, id string) error
    List(ctx, filter todo.ListFilter) ([]*todo.Todo, error)
}
// LedgerRepository / DiaryRepository / CountdownRepository / PomodoroRepository / CalendarRepository / SettingsRepository 同理

type Store interface {
    Todos() TodoRepository
    Ledger() LedgerRepository
    Diary() DiaryRepository
    Countdown() CountdownRepository
    Pomodoro() PomodoroRepository
    Calendar() CalendarRepository
    Settings() SettingsRepository
    Close() error
}

func New(cfg config.Database) (Store, error)  // driver="sqlite"|"mysql"|"couchdb"
```

`main.go` 拿到的永远是 `Store` 接口，业务代码完全不知道底层是哪种数据库。

---

## 5. 业务模块设计

### 5.1 TODO ↔ 倒计时联动

`todos.deadline_at` 是唯一数据源。`countdown_events` 中 `source='todo'` 的记录是**派生数据**，由 `TodoService` 在应用层维护：

- 创建/更新 todo 设置 deadline → `countSvc.SyncFromTodo()` upsert 倒计时记录
- deadline 清空 或 todo 完成/删除 → `countSvc.RemoveByRef()` 删除倒计时记录
- 倒计时列表同时包含手动创建的 (`source='manual'`) 和 todo 派生的，UI 混排

### 5.2 日记 ↔ 记账联动

日记编辑器有「插入记账」按钮，插入模板：

````markdown
```ledger
type: expense
amount: 35.5
category: 餐饮
note: 午饭
```
````

保存日记时，`DiaryService.Save()` 扫描 `content_md` 中所有 `ledger` 代码块：
1. 先删除该日记下所有 `source='diary'` 的旧记录
2. 重新解析并插入（**幂等**：重复保存不会越存越多）

前端 `md-viewer.js` 扩展：识别 `ledger` 代码块不渲染为代码框，而是渲染成交易卡片（金额+分类+备注）。

### 5.3 记账卡片

| 字段 | 计算 |
|---|---|
| 当月已花费 | `SUM(amount_cents) WHERE type='expense' AND entry_date IN 当月` |
| 入账 | `SUM(amount_cents) WHERE type='income' AND entry_date IN 当月` |
| 月结余 | 入账 - 已花费 |
| 储蓄 | `settings.opening_savings_cents` + 历史所有月结余累计 |

用户只需录入一次 `opening_savings_cents`，后续全自动滚动。

### 5.4 番茄钟

`pomodoro_sessions` 记录 start/end，可选关联 todo。「今日专注时间」= `SUM(actual_minutes) WHERE date(started_at)=today`，供 summary 接口给 Blog-termux。

### 5.5 日历聚合

`GET /api/calendar?month=YYYY-MM` 一次性聚合返回：`calendar_events` + `todos.deadline_at` + `countdown_events` + 有日记的 `entry_date`。前端渲染月视图，不是独立数据孤岛。

---

## 6. API 设计

所有响应统一 envelope：`{"ok": true, "data": ...}` 或 `{"ok": false, "error": "..."}`

```
GET    /api/todos              ?status=&has_deadline=
POST   /api/todos
GET    /api/todos/{id}
PUT    /api/todos/{id}
DELETE /api/todos/{id}

GET    /api/countdown           (含 todo 来源 + 手动，混排)
POST   /api/countdown           (仅手动创建)
DELETE /api/countdown/{id}

POST   /api/pomodoro/start      body: {planned_minutes, linked_todo_id}
POST   /api/pomodoro/{id}/finish body: {status: "completed"|"aborted"}
GET    /api/pomodoro/today       → {total_minutes}
GET    /api/pomodoro             ?from=&to=

GET    /api/diary/{date}         date=YYYY-MM-DD
PUT    /api/diary/{date}         保存触发 ledger 同步
GET    /api/diary                ?month=YYYY-MM

GET    /api/ledger               ?month=YYYY-MM
GET    /api/ledger/summary       ?month=YYYY-MM  → MonthlySummary
POST   /api/ledger
DELETE /api/ledger/{id}

GET    /api/calendar             ?month=YYYY-MM  → MonthView (聚合)
POST   /api/calendar/events
DELETE /api/calendar/events/{id}

GET    /api/summary              给 Blog-termux 卡片用的精简接口
```

`GET /api/summary` 响应示例：

```json
{
  "ok": true,
  "data": {
    "calendar": {"month": "2026-06", "today": "2026-06-21"},
    "ledger": {"year_month": "2026-06", "expense": 1234.50, "income": 6000.00, "balance": 4765.50, "savings": 88234.00},
    "countdown": [{"title": "项目截止", "days_left": 5, "target_at": "..."}],
    "focus_today_minutes": 95
  }
}
```

---

## 7. 变量隔离

### Go（编译器强制）
- **零 package 级可变变量**。只允许 `const` 和 `init()` 编译时注册
- Service/Handler 持有的依赖全是**结构体非导出字段**，外部包拿不到
- `main.go` 全部局部变量，显式依赖注入

### JavaScript（ES Module 作用域）
- 每个 JS 文件 `'use strict'`，顶层只用 `const`/`let`
- 模块间只通过 `import`/`export` 通信，禁止互相读写内部变量
- 唯一全局命名空间：`window.TermuxUpdate = { VERSION, MODULES }`（仅调试用）
- CSS 类名前缀 `tu-`：`.tu-card`、`.tu-btn`、`.tu-summary-grid`，避免与 Blog-termux 冲突

---

## 8. 生命周期管理

### 后端（main.go）

```
[启动]
1. config.Load("config.json")            → 失败直接 os.Exit
2. store.New(cfg.Database)               → 自动 migrate
3. ledgerSvc := ledger.NewService(st)
4. countSvc  := countdown.NewService(st)
5. todoSvc   := todo.NewService(st, countSvc)     // 注入 countdown 做联动
6. pomoSvc   := pomodoro.NewService(st)
7. diarySvc  := diary.NewService(st, ledgerSvc)   // 注入 ledger 做联动
8. calSvc    := calendar.NewService(st, todoSvc, countSvc, diarySvc)
9. summarySvc := summary.NewService(ledgerSvc, countSvc, pomoSvc)
10. mux := http.NewServeMux()
    各模块.Register(mux, svc) 注册路由
    mux.Handle("/", http.FileServerFS(webFS))
11. srv.ListenAndServe()

[关闭]
12. SIGINT/SIGTERM → 5s graceful shutdown → st.Close()
```

### 前端（app.js，复刻 Blog-termux 模式）

```
[DOMContentLoaded]
1. Theme.initTheme()
2. Calendar.init() / Todo.init() / Countdown.init() / Pomodoro.init() / Diary.init() / Ledger.init()
3. 解析 hash → switchTab(initialTab)
4. 注册 visibilitychange / hashchange

[Tab 切换 — switchTab(tabId)]
- 离开旧 tab → module.onTabLeave()（停轮询/计时器）
- 进入新 tab → module.onTabEnter()（取数据/启轮询）
- 更新 tab bar active 状态
- 更新 hash

[可见性变化]
- document.hidden → 当前 tab onTabLeave()
- !document.hidden → 当前 tab onTabEnter()

[Tab 内部模式（与 Blog-termux dashboard.js 一致）]
let _timer = null;       // setInterval handle
let _fetching = false;   // 防并发
let _tabActive = false;  // 当前 tab 是否可见

function onTabEnter() { _tabActive = true; if (!document.hidden) fetchData(); }
function onTabLeave() { _tabActive = false; clearInterval(_timer); }
```

Tab 结构（PC/Pad 顶部横排，手机底部固定栏）：
`[日历总览] [TODO] [番茄钟] [日记记账] [倒计时]`

日历总览作为默认首屏，聚合显示当月 TODO 截止日、倒计时、记账月结、日记打点。

---

## 9. Blog-termux 集成方案

### 9.1 重写 index.html（放在本项目 `web/blog-termux-index.html`）

基于 Blog-termux 原版 `index.html`，在 dashboard 8 卡片后追加第 9 张卡片：

```html
<!-- 9. 日常概览（integration card） -->
<div class="dash-card tu-summary-card" id="summaryCard">
    <div class="dash-icon">📋</div>
    <div class="dash-info">
        <div class="dash-label">日常概览</div>
        <div class="tu-summary-grid" id="summaryGrid">
            <div class="tu-summary-item">
                <span class="tu-summary-label">本月支出</span>
                <span class="tu-summary-value" id="sumExpense">--</span>
            </div>
            <!-- ... 本月收入、月结余、储蓄、今日专注、倒计时 ... -->
        </div>
    </div>
    <div class="tu-summary-footer">
        <a href="/update/" class="tu-summary-link">进入日常管理 →</a>
    </div>
</div>
```

- 复用 `.dash-card` 基础类（毛玻璃、hover 上浮），加 `tu-summary-card` 做差异化
- 响应式：桌面 4 列 grid 下 `grid-column: span 2`，平板 2 列也 `span 2`，手机 1 列正常
- 点击整卡或链接跳转到 `/update/`（nginx 反代到 Go 后端）

### 9.2 集成卡片 JS（`web/js/update-widget.js`）

- `init()` 启动 30s 轮询 `fetch('/api/summary')`
- 渲染 calendar/ledger/countdown/focus_today_minutes 到卡片
- 与 dashboard.js 相同的 fetch 去重 + 错误降级模式

### 9.3 Nginx 反代（Blog-termux 侧 `example/Blog.conf`）

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

浏览器始终请求同源（Blog-termux 的 7443 端口），不需要 CORS。

### 9.4 部署

```bash
# 部署重写的 index.html 到 Blog-termux
cp web/blog-termux-index.html /root/Projects/Blog/index.html
# 复制 update-widget.js
cp web/js/update-widget.js /root/Projects/Blog/js/update-widget.js
# 在 Blog-termux 的 app.js 中添加 import { UpdateWidget } from './update-widget.js'
# 重启 nginx
```

---

## 10. 测试策略

### 后端

**单元测试（SQLite `:memory:`）：**
- Todo+Countdown 联动：创建/更新/删除 todo → 验证 countdown 同步
- Diary+Ledger 联动：保存含 ledger 块的日记 → 幂等替换验证
- Ledger 月度汇总：跨月边界正确性
- Calendar 聚合：多数据源合并正确性
- Pomodoro：今日分钟数计算

**Handler 测试（`net/http/httptest`）：**
- 全部 API 端点：状态码 + JSON 结构 + envelope 格式

**三存储契约测试：**
- 同一套测试用例，build tag 切换 sqlite/mysql/couchdb

**静态检查：** `go vet ./...`

### 前端
- `web/js/pure/*.js` 纯逻辑函数用 `node --test` 测试（date-utils、ledger-parse、countdown-calc）
- `scripts/smoke.sh`：curl 健康检查 + TODO 联动验收场景

### 联动验收清单

| 场景 | 验证点 |
|---|---|
| TODO 设置 deadline | 倒计时列表出现，天数正确 |
| TODO deadline 清空/完成 | 倒计时对应项消失 |
| 日记保存含 ledger 块 | ledger_entries 新增，月汇总更新 |
| 重复保存日记（改金额） | 旧记录被替换，不叠加 |
| Blog-termux 新卡片 | `/api/summary` 字段齐全，点击跳转正确 |
| sqlite→mysql→couchdb 切换 | 同套 API 返回结构一致 |

---

## 11. 实施顺序

| 阶段 | 内容 | 依赖 |
|---|---|---|
| Phase 0 | 项目脚手架：go.mod、目录结构、config.go、idgen.go、httputil | 无 |
| Phase 1 | Store 抽象层 + SQLite 实现 + migrate | Phase 0 |
| Phase 2 | Todo + Countdown 模块（联动逻辑验证）| Phase 1 |
| Phase 3 | Ledger + Diary 模块（Markdown 解析联动验证）| Phase 1 |
| Phase 4 | Pomodoro + Calendar 聚合 | Phase 1-3 |
| Phase 5 | Summary 接口 + 前端 SPA（5 个 Tab + CSS）| Phase 2-4 |
| Phase 6 | MySQL + CouchDB 后端补齐 | Phase 1 |
| Phase 7 | Blog-termux 集成：重写 index.html + update-widget.js + nginx + deploy 脚本 | Phase 5 |

---

## 12. 关键注意事项

- **时区**：`entry_date` 存用户本地日期 `YYYY-MM-DD`（不转时区）。Pomodoro 的"今日"用可配置时区（`config.json` 中 `timezone` 字段，默认 UTC）
- **SQLite WAL 模式**：连接字符串加 `?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)`
- **金额解析**：支持 `"1,234.56"` 千分位格式，`strings.Replace` 后 `strconv.ParseFloat`，再 `math.Round(f*100)` 转分
- **ledger 块解析容错**：缺必填字段/非法 type → 静默跳过该块，不阻止整篇日记保存
- **CouchDB**：每表一个 database，Mango `_find` 查询，无 JavaScript 视图
- **开发期 CORS**：`cfg.Server.CORS = true` 时加简单 CORS 中间件（生产走 nginx 反代不需要）
- **CSS 类名前缀 `tu-`**：所有新项目组件类名加此前缀，避免嵌入 Blog-termux 页面时冲突
