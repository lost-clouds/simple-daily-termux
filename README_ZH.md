# simple-daily-termux — 个人日常管理工具

[简体中文](README_ZH.md) | [English](README.md)

[![CI](https://github.com/lost-clouds/simple-daily-termux/actions/workflows/ci.yml/badge.svg)](https://github.com/lost-clouds/simple-daily-termux/actions/workflows/ci.yml)

轻量级个人时间管理应用，Go 后端 + 原生 JavaScript SPA 前端。集成 **TODO 待办**、**番茄钟**、**日记记账**、**日历聚合**、**倒计时** 五大模块，编译为单个二进制文件。设计上与 [Blog-termux](https://github.com/lost-clouds/Blog-termux) 通过 nginx 反代联动，在仪表盘新增第 9 张卡片展示日常概览。

> 为 Android Termux 环境而生。纯 Go（零 CGO），单二进制部署，内存占用 <20MB。

---

## 目录

- [快速开始](#快速开始)
- [目录结构](#目录结构)
- [架构设计](#架构设计)
- [模块详解](#模块详解)
- [部署教程](#部署教程)
  - [1. 环境要求](#1-环境要求)
  - [2. 下载预编译包](#2-下载预编译包)
  - [3. 从源码编译](#3-从源码编译)
  - [4. 配置文件](#4-配置文件)
  - [5. 运行](#5-运行)
  - [6. 与 Blog-termux 联动](#6-与-blog-termux-联动)
- [API 参考](#api-参考)
- [使用指南](#使用指南)
- [常见问题](#常见问题)

---

## 快速开始

**方式 A — 下载预编译包（推荐）：**

```bash
# 从 Releases 页面下载对应平台的二进制包
# 提供 linux-amd64 / linux-arm64 / linux-armv7 三种架构
curl -sSLO https://github.com/lost-clouds/simple-daily-termux/releases/latest/download/simple-daily-termux-linux-arm64.tar.gz
tar -xzf simple-daily-termux-linux-arm64.tar.gz
# 包含：simple-daily-termux-linux-arm64、config.example.json、smoke.sh
```

**方式 B — 从源码编译：**

```bash
# 1. 克隆
git clone https://github.com/lost-clouds/simple-daily-termux.git ~/simple-daily-termux
cd ~/simple-daily-termux

# 2. 编译（Go 1.22+）
go build -o simple-daily-termux .

# 3. 构建 CSS
bash web/css/build.sh
```

**安装后操作：**

```bash
# 4. 创建配置
cp config.example.json config.json
# 按需编辑（默认：SQLite ./data/daily.db，端口 8090）

# 5. 启动
./simple-daily-termux config.json
# 服务运行于 http://127.0.0.1:8090
```

运行冒烟测试：

```bash
bash scripts/smoke.sh
# 12 项检查应全部 PASS
```

> **当前版本**：[v0.0.1](https://github.com/lost-clouds/simple-daily-termux/releases/tag/v0.0.1)

---

## 目录结构

```
simple-daily-termux/
├── main.go                          # 入口：config → store → services → handlers → server
├── go.mod / go.sum                  # Go 模块（仅 2 个直连依赖：sqlite + mysql 驱动）
├── config.json / config.example.json
├── internal/
│   ├── config/config.go             # Config 结构体 + Load() + Validate()
│   ├── idgen/idgen.go               # 应用层 ID 生成（crypto/rand，不引入 uuid 库）
│   ├── httputil/response.go         # 统一 JSON 响应封装
│   ├── store/sqlstore/
│   │   ├── store.go                 # Store 接口聚合 + 全部 Repository 实现
│   │   ├── sqlite.go                # SQLite 初始化（WAL 模式，纯 Go 无 CGO）
│   │   ├── mysql.go                 # MySQL 初始化
│   │   └── migrations.go            # DDL（SQLite + MySQL 双语法）
│   ├── todo/        {model, service, handler}.go
│   ├── countdown/   {model, service, handler}.go
│   ├── pomodoro/    {model, service, handler}.go
│   ├── diary/       {model, service, handler, ledgerparser}.go
│   ├── ledger/      {model, service, handler}.go
│   ├── calendar/    {model, service, handler}.go
│   └── summary/     {service, handler}.go
├── web/
│   ├── index.html                    # 独立 SPA（5 个 Tab）
│   ├── blog-termux-index.html        # Blog-termux 集成版 index.html（含第 9 张联动卡片）
│   ├── css/
│   │   ├── build.sh                  # CSS 构建脚本（cat 合并源文件）
│   │   ├── style.css                 # 合并产物（go:embed 嵌入二进制）
│   │   └── src/                      # CSS 源文件（模块化拆分）
│   │       ├── variables.css         #   CSS 自定义属性（与 Blog-termux 同名同值）
│   │       ├── base.css / layout.css / responsive.css
│   │       ├── themes/dark.css
│   │       └── components/*.css      #   8 个组件样式表
│   ├── js/                           # ES Modules（零打包器）
│   │   ├── main.js → app.js          #   入口 → 主控制器
│   │   ├── constants.js              #   API 路径常量
│   │   ├── utils.js                  #   共享工具函数（esc 等）
│   │   ├── theme.js                  #   主题管理器
│   │   ├── calendar.js / todo.js / countdown.js
│   │   ├── pomodoro.js / diary.js / ledger.js
│   │   └── update-widget.js          #   Blog-termux 集成卡片
│   └── lib/marked.min.js             # Markdown 解析器
├── scripts/
│   ├── start.sh / stop.sh            # 进程管理（PID 文件）
│   └── smoke.sh                      # curl 冒烟测试
└── plan.md                           # 设计计划
```

---

## 架构设计

### 整体架构

```
simple-daily-termux (Go 二进制)
  │
  ├── main.go ── 依赖注入树
  │   ├── config.Load("config.json")
  │   ├── sqlstore.NewSQLite()          → Store
  │   ├── ledger.NewService()           → Ledger + Settings 仓库
  │   ├── countdown.NewService()        → Countdown 仓库
  │   ├── todo.NewService(_, countSvc)  → Todo 仓库 + countdown 联动
  │   ├── pomodoro.NewService()         → Pomodoro 仓库
  │   ├── diary.NewService(_, ledgerSvc)→ Diary 仓库 + ledger 联动
  │   ├── calendar.NewService(...)      → 聚合 4 个仓库
  │   └── summary.NewService(...)       → Ledger + Countdown + Pomodoro
  │
  ├── http.ServeMux (Go 1.22 模式路由)
  │   ├── /api/todos, /api/countdown, /api/pomodoro, ...
  │   ├── /api/summary                  → Blog-termux 集成端点
  │   └── /                             → 嵌入式 SPA（go:embed）
  │
  └── 优雅关闭 (SIGINT/SIGTERM → 5s 超时)
```

### 前端 Tab 结构

```
index.html (SPA)
  │
  ├─ header ─── 品牌标题 + 主题切换按钮 (🌓)
  │
  ├─ tab-bar ── [📅日历总览] [✅TODO] [⏱️番茄钟] [📝日记记账] [⏳倒计时]
  │              PC/平板顶部 | 手机底部固定栏
  │
  └─ 内容区 (5 个 section，同时只显示 1 个)
      ├── #sec-calendar     月视图网格 + 事件 + 截止日 + 倒计时目标
      ├── #sec-todo         待办列表 + 筛选 + 优先级 + 截止日联动
      ├── #sec-pomodoro     倒计时时钟 + 今日专注分钟
      ├── #sec-diary        Markdown 编辑器 + 记账块内嵌 + 月度汇总
      └── #sec-countdown    倒计时列表（手动 + Todo 自动生成）
```

### 模块依赖关系图

```
main.go
  ├── config.Load()
  ├── sqlstore.NewSQLite() / NewMySQL()
  ├── ledger.NewService(st.Ledgers(), st.Settings())
  ├── countdown.NewService(st.Countdowns())
  ├── todo.NewService(st.Todos(), countSvc)          ← 注入 countdown 做 deadline 同步
  ├── pomodoro.NewService(st.Pomodoros())
  ├── diary.NewService(st.Diaries(), ledgerSvc)       ← 注入 ledger 做代码块解析
  ├── calendar.NewService(st.Calendars(), st.Todos(), st.Countdowns(), st.Diaries())
  ├── summary.NewService(ledgerSvc, countSvc, pomoSvc, timezone)
  └── http.ServeMux (路由注册 + 中间件)
```

**核心设计**：所有依赖在 `main.go` 中通过构造函数参数显式注入。零包级可变状态。无需全局服务定位器。Go 的大小写可见性在编译期强制模块边界。

### 数据流

```
浏览器客户端
  │  GET /api/todos, POST /api/ledger, PUT /api/diary/2026-06-21, ...
  ↓
Go HTTP handlers → Service 层（业务逻辑）→ Repository 接口 → SQLStore
  │                                                    ↑
  │  Todo ↔ Countdown 联动（deadline 驱动）              │
  │  Diary ↔ Ledger 联动（Markdown 代码块解析）          │
  │  Calendar 聚合（4 数据源合并）                       │
  │  Summary 聚合（供 Blog-termux 卡片使用）             │
  ↓
JSON 响应: {"ok": true, "data": {...}}
```

---

## 模块详解

### todo — TODO 待办 + 截止日联动

| | |
|---|---|
| 包 | `internal/todo/` |
| 仓库接口 | `todo.Repository`（定义在 model.go） |
| 服务 | `TodoService{repo, countSvc}` — 注入 `*countdown.Service` |
| API | `GET/POST /api/todos`, `GET/PUT/DELETE /api/todos/{id}` |

**联动逻辑**：创建/更新 todo 设置 `deadline_at` 时自动创建倒计时；清空 deadline 或完成/删除 todo 时自动删除对应倒计时。倒计时列表同时展示手动创建和 todo 自动生成的条目。

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 任务名称 |
| `status` | pending / doing / done | 当前状态 |
| `priority` | 0-3 | 优先级 |
| `deadline_at` | RFC3339 / null | 驱动倒计时联动 |

---

### countdown — 倒计时

| | |
|---|---|
| 包 | `internal/countdown/` |
| 仓库接口 | `countdown.Repository` |
| 服务 | `CountdownService{repo}` |
| API | `GET/POST /api/countdown`, `DELETE /api/countdown/{id}` |

`days_left` 在读取时计算（基于 UTC 日期边界）。支持两种来源：`manual`（用户手动创建）和 `todo`（从 todo deadline 自动派生）。两种来源在同一列表中混排。

---

### pomodoro — 番茄钟

| | |
|---|---|
| 包 | `internal/pomodoro/` |
| 仓库接口 | `pomodoro.Repository` |
| 服务 | `PomodoroService{repo}` |
| API | `POST /api/pomodoro/start`, `POST /api/pomodoro/{id}/finish`, `GET /api/pomodoro/today`, `GET /api/pomodoro` |

会话状态：`running` → `completed` / `aborted`。`actual_minutes` 由 `ended_at - started_at` 计算。今日专注时间按配置时区计算。前端切 Tab 时自动暂停并结束服务器端 session。

| 字段 | 说明 |
|------|------|
| `planned_minutes` | 计划时长（默认 25 分钟） |
| `actual_minutes` | 实际经过时间 |
| `linked_todo_id` | 可选关联的待办事项 |

---

### diary — Markdown 日记

| | |
|---|---|
| 包 | `internal/diary/` |
| 仓库接口 | `diary.Repository` |
| 服务 | `DiaryService{repo, ledgerSvc}` — 注入 `*ledger.Service` |
| API | `GET /api/diary/{date}`, `PUT /api/diary/{date}`, `GET /api/diary?month=YYYY-MM` |

**记账联动**：保存日记时扫描 `content_md` 中的 `` ```ledger `` 代码块，先删除该日记的旧记账记录，再插入新记录。过程是幂等的——重复保存不会叠加。如果所有记账块被移除，关联条目一并清理。

````markdown
```ledger
type: expense
amount: 35.5
category: 餐饮
note: 午饭
```
````

Markdown 渲染使用 `marked.js`（含纯 JS 回退方案）。ledger 代码块被渲染为带样式的交易卡片而非代码块。

---

### ledger — 个人记账

| | |
|---|---|
| 包 | `internal/ledger/` |
| 仓库接口 | `ledger.Repository` + `SettingsRepository` |
| 服务 | `LedgerService{repo, settings}` |
| API | `GET/POST /api/ledger?month=`, `GET /api/ledger/summary?month=`, `DELETE /api/ledger/{id}` |

**金额以"分"存储**（整数），避免浮点误差。

| 指标 | 计算公式 |
|------|----------|
| 月支出 | `SUM(amount_cents) WHERE type='expense'` |
| 月收入 | `SUM(amount_cents) WHERE type='income'` |
| 月结余 | 收入 − 支出 |
| 储蓄 | `opening_savings_cents` + 历史所有月结余累计 |

只需录入一次 `opening_savings_cents`（期初储蓄），后续储蓄自动滚动计算。

---

### calendar — 日历聚合

| | |
|---|---|
| 包 | `internal/calendar/` |
| 仓库接口 | `calendar.Repository` + 3 个其他仓库（todo, countdown, diary） |
| 服务 | `CalendarService{calRepo, todoRepo, countRepo, diaryRepo}` |
| API | `GET /api/calendar?month=YYYY-MM`, `POST/PUT /api/calendar/events`, `DELETE /api/calendar/events/{id}` |

单次请求聚合 4 个数据源：
- **日历事件** — 独立日程
- **Todo 截止日** — 当月有 `deadline_at` 的 todo
- **倒计时目标** — 当月到期的倒计时
- **日记日期** — 有日记的日期（日历网格打点）

前端渲染月视图网格 + 三类聚合列表（事件/截止日/倒计时），每类以独立分节展示。

---

### summary — Blog-termux 集成卡片

| | |
|---|---|
| 包 | `internal/summary/` |
| 服务 | `SummaryService{ledgerSvc, countSvc, pomoSvc, timezone}` |
| API | `GET /api/summary` |

为 Blog-termux 第 9 张仪表盘卡片设计的精简端点：

```json
{
  "calendar": {"month": "2026-06", "today": "2026-06-21"},
  "ledger": {"expense": 1234.50, "income": 6000.00, "balance": 4765.50, "savings": 88234.00},
  "countdown": [{"title": "项目截止", "days_left": 193}],
  "focus_today_minutes": 95
}
```

---

## 部署教程

### 1. 环境要求

| 组件 | 用途 | 备注 |
|------|------|------|
| Go 1.22+ | 编译 | 需要 `http.ServeMux` 模式路由支持 |
| SQLite | 默认数据库 | 内置 `modernc.org/sqlite`（纯 Go，无 CGO） |
| MySQL（可选）| 备选数据库 | 仅当 `config.json` 中 `driver=mysql` 时需要 |
| Nginx（可选）| 反代 | 仅 Blog-termux 集成时需要 |

> **无需**：Node.js、Python、PHP、Docker、C/C++ 工具链、GPU。

### 2. 下载预编译包

预编译二进制文件发布在 [Releases](https://github.com/lost-clouds/simple-daily-termux/releases) 页面。

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | amd64 (x86_64) | `simple-daily-termux-linux-amd64.tar.gz` |
| Linux | arm64 (aarch64) | `simple-daily-termux-linux-arm64.tar.gz` |
| Linux | armv7 (arm32) | `simple-daily-termux-linux-armv7.tar.gz` |

每个压缩包包含编译好的二进制文件、`config.example.json` 和 `smoke.sh`。下载后可用 `.sha256` 文件校验完整性。

```bash
curl -sSLO https://github.com/lost-clouds/simple-daily-termux/releases/latest/download/simple-daily-termux-linux-arm64.tar.gz
curl -sSLO https://github.com/lost-clouds/simple-daily-termux/releases/latest/download/simple-daily-termux-linux-arm64.tar.gz.sha256
sha256sum -c simple-daily-termux-linux-arm64.tar.gz.sha256
tar -xzf simple-daily-termux-linux-arm64.tar.gz
```

### 3. 从源码编译

```bash
cd ~/simple-daily-termux

# 下载 Go 依赖（需要网络）
go mod tidy

# 构建 CSS
bash web/css/build.sh

# 编译
go build -o simple-daily-termux .

# 验证
./simple-daily-termux --help  # (接受 config 路径作为参数)
```

交叉编译（零 CGO，无需额外工具链）：

```bash
GOOS=linux GOARCH=arm64 go build -o simple-daily-termux .
GOOS=linux GOARCH=amd64 go build -o simple-daily-termux .
```

### 4. 配置文件

复制并编辑 `config.json`：

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

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `server.addr` | `127.0.0.1:8090` | 监听地址 |
| `server.cors` | `false` | 是否启用 CORS（开发用） |
| `database.driver` | `sqlite` | 数据库驱动：`sqlite` 或 `mysql` |
| `database.sqlite.path` | `./data/daily.db` | SQLite 数据库文件路径 |
| `database.mysql.dsn` | — | MySQL 连接串（`driver=mysql` 时使用） |
| `database.timezone` | `Local` | 番茄钟"今日"计算的时区 |

### 5. 运行

**手动启动：**

```bash
./simple-daily-termux config.json
# 监听于 127.0.0.1:8090
```

**后台运行（使用脚本）：**

```bash
bash scripts/start.sh
bash scripts/stop.sh
```

**冒烟测试：**

```bash
bash scripts/smoke.sh
# 12 项检查应全部 PASS
```

加入 crontab 自动重启：

```bash
# crontab -e
# */5 * * * * cd ~/simple-daily-termux && bash scripts/start.sh
```

### 6. 与 Blog-termux 联动

**Step 1 — 添加 nginx 反代规则**

在 Blog-termux 的 nginx 配置中添加：

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

**Step 2 — 部署重写的 index.html**

```bash
cp web/blog-termux-index.html /path/to/Blog-termux/index.html
cp web/js/update-widget.js   /path/to/Blog-termux/js/update-widget.js
```

**Step 3 — 重载 nginx**

```bash
nginx -s reload
```

Blog-termux 仪表盘将显示第 9 张卡片"日常概览"，实时展示 simple-daily-termux 的数据。点击卡片跳转到完整 SPA。

---

## API 参考

所有响应使用统一 envelope 格式：`{"ok": true, "data": ...}` 或 `{"ok": false, "error": "..."}`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/todos` | 列出待办（`?status=&has_deadline=`） |
| POST | `/api/todos` | 创建待办 |
| GET | `/api/todos/{id}` | 获取待办详情 |
| PUT | `/api/todos/{id}` | 更新待办 |
| DELETE | `/api/todos/{id}` | 删除待办 |
| GET | `/api/countdown` | 列出全部倒计时 |
| POST | `/api/countdown` | 创建手动倒计时 |
| DELETE | `/api/countdown/{id}` | 删除倒计时 |
| POST | `/api/pomodoro/start` | 开始番茄钟 |
| POST | `/api/pomodoro/{id}/finish` | 结束会话（`{"status":"completed\|aborted"}`） |
| GET | `/api/pomodoro/today` | 今日专注总分钟 |
| GET | `/api/pomodoro` | 会话列表（`?from=&to=`） |
| GET | `/api/diary/{date}` | 获取日记（`date=YYYY-MM-DD`） |
| PUT | `/api/diary/{date}` | 保存日记（触发记账同步） |
| GET | `/api/diary` | 月列表（`?month=YYYY-MM`） |
| GET | `/api/ledger` | 记账列表（`?month=YYYY-MM`） |
| GET | `/api/ledger/summary` | 月度汇总（`?month=YYYY-MM`） |
| POST | `/api/ledger` | 创建记账条目 |
| DELETE | `/api/ledger/{id}` | 删除记账条目 |
| GET | `/api/calendar` | 聚合月视图（`?month=YYYY-MM`） |
| POST | `/api/calendar/events` | 创建日历事件 |
| PUT | `/api/calendar/events/{id}` | 更新日历事件 |
| DELETE | `/api/calendar/events/{id}` | 删除日历事件 |
| GET | `/api/summary` | 集成卡片数据 |
| GET | `/api/health` | 健康检查 |

---

## 使用指南

| 操作 | 步骤 |
|------|------|
| **切换标签** | PC/平板：点击顶部标签栏。手机：点击底部导航栏 |
| **深色模式** | 点击 🌓 按钮，偏好自动保存 |
| **创建待办** | TODO 标签 → "+ 新事项" → 填写表单 → 保存 |
| **设定截止日** | 在 todo 表单中添加截止日 → 自动出现在倒计时中 |
| **开始番茄钟** | 番茄钟标签 → 选择时长 → 开始，切 Tab 自动暂停 |
| **写日记** | 日记标签 → 选择日期 → 写 Markdown → 点击"插入记账"添加账目 |
| **查看记账** | 日记标签 → 顶部月度汇总 + 底部账目列表 |
| **浏览日历** | 日历总览标签 → 月网格 + 日记打点 + 三类聚合事件 |
| **管理倒计时** | 倒计时标签 → 手动创建 + Todo 自动生成 |
| **设置期初储蓄** | 通过 settings API 设置 `opening_savings_cents` |

---

## 常见问题

### Q: 如何从 SQLite 切换到 MySQL？

编辑 `config.json`：

```json
{
  "database": {
    "driver": "mysql",
    "mysql": { "dsn": "user:pass@tcp(127.0.0.1:3306)/daily?parseTime=true" }
  }
}
```

重启服务。表结构在启动时自动创建。已有的 SQLite 数据需手动迁移。

### Q: 数据库文件在哪里？

默认路径：`./data/daily.db`。可通过 `config.json` → `database.sqlite.path` 修改。

### Q: 如何设置期初储蓄？

`opening_savings_cents` 存储在 `settings` 表中。可通过 API 设置（以"分"为单位）：

```bash
# 期初储蓄设为 ¥50,000.00（5,000,000 分）
# 需要 settings API 端点 — 目前可通过直接操作数据库或等待设置 UI
```

### Q: 切 Tab 后番茄钟计时不准？

前端计时器使用 `setInterval`，浏览器在后台 Tab 会节流。现在切 Tab 自动暂停并结束 session，避免计时偏差。服务端记录了 `started_at`/`ended_at`，可作为准确时间来源。

### Q: 重复保存日记会不会重复记账？

不会。日记保存逻辑是幂等的——保存时先删除该日记关联的全部旧记账记录，再重新解析插入。修改金额后重保存，旧条目被替换而非叠加。

### Q: 日历显示为空？

确认至少有一类数据：日历事件、Todo 截止日、倒计时、日记。日历模块从四个数据源聚合。月初无数据时显示"本月无事件"。

### Q: Blog-termux 仪表盘卡片显示 "--"？

排查：
1. simple-daily-termux 是否在运行？`curl http://127.0.0.1:8090/api/health`
2. nginx 反代是否配置？`curl http://127.0.0.1:7443/api/summary`
3. `update-widget.js` 是否已部署到 Blog-termux？

---

## 技术要点

| 特性 | 实现方式 |
|------|----------|
| 零框架 | Go `net/http` 标准库 + `http.ServeMux` 模式路由 |
| 纯 Go SQLite | `modernc.org/sqlite` — 零 CGO，任意平台交叉编译 |
| 单二进制 | 前端通过 `go:embed` 嵌入，部署只需一个文件 |
| 模块化架构 | handler → service → repository，每包自包含 |
| 变量隔离 | Go：零包级可变状态。JS：ES Module 作用域 + `tu-` CSS 前缀 |
| 金额精度 | 整数分（int64），`math.Round(f * 100)` 转换 |
| 幂等同步 | 日记记账：每次保存先删旧再插新 |
| 统一响应 | 全端点 `{"ok": true/false, "data": ...}` |
| 主题 | CSS 变量 + `body.dark` 切换（与 Blog-termux 同一套配色） |
| 响应式 | 4 个断点（1024/640/400px），手机底部导航 |
| 直角风格 | `border-radius: 0` — 与 Blog-termux 圆角风做视觉区分 |
| Markdown 渲染 | marked.js + ledger 代码块预处理 |
| 零构建工具 | ES Modules（零打包器），CSS 通过 `cat` 合并 |
| 优雅关闭 | SIGINT/SIGTERM → 5s context 超时 → server.Shutdown |
| 时区感知 | 可配置时区用于番茄钟"今日"计算 |

---

## 友链

[Blog-termux](https://github.com/lost-clouds/Blog-termux) — 联动仪表盘项目
