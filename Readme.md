# Plex Summary Bot

A **Go-based Telegram + Gotify bot** for daily Plex media usage summaries.

It connects **Tautulli**, **Telegram**, and **Gotify** to deliver activity summaries, live session info, and automated daily reports.

---

## Features

- **Telegram Bot Integration**: Interact with the bot via Telegram commands to receive activity summaries directly in your chat.
- **Gotify API Integration**: Sends daily media summaries and live notifications to your Gotify instance.
- **Flexible Date Queries**: Retrieve Plex usage data for specific dates, date ranges, or even all-time.
- **Daily Scheduler**:
   - Automatically fetches and summarizes data from Tautulli daily.
   - Sends the summary to Gotify at a configurable time using cron syntax.
- **Active User Sessions**: Query active Plex sessions and display live-streaming information.
- **Pagination Support**: Efficiently fetches large datasets using incremental pagination.
- **Access Control**: Restrict Telegram bot commands to allowed Telegram User IDs using the environment configuration.

---

## Getting Started

### Prerequisites

- Installed & running **Go (v1.24 or later)**.
- A configured Tautulli instance (required for API integration).
- A Gotify instance (optional but recommended for push notifications).
- A Telegram bot set up through the [BotFather](https://core.telegram.org/bots#botfather).

---

### Configuration

Set your environment variables in a `.env` file or pass them directly as environment variables. Below is a list of the variables supported:

| Variable                  | Description                                                            | Example Value                     |
|---------------------------|------------------------------------------------------------------------|-----------------------------------|
| `TAUTULLI_URL`            | URL of your Tautulli server, including protocol and port if relevant. | `http://localhost:8181`          |
| `TAUTULLI_API_KEY`        | API key for Tautulli for accessing its API via this bot.              | `YOUR_SECRET_API_KEY`             |
| `GOTIFY_URL`              | Gotify server URL. Optional but required for sending Gotify notifications. | `http://gotify.example.com`       |
| `GOTIFY_TOKEN`            | Gotify token for authenticating API requests.                        | `YOUR_GOTIFY_TOKEN`               |
| `TELEGRAM_TOKEN`          | Telegram Bot Token generated by BotFather.                           | `123456789:ABCDEFYOURTOKEN`       |
| `TELEGRAM_ALLOWED_USERS`  | Comma-separated list of allowed Telegram user IDs.                   | `123456789,987654321`             |
| `DAILY_SUMMARY_SCHEDULE`  | Cron syntax defining when the daily summary is sent. Optional.        | `0 8 * * *` (8:00 AM daily)       |

---

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/plex-summary-bot.git
   cd plex-summary-bot
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up your `.env` file:
   Create an `.env` file in the project root directory and configure all required environment variables.

4. Build the project:
   ```bash
   go build -o plex-summary-bot
   ```

5. Run the bot:
   ```bash
   ./plex-summary-bot
   ```

---

### Usage

#### Run in Default Mode
Start the Telegram bot and scheduler together:
```bash
./plex-summary-bot
```


#### Run Once and Exit
Generate a summary for a specific date or use the default (yesterday's summary):
```bash
./plex-summary-bot -run-once -date YYYY-MM-DD
```


#### Key Telegram Bot Commands

| Command                              | Description                                        |
|--------------------------------------|----------------------------------------------------|
| `/start`                             | Starts the bot and provides instructions.          |
| `/today`                             | Fetch today's activity summary.                   |
| `/yesterday`                         | Fetch yesterday's activity summary.               |
| `/lastweek`                          | Fetch the summary for the last 7 days.            |
| `/range YYYY-MM-DD YYYY-MM-DD`       | Fetch the summary for a custom date range.        |
| `/all`                               | Fetch the summary for all time.                   |
| `/active`                            | Display active Plex sessions in real-time.        |

---

### Examples

#### Cron-Scheduled Daily Summary
Add the following environment variable in `.env` to schedule a daily summary at 8:00 AM:
```dotenv
DAILY_SUMMARY_SCHEDULE=0 8 * * *
```

The bot will automatically send a summary through Gotify daily using this schedule.

#### Allowed Telegram Users
Restrict Telegram bot access to specific user IDs:
```dotenv
TELEGRAM_ALLOWED_USERS=123456789,987654321
```

These IDs must match the user IDs of the Telegram accounts interacting with the bot.

---

### License

This project is licensed under the MIT License. See the LICENSE file for details.
