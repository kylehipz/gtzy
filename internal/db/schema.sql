CREATE TABLE IF NOT EXISTS categories (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL UNIQUE,
  color      TEXT NOT NULL DEFAULT 'mauve',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS recurrences (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  title           TEXT NOT NULL,
  notes           TEXT NOT NULL DEFAULT '',
  category_id     INTEGER REFERENCES categories(id) ON DELETE SET NULL,
  priority        TEXT NOT NULL DEFAULT 'medium',
  estimated_minutes INTEGER NOT NULL DEFAULT 0,
  scheduled_start TEXT,
  freq            TEXT NOT NULL,
  interval        INTEGER NOT NULL DEFAULT 1,
  days_of_week    TEXT NOT NULL DEFAULT '',
  day_of_month    INTEGER,
  start_date      TEXT NOT NULL,
  end_date        TEXT,
  active          INTEGER NOT NULL DEFAULT 1,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  title           TEXT NOT NULL,
  notes           TEXT NOT NULL DEFAULT '',
  category_id     INTEGER REFERENCES categories(id) ON DELETE SET NULL,
  priority        TEXT NOT NULL DEFAULT 'medium',
  status          TEXT NOT NULL DEFAULT 'todo',
  estimated_minutes INTEGER NOT NULL DEFAULT 0,
  actual_seconds  INTEGER NOT NULL DEFAULT 0,
  scheduled_date  TEXT,
  scheduled_start TEXT,
  active_started_at TEXT,
  completed_at    TEXT,
  recurrence_id   INTEGER REFERENCES recurrences(id) ON DELETE SET NULL,
  sort_order      INTEGER NOT NULL DEFAULT 0,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tasks_date ON tasks(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_recur_day
  ON tasks(recurrence_id, scheduled_date) WHERE recurrence_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS timer_sessions (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  started_at TEXT NOT NULL,
  ended_at   TEXT,
  duration_seconds INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sessions_task ON timer_sessions(task_id);

CREATE TABLE IF NOT EXISTS journal_entries (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  date       TEXT NOT NULL,
  title      TEXT NOT NULL DEFAULT '',
  content    TEXT NOT NULL DEFAULT '',
  mood       TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_journal_date ON journal_entries(date);

CREATE TABLE IF NOT EXISTS ai_summaries (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  period_type TEXT NOT NULL,
  period_key  TEXT NOT NULL,
  content     TEXT NOT NULL,
  model       TEXT NOT NULL,
  created_at  TEXT NOT NULL,
  UNIQUE(period_type, period_key)
);

CREATE TABLE IF NOT EXISTS blood_sugar_readings (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  value_mgdl REAL    NOT NULL,
  taken_at   TEXT    NOT NULL,                  -- RFC3339, from meter or user
  meal_tag   TEXT    NOT NULL DEFAULT '',       -- fasting|pre_meal|post_meal|bedtime|random|''
  notes      TEXT    NOT NULL DEFAULT '',
  source     TEXT    NOT NULL DEFAULT 'manual', -- manual|meter|import
  seq_number INTEGER,                           -- meter record sequence #, NULL for manual
  created_at TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bsr_taken_at ON blood_sugar_readings(taken_at);
-- Idempotent meter sync: re-running sync never double-inserts a meter record.
-- Mirrors the recurrence idempotency-by-DB-constraint pattern (idx_tasks_recur_day).
CREATE UNIQUE INDEX IF NOT EXISTS idx_bsr_meter_seq
  ON blood_sugar_readings(seq_number) WHERE source = 'meter' AND seq_number IS NOT NULL;
