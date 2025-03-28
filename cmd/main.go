package main

import (
	"log/slog"
	"bytes"
	"flag"
	"image/png"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/skip2/go-qrcode"

	"tg_bot/internal/config"
	"tg_bot/internal/externalapi"
)

var (
	links        = make(map[string]string)
	linksLock    sync.Mutex
	awaitingURL  = make(map[int64]string)
	awaitingLock sync.Mutex
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", ".env", "configuration file")
	flag.Parse()

	cfg := config.MustLoadCfg(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting bot")

	log.Debug("debug messages are enabled")

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Error("error starting bot", "error", err)
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
				msg := tgbotapi.NewMessage(chatID, "Please send the URL to shorten in https://.. format")
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
					shortURL, err := externalapi.ShortenURL(text, cfg.BitlyToken)
					if err != nil {
						log.Error("Error generating short URL:", "error", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to generate short URL. Please, try again. Make sure it is in https://.. format")
						bot.Send(msg)
						continue
					}
					msg := tgbotapi.NewMessage(chatID, "Shortened URL: "+shortURL)
					bot.Send(msg)
				case "qrcode":
					var buf bytes.Buffer
					qr, err := qrcode.New(text, qrcode.Medium)
					if err != nil {
						log.Error("Error generating QR-code:", "error", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to generate QR-code. Please, try again.")
						bot.Send(msg)
						continue
					}

					err = png.Encode(&buf, qr.Image(256))
					if err != nil {
						log.Error("Error encoding QR-code to PNG", "error", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to encode QR-code. Please, try again.")
						bot.Send(msg)
						continue
					}

					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "qrcode.png", Bytes: buf.Bytes()})
					if _, err = bot.Send(photo); err != nil {
						log.Error("Error sending photo:", "error", err)
						msg := tgbotapi.NewMessage(chatID, "Failed to send QR-code. Please, try again.")
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

func mustMakeLogger(logLevel string) *slog.Logger {
	return slog.Default()
}
