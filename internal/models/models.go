package models

// Category is a user-defined task grouping with a Catppuccin accent color.
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

// Recurrence is a template + rule from which concrete Task instances are
// lazily materialized per day.
type Recurrence struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	Notes            string  `json:"notes"`
	CategoryID       *int64  `json:"category_id"`
	Priority         string  `json:"priority"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	ScheduledStart   *string `json:"scheduled_start"`
	Freq             string  `json:"freq"` // daily|weekly|monthly
	Interval         int     `json:"interval"`
	DaysOfWeek       string  `json:"days_of_week"`
	DayOfMonth       *int    `json:"day_of_month"`
	StartDate        string  `json:"start_date"`
	EndDate          *string `json:"end_date"`
	Active           bool    `json:"active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// Task is a single actionable item, optionally materialized from a Recurrence.
type Task struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	Notes            string  `json:"notes"`
	CategoryID       *int64  `json:"category_id"`
	Priority         string  `json:"priority"` // low|medium|high|urgent
	Status           string  `json:"status"`   // todo|in_progress|paused|done
	EstimatedMinutes int     `json:"estimated_minutes"`
	ActualSeconds    int64   `json:"actual_seconds"`
	ScheduledDate    *string `json:"scheduled_date"`
	ScheduledStart   *string `json:"scheduled_start"`
	ActiveStartedAt  *string `json:"active_started_at"`
	CompletedAt      *string `json:"completed_at"`
	RecurrenceID     *int64  `json:"recurrence_id"`
	SortOrder        int     `json:"sort_order"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`

	// Computed, not persisted.
	ElapsedSeconds int64 `json:"elapsed_seconds"`
	IsActive       bool  `json:"is_active"`
}

// TimerSession records one contiguous run of active time on a task, the
// source of truth for time analytics.
type TimerSession struct {
	ID              int64   `json:"id"`
	TaskID          int64   `json:"task_id"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	DurationSeconds int64   `json:"duration_seconds"`
}

// JournalEntry is a markdown journal note for a given date.
type JournalEntry struct {
	ID        int64   `json:"id"`
	Date      string  `json:"date"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Mood      *string `json:"mood"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// AISummary is a cached AI-generated growth summary for a period.
type AISummary struct {
	ID         int64  `json:"id"`
	PeriodType string `json:"period_type"` // day|week|month
	PeriodKey  string `json:"period_key"`
	Content    string `json:"content"`
	Model      string `json:"model"`
	CreatedAt  string `json:"created_at"`
}

// CalendarDay is a single day's aggregate state for the month view.
type CalendarDay struct {
	Date  string  `json:"date"`
	Total int     `json:"total"`
	Done  int     `json:"done"`
	Ratio float64 `json:"ratio"`
	State string  `json:"state"` // empty|complete|partial|missed
}

// Stats is the aggregate analytics payload for a date range.
type Stats struct {
	TasksCompleted        int              `json:"tasks_completed"`
	TasksTotal            int              `json:"tasks_total"`
	CompletionRate        float64          `json:"completion_rate"`
	EstimatedMinutesTotal int              `json:"estimated_minutes_total"`
	ActualSecondsTotal    int64            `json:"actual_seconds_total"`
	EstVsActualPerDay     []EstVsActualDay `json:"est_vs_actual"`
	TimeByCategory        []CategoryTime   `json:"time_by_category"`
	CurrentStreak         int              `json:"current_streak"`
	BusiestCategory       string           `json:"busiest_category"`
	AvgDailyCompletion    float64          `json:"avg_daily_completion"`
}

type EstVsActualDay struct {
	Date             string `json:"date"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualSeconds    int64  `json:"actual_seconds"`
	Total            int    `json:"total"`
	Done             int    `json:"done"`
}

type CategoryTime struct {
	CategoryID   *int64 `json:"category_id"`
	CategoryName string `json:"category_name"`
	Seconds      int64  `json:"seconds"`
}
