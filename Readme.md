# Plex Summary

A command-line tool for generating and sending Plex activity summaries using Tautulli data through
Gotify notifications.

## Features

- Generate activity summaries for specific dates or date ranges
- Summary of watched content per user including duration
- Live TV viewing statistics
- Gotify integration for notifications
- Support for movies, TV shows, and live content

## Requirements

- Go 1.24 or higher
- Tautulli server
- Gotify server
- Valid API keys for both services

## Installation

1. Clone the repository
2. Create a `.env` file with the following variables:
   - `TAUTULLI_URL=http://your-tautulli-host:8181`
   - `TAUTULLI_API_KEY=your_tautulli_api_key`
   - `GOTIFY_URL=https://your.gotify.server`
   - `GOTIFY_TOKEN=your_gotify_app_token`
3. Run the tool
   `go run main.go --yesterday`

## Usage
Use one of the provided date filters:

- `--today` – Summary for today
- `--yesterday` – Summary for yesterday
- `--last-week` – Last 7 full days (excluding today)
- `--start=YYYY-MM-DD`
- `--after=YYYY-MM-DD`
- `--before=YYYY-MM-DD`

### Examples:
`go run main.go --today`
`go run main.go --after=2025-05-01 --before=2025-05-07`

## Output
- Each user is listed with total watch time and individual items.
- Live TV is summarized separately.
- Output is sent to Gotify and also printed to stdout.

## License
MIT
