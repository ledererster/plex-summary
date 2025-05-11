package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TautulliURL string
	APIKey      string
	GotifyURL   string
	GotifyToken string
}

var config Config

var (
	startDate  = flag.String("start", "", "Exact date")
	afterDate  = flag.String("after", "", "After date")
	beforeDate = flag.String("before", "", "Before date")
	today      = flag.Bool("today", false, "Summary for today")
	yesterday  = flag.Bool("yesterday", false, "Summary for yesterday")
	lastWeek   = flag.Bool("last-week", false, "Summary for the last 7 days")
	allTime    = flag.Bool("all", false, "Fetch all history (ignores date filters)")
	dryRun     = flag.Bool("dry-run", false, "Don't send to Gotify")
)

const dateLayout = "2006-01-02"

func init() {
	_ = godotenv.Load()

	config = Config{
		TautulliURL: os.Getenv("TAUTULLI_URL"),
		APIKey:      os.Getenv("TAUTULLI_API_KEY"),
		GotifyURL:   os.Getenv("GOTIFY_URL"),
		GotifyToken: os.Getenv("GOTIFY_TOKEN"),
	}

	if config.TautulliURL == "" || config.APIKey == "" || config.GotifyURL == "" || config.GotifyToken == "" {
		log.Fatal("Missing one or more required environment variables.")
	}
}

type HistoryItem struct {
	Username          string  `json:"user"`
	Title             string  `json:"full_title"`
	MediaType         string  `json:"media_type"`
	Date              int64   `json:"date"`
	Platform          string  `json:"platform"`
	Player            string  `json:"player"`
	Product           string  `json:"product"`
	IPAddress         string  `json:"ip_address"`
	TranscodeDecision string  `json:"transcode_decision"`
	Duration          int     `json:"duration"`
	WatchedStatus     float64 `json:"watched_status"`
	Episode           FlexInt `json:"media_index"`
	Season            FlexInt `json:"parent_media_index"`
	Live              int     `json:"live"`
	GrandparentTitle  string  `json:"grandparent_title"`
}

type HistoryData struct {
	History       []HistoryItem `json:"data"`
	TotalDuration string        `json:"filter_duration"`
	TotalRecords  int           `json:"recordsFiltered"`
}

type HistoryResponse struct {
	Response struct {
		Data HistoryData `json:"data"`
	} `json:"response"`
}

type FlexInt int

func (i *FlexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*i = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = FlexInt(n)
	return nil
}

func validateDateFormat(dateStr string) {
	_, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		log.Fatalf("Invalid date format: %s (expected YYYY-MM-DD)", dateStr)
	}
}

func buildHistoryUrl(start int) string {

	params := make([]string, 0)

	if !*allTime {
		// Default to yesterday if no date filters are set
		if *startDate == "" && *afterDate == "" && *beforeDate == "" && !*today && !*lastWeek {
			*yesterday = true
		}
		now := time.Now()
		if *today {
			*startDate = now.Format(dateLayout)
		}
		if *yesterday {
			*startDate = now.AddDate(0, 0, -1).Format(dateLayout)
		}
		if *lastWeek {
			*afterDate = now.AddDate(0, 0, -7).Format(dateLayout)
		}
		if *startDate != "" {
			validateDateFormat(*startDate)
			params = append(params, "start_date="+*startDate)
		}
		if *afterDate != "" {
			validateDateFormat(*afterDate)
			params = append(params, "after="+*afterDate)
		}
		if *beforeDate != "" {
			validateDateFormat(*beforeDate)
			params = append(params, "before="+*beforeDate)
		}
	}

	// Add fixed length param
	params = append(params, "length=100")

	//add start
	params = append(params, "start="+strconv.Itoa(start))

	return fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_history&%s",
		config.TautulliURL,
		config.APIKey,
		strings.Join(params, "&"),
	)
}

func main() {
	flag.Parse()

	history, err := fetchAllHistory()
	if err != nil {
		log.Fatal("Failed to fetch history:", err)
	}

	summary := generateSummary(history)

	var title string

	switch {
	case *startDate != "":
		title = fmt.Sprintf("📅 Plex activity for %s\n\n", *startDate)
	case *afterDate != "" && *beforeDate != "":
		title = fmt.Sprintf("📅 Plex activity from %s to %s\n\n", *afterDate, *beforeDate)
	case *afterDate != "":
		title = fmt.Sprintf("📅 Plex activity since %s\n\n", *afterDate)
	case *beforeDate != "":
		title = fmt.Sprintf("📅 Plex activity until %s\n\n", *beforeDate)
	default:
		title = "📅 Plex activity summary\n\n"
	}
	if *dryRun {
		fmt.Print(title)
		fmt.Print(summary)
	}
	if !*dryRun {
		if err := sendToGotify(title, summary); err != nil {
			log.Fatal("Gotify send failed:", err)
		}
	}
}

func fetchHistory(url string) (*HistoryData, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Response.Data, nil
}

func fetchAllHistory() (*HistoryData, error) {
	var allItems []HistoryItem
	var totalDuration string
	var totalRecords int

	for start := 0; ; start += 100 {
		data, err := fetchHistory(buildHistoryUrl(start))
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, data.History...)
		if totalDuration == "" {
			totalDuration = data.TotalDuration
			totalRecords = data.TotalRecords
		}
		if start >= totalRecords || len(data.History) == 0 {
			break
		}
	}

	return &HistoryData{
		History:       allItems,
		TotalDuration: totalDuration,
		TotalRecords:  totalRecords,
	}, nil
}

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

func sendToGotify(title, message string) error {
	payload := map[string]interface{}{
		"title":    title,
		"message":  message,
		"priority": 5,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config.GotifyURL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", config.GotifyToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
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

	return fmt.Sprintf("  %s %s%s for ~%d min [%s] @ %s (%s)\n",
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
