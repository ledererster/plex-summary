package main

import (
	"flag"
	"log"
	"time"
)

var shouldRunOnce = flag.Bool("run-once", false, "Run summary and exit")
var runDate = flag.String("date", "", "Run summary for a specific date (YYYY-MM-DD)")

func main() {
	flag.Parse()
	LoadConfig()

	if *shouldRunOnce {
		runOnce(*runDate)
		return
	}

	StartScheduler()
	StartTelegramBot()
}

func runOnce(dateArg string) {
	if dateArg == "" {
		dateArg = time.Now().AddDate(0, 0, -1).Format(dateLayout)
	}
	history, err := fetchAllHistoryForDate(dateArg)
	if err != nil {
		log.Fatal("Fetch error:", err)
	}
	summary := generateSummary(history, false)
	if err := sendToGotify("📅 Plex summary", summary); err != nil {
		log.Fatal("Gotify error:", err)
	}
}
