package main

import (
	"log"
	"github.com/skip2/go-qrcode"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"bytes"
	"encoding/json"
	"net/http"
	"fmt"
	"sync"
	"os"
	"image/png"
	"github.com/joho/godotenv"
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
		return "", fmt.Errorf("failed to shorten URL, status code: %d, response: %s", resp.StatusCode, body.String())
	}

	var bitlyResponse BitlyResponse
	err = json.NewDecoder(resp.Body).Decode(&bitlyResponse)
	if err != nil {
		return "", err
	}

	return bitlyResponse.Link, nil
}

var (
	links = make(map[string]string)
	linksLock sync.Mutex
	awaitingURL = make(map[int64]string)
	awaitingLock sync.Mutex
)

func main() {

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

	bitlyToken := os.Getenv("BITLY_TOKEN")
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if bitlyToken == ""{
		log.Fatal("Missing bitly token")
	}

	if telegramBotToken == ""{
		log.Fatal("Missing telegram bot token")
	}

	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(chatID, "Send me a URL to shorten (/short) or to create QR-code (/qrcode)!")
				bot.Send(msg)
			case "short":
				awaitingLock.Lock()
				awaitingURL[chatID] = "short"
				awaitingLock.Unlock()
				msg := tgbotapi.NewMessage(chatID, "Please send the URL to shorten.")
				bot.Send(msg)
			case "qrcode":
				awaitingLock.Lock()
				awaitingURL[chatID] = "qrcode"
				awaitingLock.Unlock()
				msg := tgbotapi.NewMessage(chatID, "Please send the URL to create QR-code.")
				bot.Send(msg)
			}
		} else {
			awaitingLock.Lock()
			command, exists := awaitingURL[chatID]
			awaitingLock.Unlock()

			if exists {
				switch command {
				case "short":
					shortURL, err := ShortenURL(text, bitlyToken)
					if err != nil{
						log.Println("Error generating short URL:", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to generate short URL.")
						bot.Send(msg)
						continue
					}
					msg := tgbotapi.NewMessage(chatID, "Shortened URL: "+shortURL)
					bot.Send(msg)
				case "qrcode":
					var buf bytes.Buffer
					qr, err := qrcode.New(text, qrcode.Medium)
					if err != nil {
						log.Println("Error generating QR code:", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to generate QR code.")
						bot.Send(msg)
						continue
					}

					err = png.Encode(&buf, qr.Image(256))
					if err != nil {
						log.Println("Error encoding QR code to PNG:", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to encode QR code.")
						bot.Send(msg)
						continue
					}

					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "qrcode.png", Bytes: buf.Bytes()})
					if _, err = bot.Send(photo); err != nil {
						log.Println("Error sending photo:", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to send QR code.")
						bot.Send(msg)
					}
				}
				awaitingLock.Lock()
				delete(awaitingURL, chatID)
				awaitingLock.Unlock()
			}
		}
	}
}