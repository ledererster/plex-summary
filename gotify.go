package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func sendToGotify(title, message string) error {
	payload := map[string]interface{}{
		"title":    title,
		"message":  message,
		"priority": 5,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", AppConfig.GotifyURL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", AppConfig.GotifyToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
