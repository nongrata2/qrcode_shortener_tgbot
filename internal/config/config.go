package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	BitlyToken       string `env:"BITLY_TOKEN"`
	LogLevel         string `env:"LOG_LEVEL" env-default:"DEBUG"`
}

func MustLoadCfg(configPath string) Config {
	if err := godotenv.Load(configPath); err != nil {
		log.Fatalf("failed to load .env file: %s", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read environment variables: %s", err)
	}

	return cfg
}
