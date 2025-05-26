package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TautulliURL          string
	APIKey               string
	GotifyURL            string
	GotifyToken          string
	TelegramBotToken     string
	AllowedTelegramIDs   map[int64]bool
	DailySummarySchedule string
}

var AppConfig Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = Config{
		TautulliURL:          os.Getenv("TAUTULLI_URL"),
		APIKey:               os.Getenv("TAUTULLI_API_KEY"),
		GotifyURL:            os.Getenv("GOTIFY_URL"),
		GotifyToken:          os.Getenv("GOTIFY_TOKEN"),
		TelegramBotToken:     os.Getenv("TELEGRAM_TOKEN"),
		AllowedTelegramIDs:   make(map[int64]bool),
		DailySummarySchedule: os.Getenv("DAILY_SUMMARY_SCHEDULE"),
	}

	idStr := os.Getenv("TELEGRAM_ALLOWED_USERS")
	for _, id := range strings.Split(idStr, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Fatalf("Invalid TELEGRAM_ALLOWED_USERS entry: %s", id)
		}
		AppConfig.AllowedTelegramIDs[parsed] = true
	}

	if AppConfig.TautulliURL == "" || AppConfig.APIKey == "" || AppConfig.GotifyURL == "" || AppConfig.GotifyToken == "" {
		log.Fatal("Missing required environment variables")
	}
}
