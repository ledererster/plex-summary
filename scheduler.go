package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func StartScheduler() {
	if AppConfig.DailySummarySchedule == "" {
		log.Println("No DAILY_SUMMARY_SCHEDULE set — scheduler disabled.")
		return
	}

	c := cron.New()

	_, err := c.AddFunc(AppConfig.DailySummarySchedule, func() {
		date := time.Now().AddDate(0, 0, -1).Format(dateLayout)
		history, err := fetchAllHistoryForDate(date)
		if err != nil {
			log.Println("Scheduler error:", err)
			return
		}
		summary := generateSummary(history, false)
		if err := sendToGotify("📅 Daily Plex Summary", summary); err != nil {
			log.Println("Gotify error:", err)
		}
	})
	if err != nil {
		log.Fatalf("Invalid cron schedule: %v", err)
	}

	log.Printf("Scheduler started with schedule: %s", AppConfig.DailySummarySchedule)
	c.Start()
}
