package main

import (
	"fmt"
	"strings"
	"time"
)

func generateSummary(data *HistoryData) string {
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
