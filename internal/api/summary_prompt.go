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
		stats.EstimatedSecondsTotal/60, int(stats.ActualSecondsTotal/60), stats.CurrentStreak, stats.BusiestCategory)

	b.WriteString("Tasks:\n")
	for _, t := range tasks {
		status := "missed"
		switch t.Status {
		case "done", "todo", "paused", "in_progress":
			status = t.Status
		}
		fmt.Fprintf(&b, "- [%s] %s (priority: %s, est: %ds, actual: %ds)\n", status, t.Title, t.Priority, t.EstimatedSeconds, t.ActualSeconds)
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

func buildJournalSummaryPrompt(from, to string, entries []models.JournalEntry) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are a reflective journaling coach analyzing journal entries from %s to %s.\n\n", from, to)

	b.WriteString("Journal entries:\n")
	for _, e := range entries {
		mood := "n/a"
		if e.Mood != nil {
			mood = *e.Mood
		}
		fmt.Fprintf(&b, "- %s (mood: %s): %s\n", e.Date, mood, truncate(e.Content, 500))
	}

	b.WriteString("\nWrite a concise, reflective markdown summary covering: recurring themes, mood patterns over the range, " +
		"and 2-3 gentle observations or suggestions. Keep it warm but direct, under 400 words.")

	return b.String()
}

func buildBloodSugarSummaryPrompt(from, to string, readings []models.BloodSugarReading, st bloodSugarStats) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are a knowledgeable, cautious diabetes educator reviewing blood-glucose readings from %s to %s. "+
		"You are not the patient's doctor and must not give definitive medical advice or dosing changes.\n\n", from, to)

	fmt.Fprintf(&b, "Aggregate stats (mg/dL): %d readings, mean %.0f, range %.0f-%.0f, std dev %.0f, "+
		"estimated A1C %.1f%%, time-in-range (70-180) %.0f%%, low (<70) %.0f%%, high (>180) %.0f%%.\n",
		st.Count, st.Mean, st.Min, st.Max, st.StdDev, st.EstA1C, st.InRangePct, st.LowPct, st.HighPct)

	if len(st.MealTagMeans) > 0 {
		b.WriteString("Mean by tag:")
		for tag, mean := range st.MealTagMeans {
			fmt.Fprintf(&b, " %s %.0f;", tag, mean)
		}
		b.WriteString("\n")
	}

	b.WriteString("\nReadings:\n")
	for _, r := range readings {
		tag := r.MealTag
		if tag == "" {
			tag = "untagged"
		}
		fmt.Fprintf(&b, "- %s: %.0f mg/dL (%s)", r.TakenAt, r.ValueMgdl, tag)
		if r.Notes != "" {
			fmt.Fprintf(&b, " — %s", truncate(r.Notes, 120))
		}
		b.WriteString("\n")
	}

	b.WriteString("\nWrite a concise markdown summary covering: overall control (mean, estimated A1C, time-in-range), " +
		"notable patterns (fasting/pre-meal vs post-meal spikes, dawn phenomenon, lows), and 2-3 gentle, general " +
		"lifestyle-oriented observations. Flag any concerning lows or highs plainly. Keep it warm but direct, under 400 words. " +
		"End with a one-line disclaimer that this is informational only and not a substitute for advice from the user's care team.")

	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
