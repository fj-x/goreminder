package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateBot() *tgbotapi.BotAPI {
	botToken := "6227206106:AAEsGtMRitg17YSVsSco7cWySFxSqCO0uTc"

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	config := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "/list",
			Description: "All active events",
		},
		tgbotapi.BotCommand{
			Command:     "/create",
			Description: "Create",
		},
		tgbotapi.BotCommand{
			Command:     "/update",
			Description: "Update",
		},
		tgbotapi.BotCommand{
			Command:     "/delete",
			Description: "Delete",
		},
	)

	bot.Request(config)

	return bot
}
