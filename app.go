package main

import (
	"fmt"

	"github.com/fj-x/goreminder/db"
	"github.com/fj-x/goreminder/event"
	"github.com/fj-x/goreminder/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"strconv"
)

func main() {

	db := db.Connect()

	bot := telegram.CreateBot()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		fmt.Println(update.Message.Text)

		// Check if the user sent a command to create a new event.
		if update.Message.IsCommand() {
			chatId := strconv.FormatInt(update.Message.Chat.ID, 10)

			if update.Message.Command() == "create" {
				event.CreateAction(bot, updates, chatId, db)
			}
			if update.Message.Command() == "list" {
				event.ListAction(bot, chatId, db)
			}
			if update.Message.Command() == "delete" {
				event.DeleteAction(bot, updates, chatId, db)
			}
		}
	}
}
