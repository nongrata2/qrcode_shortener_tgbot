package externalapi

import (
	"encoding/json"
	"net/http"
	"bytes"
	"fmt"
)

type BitlyResponse struct {
	Link string `json:"link"`
}

func ShortenURL(longURL string, accessToken string) (string, error) {
	apiURL := "https://api-ssl.bitly.com/v4/shorten"

	requestBody, err := json.Marshal(map[string]string{
		"long_url": longURL,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return "", fmt.Errorf("Failed to shorten URL, status code: %d, response: %s", resp.StatusCode, body.String())
	}

	var bitlyResponse BitlyResponse
	err = json.NewDecoder(resp.Body).Decode(&bitlyResponse)
	if err != nil {
		return "", err
	}

	return bitlyResponse.Link, nil
}