package main

import (
	"fmt"

	"github.com/fj-x/goreminder/db"
	"github.com/fj-x/goreminder/event"
	"github.com/fj-x/goreminder/telegram"

	"strconv"
)

func main() {

	db := db.Connect()

	bot := telegram.CreateBot()
	updates := bot.ReadChannel()

	eventRepository := event.NewEventRepository(db)
	eventService := event.NewEventService(eventRepository, bot)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		fmt.Println(update.Message.Text)

		// Check if the user sent a command to create a new event.
		if update.Message.IsCommand() {
			chatId := strconv.FormatInt(update.Message.Chat.ID, 10)

			if update.Message.Command() == "create" {
				eventService.CreateAction(updates, chatId)
			}
			if update.Message.Command() == "list" {
				eventService.ListAction(chatId)
			}
			if update.Message.Command() == "delete" {
				eventService.DeleteAction(updates, chatId)
			}
		}
	}
}
