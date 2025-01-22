# Telegram Bot for URL Shortening and QR-code Generation
This Telegram bot allows users to shorten URLs and generate QR-codes. It provides two main commands: /short for shortening URLs and /qrcode for generating QR-codes from URLs.

## Commands

- /start
The /start command gives short information about bot's functionality

- /short
The /short command allows users to shorten a given URL. After sending this command, send the URL you want to shorten. The bot will then return a shortened URL using the Bitly API.

- /qrcode
The /qrcode command allows users to generate a QR-code from a given URL. After sending this command, send the URL they want to convert into a QR-code. The bot will then generate the QR-code and send it as an image.

## Setup

1. Clone the repository:

git clone https://github.com/nongrata2/qrcode_shortener_tgbot
cd qrcode_shortener_tgbot

2. Install dependencies:

go get -u github.com/go-telegram-bot-api/telegram-bot-api/v5
go get -u github.com/skip2/go-qrcode
go get -u github.com/joho/godotenv

3. Create a .env file:

Create a file named .env in the root directory of the project.
Add your Bitly access token and Telegram bot token to the .env file:

BITLY_TOKEN=your_bitly_token
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

4. Run the bot:

go run main.go