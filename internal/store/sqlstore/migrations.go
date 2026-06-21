package sqlstore

const schemaSQLite = `
CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    notes TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    deadline_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE TABLE IF NOT EXISTS countdown_events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    target_at TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'manual',
    ref_id TEXT,
    note TEXT DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id TEXT PRIMARY KEY,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    planned_minutes INTEGER NOT NULL DEFAULT 25,
    actual_minutes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'completed',
    linked_todo_id TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS diary_entries (
    id TEXT PRIMARY KEY,
    entry_date TEXT NOT NULL UNIQUE,
    content_md TEXT NOT NULL DEFAULT '',
    mood TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id TEXT PRIMARY KEY,
    entry_date TEXT NOT NULL,
    type TEXT NOT NULL,
    amount_cents INTEGER NOT NULL,
    category TEXT NOT NULL,
    note TEXT DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual',
    source_diary_id TEXT DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS calendar_events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_at TEXT NOT NULL,
    end_at TEXT,
    all_day INTEGER NOT NULL DEFAULT 0,
    note TEXT DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_deadline ON todos(deadline_at);
CREATE INDEX IF NOT EXISTS idx_countdown_ref ON countdown_events(ref_id);
CREATE INDEX IF NOT EXISTS idx_ledger_date ON ledger_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_ledger_diary ON ledger_entries(source_diary_id);
CREATE INDEX IF NOT EXISTS idx_diary_date ON diary_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_calendar_start ON calendar_events(start_at);
CREATE INDEX IF NOT EXISTS idx_pomodoro_date ON pomodoro_sessions(started_at);
`

// MySQL schema uses VARCHAR for datetime columns (store RFC3339 strings, same as SQLite).
// VARCHAR(35) holds full RFC3339: "2006-01-02T15:04:05.999999999Z07:00"
const schemaMySQL = `
CREATE TABLE IF NOT EXISTS todos (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 0,
    deadline_at VARCHAR(35) NULL,
    created_at VARCHAR(35) NOT NULL,
    updated_at VARCHAR(35) NOT NULL,
    completed_at VARCHAR(35) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS countdown_events (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    target_at VARCHAR(35) NOT NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'manual',
    ref_id VARCHAR(64),
    note TEXT,
    created_at VARCHAR(35) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id VARCHAR(64) PRIMARY KEY,
    started_at VARCHAR(35) NOT NULL,
    ended_at VARCHAR(35) NULL,
    planned_minutes INT NOT NULL DEFAULT 25,
    actual_minutes INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    linked_todo_id VARCHAR(64) DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS diary_entries (
    id VARCHAR(64) PRIMARY KEY,
    entry_date VARCHAR(10) NOT NULL UNIQUE,
    content_md TEXT NOT NULL,
    mood VARCHAR(20) DEFAULT '',
    created_at VARCHAR(35) NOT NULL,
    updated_at VARCHAR(35) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ledger_entries (
    id VARCHAR(64) PRIMARY KEY,
    entry_date VARCHAR(10) NOT NULL,
    type VARCHAR(10) NOT NULL,
    amount_cents BIGINT NOT NULL,
    category VARCHAR(100) NOT NULL,
    note TEXT,
    source VARCHAR(20) NOT NULL DEFAULT 'manual',
    source_diary_id VARCHAR(64) DEFAULT '',
    created_at VARCHAR(35) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS calendar_events (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    start_at VARCHAR(35) NOT NULL,
    end_at VARCHAR(35) NULL,
    all_day TINYINT NOT NULL DEFAULT 0,
    note TEXT,
    created_at VARCHAR(35) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS settings (
    ` + "`key`" + ` VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`
