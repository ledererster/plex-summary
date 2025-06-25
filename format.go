package main

import (
	"fmt"
	"strings"
	"time"
)

func generateAggregatedSummary(data *HistoryData) string {
	type ShowGroup struct {
		Episodes int
		Duration int
	}
	type MovieGroup struct {
		Count    int
		Duration int
	}
	type UserSummary struct {
		Movies map[string]*MovieGroup
		Shows  map[string]*ShowGroup
		Total  int
	}

	// Global live TV data
	liveByShow := make(map[string]int)
	totalLive := 0

	users := make(map[string]*UserSummary)

	for _, item := range data.History {
		if item.Live == 1 {
			title := item.GrandparentTitle
			if title == "" {
				title = item.Title
			}
			liveByShow[title] += item.Duration
			totalLive += item.Duration
			continue
		}

		user := item.Username
		if _, ok := users[user]; !ok {
			users[user] = &UserSummary{
				Movies: make(map[string]*MovieGroup),
				Shows:  make(map[string]*ShowGroup),
			}
		}
		summary := users[user]
		summary.Total += item.Duration

		switch item.MediaType {
		case "movie":
			group := summary.Movies[item.Title]
			if group == nil {
				group = &MovieGroup{}
				summary.Movies[item.Title] = group
			}
			group.Count++
			group.Duration += item.Duration

		case "episode":
			title := item.GrandparentTitle
			if title == "" {
				title = item.Title
			}
			group := summary.Shows[title]
			if group == nil {
				group = &ShowGroup{}
				summary.Shows[title] = group
			}
			group.Episodes++
			group.Duration += item.Duration
		}
	}

	var b strings.Builder

	// 🔊 Global live TV section
	if totalLive > 0 {
		b.WriteString(fmt.Sprintf("📡 You watched %s of Live TV\n", formatDuration(totalLive)))
		for show, dur := range liveByShow {
			b.WriteString(fmt.Sprintf("  - %s: %s\n", show, formatDuration(dur)))
		}
		b.WriteString("\n")
	}

	// 👤 Per-user summaries
	for user, summary := range users {
		b.WriteString(fmt.Sprintf("👤 %s\n", user))

		if len(summary.Movies) > 0 {
			var dur int
			b.WriteString(fmt.Sprintf("🎬 Movies (%d titles):\n", len(summary.Movies)))
			for title, g := range summary.Movies {
				b.WriteString(fmt.Sprintf("  - %s (%dx)\n", title, g.Count))
				dur += g.Duration
			}
			b.WriteString(fmt.Sprintf("  Total movie time: %s\n", formatDuration(dur)))
		}

		if len(summary.Shows) > 0 {
			var dur, eps int
			b.WriteString(fmt.Sprintf("📺 Shows (%d titles):\n", len(summary.Shows)))
			for title, g := range summary.Shows {
				b.WriteString(fmt.Sprintf("  - %s (%d eps)\n", title, g.Episodes))
				dur += g.Duration
				eps += g.Episodes
			}
			b.WriteString(fmt.Sprintf("  Total: %d episodes — %s\n", eps, formatDuration(dur)))
		}

		b.WriteString(fmt.Sprintf("🕒 Total watched: %s\n\n", formatDuration(summary.Total)))
	}

	b.WriteString(fmt.Sprintf("📊 Grand total duration: %s\n", data.TotalDuration))
	return b.String()
}

func generateSummary(data *HistoryData, compressed bool) string {
	if compressed {
		return generateAggregatedSummary(data)
	}
	items := data.History
	userSummaries := make(map[string][]string)
	userDurations := make(map[string]int)
	liveByShow := make(map[string]int)
	totalLive := 0

	for _, item := range items {
		if item.Live == 1 {
			show := item.GrandparentTitle
			if show == "" {
				show = item.Title
			}
			liveByShow[show] += item.Duration
			totalLive += item.Duration
			continue
		}
		summary := FormatSummary(item)
		userSummaries[item.Username] = append(userSummaries[item.Username], summary)
		userDurations[item.Username] += item.Duration
	}

	var builder strings.Builder

	if totalLive > 0 {
		builder.WriteString(fmt.Sprintf("📡 You watched %s of Live TV\n", formatDuration(totalLive)))
		for show, dur := range liveByShow {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", show, formatDuration(dur)))
		}
		builder.WriteString("\n")
	}

	for user, lines := range userSummaries {
		totalDur := formatDuration(userDurations[user])
		builder.WriteString(fmt.Sprintf("%s (%s):\n", user, totalDur))
		for _, line := range lines {
			builder.WriteString(line)
		}
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("🕒 Total duration: %s\n", data.TotalDuration))

	return builder.String()
}

func FormatSummary(item HistoryItem) string {
	t := time.Unix(item.Date, 0).Local().Format("15:04:05")

	var status string
	switch {
	case item.WatchedStatus >= 0.95:
		status = "Complete"
	case item.WatchedStatus <= 0.05:
		status = "Unwatched"
	default:
		status = fmt.Sprintf("%d%% Watched", int(item.WatchedStatus*100))
	}

	minutes := item.Duration / 60

	epInfo := ""
	if item.MediaType == "episode" && item.Season != 0 && item.Episode != 0 {
		epInfo = fmt.Sprintf(" S%dE%d", item.Season, item.Episode)
	}

	prefix := "  "
	if item.State == "playing" {
		prefix = "▶️ "
	}

	return fmt.Sprintf("%s%s %s%s for ~%d min [%s] @ %s (%s)\n",
		prefix,
		MediaIcon(item),
		item.Title,
		epInfo,
		minutes,
		status,
		t,
		item.Player,
	)
}

func MediaIcon(item HistoryItem) string {
	switch {
	case item.Live == 1:
		return "📡"
	case item.MediaType == "episode":
		return "📺"
	case item.MediaType == "movie":
		return "🎬"
	default:
		return "❓"
	}
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
