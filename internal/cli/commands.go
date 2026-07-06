package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gtzy/internal/models"
)

func today() string { return time.Now().Format("2006-01-02") }

// --- gtzy next ---

func cmdNext(args []string) error {
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

func cmdStart(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: gtzy start <id>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id %q", args[0])
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

func cmdPause(args []string) error {
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

func cmdComplete(args []string) error {
	var id int64
	if len(args) >= 1 {
		parsed, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q", args[0])
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

func cmdCurrent(args []string) error {
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
	if t.EstimatedMinutes > 0 {
		est = fmt.Sprintf("est %s", fmtDuration(int64(t.EstimatedMinutes)*60))
	}
	fmt.Printf("%s — %s elapsed (%s)\n", t.Title, fmtDuration(t.ElapsedSeconds), est)
	return nil
}

// --- gtzy add "<title>" [--priority] [--category] [--est] [--date] [--at] [--repeat] ---

func cmdAdd(args []string) error {
	if len(args) < 1 || strings.HasPrefix(args[0], "-") {
		return fmt.Errorf(`usage: gtzy add "<title>" [--priority p] [--category name] [--est minutes] [--date YYYY-MM-DD] [--at HH:MM] [--repeat daily|weekly:1,3,5|monthly:15]`)
	}
	title := args[0]

	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	priority := fs.String("priority", "medium", "")
	category := fs.String("category", "", "")
	est := fs.Int("est", 0, "")
	date := fs.String("date", today(), "")
	at := fs.String("at", "", "")
	repeat := fs.String("repeat", "", "")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	var categoryID *int64
	if *category != "" {
		id, err := lookupCategoryID(*category)
		if err != nil {
			return err
		}
		categoryID = &id
	}

	var scheduledStart *string
	if *at != "" {
		scheduledStart = at
	}

	if *repeat != "" {
		freq, interval, daysOfWeek, dayOfMonth, err := parseRepeat(*repeat)
		if err != nil {
			return err
		}
		payload := map[string]any{
			"title": title, "priority": *priority, "estimated_minutes": *est,
			"scheduled_start": scheduledStart, "freq": freq, "interval": interval,
			"days_of_week": daysOfWeek, "day_of_month": dayOfMonth, "start_date": *date,
		}
		if categoryID != nil {
			payload["category_id"] = *categoryID
		}
		if _, err := httpPost("/api/recurrences", payload); err != nil {
			return err
		}
		fmt.Printf("created recurring rule: %s (%s)\n", title, *repeat)
		return nil
	}

	payload := map[string]any{
		"title": title, "priority": *priority, "estimated_minutes": *est,
		"scheduled_date": *date, "scheduled_start": scheduledStart,
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

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	todayOnly := fs.Bool("today", false, "")
	wofi := fs.Bool("wofi", false, "")
	asJSON := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := "/api/tasks"
	if *todayOnly {
		path += "?date=" + today()
	}
	body, err := httpGet(path)
	if err != nil {
		return err
	}

	if *asJSON {
		fmt.Println(string(body))
		return nil
	}

	var tasks []models.Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return err
	}

	if *wofi {
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

func cmdWaybar(args []string) error {
	out := computeWaybarOutput()
	enc, err := json.Marshal(out)
	if err != nil {
		// Never break the waybar module: print a safe idle fallback instead.
		fmt.Println(`{"text":"  idle","tooltip":"gtzy: error building status","class":"idle","percentage":0}`)
		return nil
	}
	fmt.Println(string(enc))
	return nil
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
				Tooltip:    fmt.Sprintf("%s\nElapsed %s / Est %s", t.Title, fmtDuration(t.ElapsedSeconds), fmtDuration(int64(t.EstimatedMinutes)*60)),
				Class:      "running",
				Percentage: percentage(t.ElapsedSeconds, t.EstimatedMinutes),
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
				Tooltip:    fmt.Sprintf("%s (paused)\nElapsed %s / Est %s", t.Title, fmtDuration(t.ElapsedSeconds), fmtDuration(int64(t.EstimatedMinutes)*60)),
				Class:      "paused",
				Percentage: percentage(t.ElapsedSeconds, t.EstimatedMinutes),
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
