package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"time"
)

func StartTelegramBot() {
	bot, err := tgbotapi.NewBotAPI(AppConfig.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{Command: "today", Description: "Summary for today"},
		{Command: "yesterday", Description: "Summary for yesterday"},
		{Command: "lastweek", Description: "Summary for the last 7 days"},
		{Command: "range", Description: "Summary for date range: /range YYYY-MM-DD YYYY-MM-DD"},
		{Command: "all", Description: "Summary for all time"},
		{Command: "active", Description: "Show current Plex sessions"},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	_, err = bot.Request(cfg)
	if err != nil {
		log.Fatal("Failed to set bot commands: ", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		// Restrict to allowed users if list is non-empty
		if len(AppConfig.AllowedTelegramIDs) > 0 &&
			!AppConfig.AllowedTelegramIDs[update.Message.From.ID] {

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Access denied."))
			continue
		}

		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {

		case "start":
			log.Printf("Received message from user ID: %d (%s)", update.Message.From.ID, update.Message.From.UserName)
			msg := fmt.Sprintf(`Hello! Your Telegram ID is %d.

				Available commands:
				/today - Summary for today
				/yesterday - Summary for yesterday
				/lastweek - Summary for last 7 days
				/range YYYY-MM-DD YYYY-MM-DD - Summary for custom date range
				/all - Summary for all time
				/active - Show current Plex sessions
				`, update.Message.From.ID)

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		case "today":
			date := time.Now().Format(dateLayout)
			opts := HistoryRequest{StartDate: date}
			sendTelegramSummary(bot, update.Message.Chat.ID, opts)

		case "yesterday":
			date := time.Now().AddDate(0, 0, -1).Format(dateLayout)
			opts := HistoryRequest{StartDate: date}
			sendTelegramSummary(bot, update.Message.Chat.ID, opts)

		case "lastweek":
			after := time.Now().AddDate(0, 0, -7).Format(dateLayout)
			opts := HistoryRequest{AfterDate: after}
			sendTelegramSummary(bot, update.Message.Chat.ID, opts)

		case "range":
			dates := strings.Fields(args)
			if len(dates) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /range YYYY-MM-DD YYYY-MM-DD"))
				continue
			}
			opts := HistoryRequest{AfterDate: dates[0], BeforeDate: dates[1]}
			sendTelegramSummary(bot, update.Message.Chat.ID, opts)

		case "all":
			opts := HistoryRequest{AllTime: true}
			sendTelegramSummary(bot, update.Message.Chat.ID, opts)

		case "active":
			text, err := fetchActiveSessions()
			if err != nil {
				text = "Error fetching active sessions: " + err.Error()
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
		}
	}
}

func sendTelegramSummary(bot *tgbotapi.BotAPI, chatID int64, opts HistoryRequest) {
	data, err := fetchAllHistory(opts)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Error: "+err.Error()))
		return
	}

	summary := generateSummary(data)
	chunks := splitMessage(summary, 4096)

	for _, chunk := range chunks {
		bot.Send(tgbotapi.NewMessage(chatID, chunk))
	}
}

func splitMessage(s string, maxLen int) []string {
	var parts []string
	for len(s) > maxLen {
		splitAt := strings.LastIndex(s[:maxLen], "\n")
		if splitAt == -1 {
			splitAt = maxLen
		}
		parts = append(parts, s[:splitAt])
		s = s[splitAt:]
	}
	if len(s) > 0 {
		parts = append(parts, s)
	}
	return parts
}
