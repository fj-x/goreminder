package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func CreateBot() *TelegramBot {
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

	return &TelegramBot{Bot: bot}
}

func (bot *TelegramBot) ReadChannel() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.Bot.GetUpdatesChan(u)
}
