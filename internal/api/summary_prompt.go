package api

import (
	"fmt"
	"strings"

	"gtzy/internal/models"
)

func buildSummaryPrompt(periodType, periodKey string, stats models.Stats, tasks []models.Task, entries []models.JournalEntry) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are a personal growth coach analyzing a %s of productivity data (%s).\n\n", periodType, periodKey)
	fmt.Fprintf(&b, "Stats: %d/%d tasks completed (%.0f%%), %d min estimated vs %d min actual, current streak %d days, busiest category: %s.\n\n",
		stats.TasksCompleted, stats.TasksTotal, stats.CompletionRate*100,
		stats.EstimatedMinutesTotal, int(stats.ActualSecondsTotal/60), stats.CurrentStreak, stats.BusiestCategory)

	b.WriteString("Tasks:\n")
	for _, t := range tasks {
		status := "missed"
		switch t.Status {
		case "done", "todo", "paused", "in_progress":
			status = t.Status
		}
		fmt.Fprintf(&b, "- [%s] %s (priority: %s, est: %dm, actual: %ds)\n", status, t.Title, t.Priority, t.EstimatedMinutes, t.ActualSeconds)
	}

	if len(entries) > 0 {
		b.WriteString("\nJournal entries:\n")
		for _, e := range entries {
			mood := "n/a"
			if e.Mood != nil {
				mood = *e.Mood
			}
			fmt.Fprintf(&b, "- %s (mood: %s): %s\n", e.Date, mood, truncate(e.Content, 500))
		}
	}

	b.WriteString("\nWrite a concise markdown growth summary covering: what went well, what was missed and why it might have happened, " +
		"estimation accuracy (est vs actual), and 2-3 concrete, specific suggestions for improvement. Keep it warm but direct, under 400 words.")

	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
