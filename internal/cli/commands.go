package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"gtzy/internal/models"
)

func today() string { return time.Now().Format("2006-01-02") }

// --- gtzy sync ---

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Pull new readings from the paired Bluetooth glucose meter",
		Long: `Asks the running gtzy server to connect to the bonded Accu-Chek meter over
Bluetooth and import any new blood-sugar records. The meter must already be
paired to this machine (pair once with bluetoothctl); set GTZY_METER_ADDR to
its MAC, or leave it unset to scan by name. Re-running sync never duplicates
records already imported.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync()
		},
	}
}

func runSync() error {
	body, err := httpPost("/api/bloodsugar/sync", nil)
	if err != nil {
		return err
	}
	var res struct {
		Synced  int `json:"synced"`
		Fetched int `json:"fetched"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	fmt.Printf("synced %d new reading(s) (%d fetched from meter)\n", res.Synced, res.Fetched)
	return nil
}

// --- gtzy next ---

func newNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Advance to the next eligible task",
		Long:  "Pauses whatever task is currently running (if any) and starts the next eligible task, chosen by priority and schedule. Prints \"no eligible task — idle\" if nothing qualifies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNext()
		},
	}
}

func runNext() error {
	body, err := httpPost("/api/timer/next", nil)
	if err != nil {
		return err
	}
	var res struct {
		Current *models.Task `json:"current"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Current == nil {
		fmt.Println("no eligible task — idle")
		return nil
	}
	fmt.Println(res.Current.Title)
	return nil
}

// --- gtzy start <id> ---

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <id>",
		Short: "Start a task by id",
		Long:  "Starts the given task, pausing whatever else is currently running. The id is printed by `gtzy list`.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gtzy start <id>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(args[0])
		},
	}
}

func runStart(idArg string) error {
	id, err := strconv.ParseInt(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id %q", idArg)
	}
	body, err := httpPost(fmt.Sprintf("/api/tasks/%d/start", id), nil)
	if err != nil {
		return err
	}
	var t models.Task
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}
	fmt.Printf("started: %s\n", t.Title)
	return nil
}

// --- gtzy pause ---

func newPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause",
		Short: "Pause the currently running task",
		Long:  "Pauses whatever task is currently running. Prints \"nothing was running\" if the timer was already idle.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPause()
		},
	}
}

func runPause() error {
	body, err := httpPost("/api/timer/pause", nil)
	if err != nil {
		return err
	}
	var res struct {
		Current *models.Task `json:"current"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Current == nil {
		fmt.Println("nothing was running")
		return nil
	}
	fmt.Printf("paused: %s (%s elapsed)\n", res.Current.Title, fmtDuration(res.Current.ElapsedSeconds))
	return nil
}

// --- gtzy complete [<id>] ---

func newCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete [id]",
		Short: "Complete the current task, or a given task by id",
		Long:  "Marks a task done. With no id, completes whatever task is currently running (error if nothing is running). With an id, completes that task regardless of its running state.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idArg := ""
			if len(args) == 1 {
				idArg = args[0]
			}
			return runComplete(idArg)
		},
	}
}

func runComplete(idArg string) error {
	var id int64
	if idArg != "" {
		parsed, err := strconv.ParseInt(idArg, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q", idArg)
		}
		id = parsed
	} else {
		body, err := httpGet("/api/timer/current")
		if err != nil {
			return err
		}
		var res struct {
			Current *models.Task `json:"current"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			return err
		}
		if res.Current == nil {
			return fmt.Errorf("no task is currently running; pass an id: gtzy complete <id>")
		}
		id = res.Current.ID
	}

	body, err := httpPost(fmt.Sprintf("/api/tasks/%d/complete", id), nil)
	if err != nil {
		return err
	}
	var t models.Task
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}
	fmt.Printf("completed: %s\n", t.Title)
	return nil
}

// --- gtzy current ---

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the currently running task",
		Long:  "Prints the running task's title, elapsed time, and estimate. Prints \"idle — no task running\" if nothing is running.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent()
		},
	}
}

func runCurrent() error {
	body, err := httpGet("/api/timer/current")
	if err != nil {
		return err
	}
	var res struct {
		Current *models.Task `json:"current"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Current == nil {
		fmt.Println("idle — no task running")
		return nil
	}
	t := res.Current
	est := "no estimate"
	if t.EstimatedSeconds > 0 {
		est = fmt.Sprintf("est %s", fmtDuration(int64(t.EstimatedSeconds)))
	}
	fmt.Printf("%s — %s elapsed (%s)\n", t.Title, fmtDuration(t.ElapsedSeconds), est)
	return nil
}

// --- gtzy add "<title>" [--priority] [--category] [--est] [--date] [--at] [--repeat] ---

func newAddCmd() *cobra.Command {
	var priority, category, date, at, repeat string
	var est int
	cmd := &cobra.Command{
		Use:   `add "<title>" [flags]`,
		Short: "Create a task, or a recurring rule",
		Long: `Creates a one-off task, or — if --repeat is given — a recurring rule that
materializes into daily task instances.

Examples:
  gtzy add "Write report"
  gtzy add "Standup" --at 09:00 --est 15
  gtzy add "Water plants" --repeat daily
  gtzy add "Gym" --repeat weekly:1,3,5
  gtzy add "Pay rent" --repeat monthly:1`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf(`usage: gtzy add "<title>" [--priority p] [--category name] [--est minutes] [--date YYYY-MM-DD] [--at HH:MM] [--repeat daily|weekly:1,3,5|monthly:15]`)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0], priority, category, est, date, at, repeat)
		},
	}
	cmd.Flags().StringVar(&priority, "priority", "medium", `task priority: "low", "medium", or "high"`)
	cmd.Flags().StringVar(&category, "category", "", "category name (must already exist)")
	cmd.Flags().IntVar(&est, "est", 0, "estimated duration in minutes")
	cmd.Flags().StringVar(&date, "date", today(), "scheduled date, YYYY-MM-DD (used as the recurrence's start date when --repeat is set)")
	cmd.Flags().StringVar(&at, "at", "", "scheduled start time, HH:MM")
	cmd.Flags().StringVar(&repeat, "repeat", "", `recurrence: "daily", "weekly:1,3,5" (0=Sun..6=Sat), or "monthly:15"`)
	return cmd
}

func runAdd(title, priority, category string, est int, date, at, repeat string) error {
	var categoryID *int64
	if category != "" {
		id, err := lookupCategoryID(category)
		if err != nil {
			return err
		}
		categoryID = &id
	}

	var scheduledStart *string
	if at != "" {
		scheduledStart = &at
	}

	if repeat != "" {
		freq, interval, daysOfWeek, dayOfMonth, err := parseRepeat(repeat)
		if err != nil {
			return err
		}
		payload := map[string]any{
			"title": title, "priority": priority, "estimated_seconds": est * 60,
			"scheduled_start": scheduledStart, "freq": freq, "interval": interval,
			"days_of_week": daysOfWeek, "day_of_month": dayOfMonth, "start_date": date,
		}
		if categoryID != nil {
			payload["category_id"] = *categoryID
		}
		if _, err := httpPost("/api/recurrences", payload); err != nil {
			return err
		}
		fmt.Printf("created recurring rule: %s (%s)\n", title, repeat)
		return nil
	}

	payload := map[string]any{
		"title": title, "priority": priority, "estimated_seconds": est * 60,
		"scheduled_date": date, "scheduled_start": scheduledStart,
	}
	if categoryID != nil {
		payload["category_id"] = *categoryID
	}
	body, err := httpPost("/api/tasks", payload)
	if err != nil {
		return err
	}
	var t models.Task
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}
	fmt.Printf("added task #%d: %s\n", t.ID, t.Title)
	return nil
}

func parseRepeat(spec string) (freq string, interval int, daysOfWeek string, dayOfMonth *int, err error) {
	interval = 1
	parts := strings.SplitN(spec, ":", 2)
	freq = parts[0]
	switch freq {
	case "daily":
		return freq, interval, "", nil, nil
	case "weekly":
		if len(parts) != 2 || parts[1] == "" {
			return "", 0, "", nil, fmt.Errorf("weekly repeat needs weekdays, e.g. --repeat weekly:1,3,5")
		}
		return freq, interval, parts[1], nil, nil
	case "monthly":
		if len(parts) != 2 {
			return "", 0, "", nil, fmt.Errorf("monthly repeat needs a day, e.g. --repeat monthly:15")
		}
		d, e := strconv.Atoi(parts[1])
		if e != nil {
			return "", 0, "", nil, fmt.Errorf("invalid day of month %q", parts[1])
		}
		return freq, interval, "", &d, nil
	default:
		return "", 0, "", nil, fmt.Errorf("unknown repeat freq %q (want daily, weekly:D1,D2, or monthly:DD)", freq)
	}
}

func lookupCategoryID(name string) (int64, error) {
	body, err := httpGet("/api/categories")
	if err != nil {
		return 0, err
	}
	var cats []models.Category
	if err := json.Unmarshal(body, &cats); err != nil {
		return 0, err
	}
	for _, c := range cats {
		if strings.EqualFold(c.Name, name) {
			return c.ID, nil
		}
	}
	return 0, fmt.Errorf("no category named %q", name)
}

// --- gtzy list [--today] [--wofi] [--json] ---

func newListCmd() *cobra.Command {
	var todayOnly, wofi, asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `Lists tasks. By default prints a human-readable table of all tasks.

  --today  only tasks scheduled for today
  --wofi   tab-separated "<id>\t<title> · <elapsed>" lines for piping into wofi --dmenu (done tasks excluded)
  --json   raw JSON response from the server, for scripting`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(todayOnly, wofi, asJSON)
		},
	}
	cmd.Flags().BoolVar(&todayOnly, "today", false, "only tasks scheduled for today")
	cmd.Flags().BoolVar(&wofi, "wofi", false, `wofi-friendly output: "<id>\t<title> · <elapsed>"`)
	cmd.Flags().BoolVar(&asJSON, "json", false, "print raw JSON instead of a table")
	return cmd
}

func runList(todayOnly, wofi, asJSON bool) error {
	path := "/api/tasks"
	if todayOnly {
		path += "?date=" + today()
	}
	body, err := httpGet(path)
	if err != nil {
		return err
	}

	if asJSON {
		fmt.Println(string(body))
		return nil
	}

	var tasks []models.Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return err
	}

	if wofi {
		for _, t := range tasks {
			if t.Status == "done" {
				continue
			}
			fmt.Printf("%d\t%s · %s\n", t.ID, t.Title, fmtDuration(t.ElapsedSeconds))
		}
		return nil
	}

	for _, t := range tasks {
		marker := " "
		if t.IsActive {
			marker = "*"
		} else if t.Status == "done" {
			marker = "x"
		}
		fmt.Printf("[%s] #%d %s (%s, %s) — %s\n", marker, t.ID, t.Title, t.Priority, t.Status, fmtDuration(t.ElapsedSeconds))
	}
	return nil
}

// --- gtzy waybar ---

type waybarOutput struct {
	Text       string `json:"text"`
	Tooltip    string `json:"tooltip"`
	Class      string `json:"class"`
	Percentage int    `json:"percentage"`
}

func newWaybarCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "waybar",
		Short: "Print a Waybar custom-module JSON status line",
		Long: `Prints a single-line JSON object {"text","tooltip","class","percentage"}
describing the currently running (or most recently paused) task, suitable
for a Waybar "custom/gtzy" module's "exec". Always exits 0 and always prints
valid JSON, even if the gtzy server is unreachable or something else goes
wrong internally — this command must never break the Waybar bar.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			runWaybar()
			return nil
		},
	}
}

func runWaybar() {
	out := computeWaybarOutput()
	enc, err := json.Marshal(out)
	if err != nil {
		// Never break the waybar module: print a safe idle fallback instead.
		fmt.Println(`{"text":"  idle","tooltip":"gtzy: error building status","class":"idle","percentage":0}`)
		return
	}
	fmt.Println(string(enc))
}

func computeWaybarOutput() waybarOutput {
	body, err := httpGet("/api/timer/current")
	if err == nil {
		var res struct {
			Current *models.Task `json:"current"`
		}
		if json.Unmarshal(body, &res) == nil && res.Current != nil {
			t := res.Current
			return waybarOutput{
				Text:       fmt.Sprintf("  %s · %s", t.Title, fmtDuration(t.ElapsedSeconds)),
				Tooltip:    fmt.Sprintf("%s\nElapsed %s / Est %s", t.Title, fmtDuration(t.ElapsedSeconds), fmtDuration(int64(t.EstimatedSeconds))),
				Class:      "running",
				Percentage: percentage(t.ElapsedSeconds, t.EstimatedSeconds),
			}
		}
	}

	// Nothing actively running — check for a paused task touched today.
	if body, err := httpGet("/api/tasks?date=" + today() + "&status=paused"); err == nil {
		var tasks []models.Task
		if json.Unmarshal(body, &tasks) == nil && len(tasks) > 0 {
			t := mostRecentlyUpdated(tasks)
			return waybarOutput{
				Text:       fmt.Sprintf("  %s · %s", t.Title, fmtDuration(t.ElapsedSeconds)),
				Tooltip:    fmt.Sprintf("%s (paused)\nElapsed %s / Est %s", t.Title, fmtDuration(t.ElapsedSeconds), fmtDuration(int64(t.EstimatedSeconds))),
				Class:      "paused",
				Percentage: percentage(t.ElapsedSeconds, t.EstimatedSeconds),
			}
		}
	}

	return waybarOutput{Text: "  idle", Tooltip: "gtzy: no active task", Class: "idle", Percentage: 0}
}

func mostRecentlyUpdated(tasks []models.Task) models.Task {
	best := tasks[0]
	for _, t := range tasks[1:] {
		if t.UpdatedAt > best.UpdatedAt {
			best = t
		}
	}
	return best
}
