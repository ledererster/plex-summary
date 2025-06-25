package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type HistoryRequest struct {
	StartDate  string
	AfterDate  string
	BeforeDate string
	AllTime    bool
	Compressed bool
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
	State             string  `json:"state"`
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

type ActiveResponse struct {
	Response struct {
		Data struct {
			Sessions []struct {
				User             string `json:"username"`
				Title            string `json:"title"`
				GrandparentTitle string `json:"grandparent_title"`
				MediaType        string `json:"media_type"`
				Player           string `json:"player"`
				Platform         string `json:"platform"`
				DurationStr      string `json:"duration"`
				ViewOffsetStr    string `json:"view_offset"`
				SeasonStr        string `json:"parent_media_index"`
				EpisodeStr       string `json:"media_index"`
			} `json:"sessions"`
		} `json:"data"`
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

func fetchAllHistory(opts HistoryRequest) (*HistoryData, error) {
	var allItems []HistoryItem
	var totalDuration time.Duration
	var totalRecords int

	for start := 0; ; start += 100 {
		url := buildHistoryURL(opts, start)
		data, err := fetchHistory(url)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, data.History...)
		if data.TotalDuration != "" {
			dur, err := parseCustomDuration(data.TotalDuration)
			if err == nil {
				totalDuration += dur
			}
		}
		if totalRecords == 0 {
			totalRecords = data.TotalRecords
		}
		if start >= totalRecords || len(data.History) == 0 {
			break
		}
	}

	return &HistoryData{
		History:       allItems,
		TotalDuration: formatCustomDuration(totalDuration),
		TotalRecords:  totalRecords,
	}, nil
}

func buildHistoryURL(opts HistoryRequest, start int) string {
	params := []string{}

	if !opts.AllTime {
		if opts.StartDate != "" {
			validateDateFormat(opts.StartDate)
			params = append(params, "start_date="+opts.StartDate)
		}
		if opts.AfterDate != "" {
			validateDateFormat(opts.AfterDate)
			params = append(params, "after="+opts.AfterDate)
		}
		if opts.BeforeDate != "" {
			validateDateFormat(opts.BeforeDate)
			params = append(params, "before="+opts.BeforeDate)
		}
	}

	params = append(params, "length=100")
	params = append(params, "start="+strconv.Itoa(start))

	return fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_history&%s",
		AppConfig.TautulliURL,
		AppConfig.APIKey,
		strings.Join(params, "&"),
	)
}
func fetchAllHistoryForDate(date string) (*HistoryData, error) {
	if date != "" {
		validateDateFormat(date)
	}

	params := make([]string, 0)
	if date != "" {
		params = append(params, "start_date="+date)
	}

	params = append(params, "length=100")

	var allItems []HistoryItem
	var totalDuration string
	var totalRecords int

	for start := 0; ; start += 100 {
		fullParams := append(params, "start="+strconv.Itoa(start))
		url := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_history&%s",
			AppConfig.TautulliURL,
			AppConfig.APIKey,
			strings.Join(fullParams, "&"),
		)

		data, err := fetchHistory(url)
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

func fetchActiveSessions() (string, error) {
	url := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", AppConfig.TautulliURL, AppConfig.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result ActiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	sessions := result.Response.Data.Sessions
	if len(sessions) == 0 {
		return "No active sessions.", nil
	}

	var b strings.Builder
	for _, s := range sessions {
		duration, _ := strconv.Atoi(s.DurationStr) //get_activity returns it all as strings for some reason
		offset, _ := strconv.Atoi(s.ViewOffsetStr)
		season, _ := strconv.Atoi(s.SeasonStr)
		episode, _ := strconv.Atoi(s.EpisodeStr)

		durationMin := duration / 60000
		var progress float64
		if duration > 0 {
			progress = float64(offset) / float64(duration) * 100
		}

		var title string
		if s.MediaType == "episode" && s.GrandparentTitle != "" {
			title = fmt.Sprintf("%s - %s S%02dE%02d", s.GrandparentTitle, s.Title, season, episode)
		} else {
			title = s.Title
		}

		fmt.Fprintf(&b, "▶️ %s is watching %s on %s [%s] for ~%d min [%.0f%% Watched]\n",
			s.User, title, s.Player, s.Platform, durationMin, progress)
	}
	return b.String(), nil
}
func parseCustomDuration(s string) (time.Duration, error) {
	var total time.Duration
	re := regexp.MustCompile(`(\d+)\s*(days?|hrs?|mins?|secs?)`)
	matches := re.FindAllStringSubmatch(s, -1)

	for _, match := range matches {
		val, _ := strconv.Atoi(match[1])
		switch match[2] {
		case "day", "days":
			total += time.Duration(val) * 24 * time.Hour
		case "hr", "hrs":
			total += time.Duration(val) * time.Hour
		case "min", "mins":
			total += time.Duration(val) * time.Minute
		case "sec", "secs":
			total += time.Duration(val) * time.Second
		}
	}
	return total, nil
}
func formatCustomDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hrs", hours))
	}
	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%d mins", mins))
	}
	if secs > 0 {
		parts = append(parts, fmt.Sprintf("%d secs", secs))
	}
	return strings.Join(parts, " ")
}
