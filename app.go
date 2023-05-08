package main

import (
	"encoding/json"
	"fmt"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/dynamodb"
	// "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"

	// "os"
	// "strconv"
	"strings"
	"time"
)

type Event struct {
	Id         string
	Name       string
	DateTime   time.Time
	RepeatType string
}

var events = make(map[int64][]Event)

func main() {
	botToken := ""

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
			Description: "Use next pattern: name|date",
		},
		tgbotapi.BotCommand{
			Command:     "/update",
			Description: "Use next pattern: id|name|date",
		},
		tgbotapi.BotCommand{
			Command: "/delete",
		},
	)

	bot.Request(config)

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
			if update.Message.Command() == "create" {
				createAction(bot, updates, update.Message.Chat.ID)
			}
			if update.Message.Command() == "list" {
				listAction(bot, update.Message.Chat.ID)
			}
			if update.Message.Command() == "delete" {
				deleteAction(bot, updates, update.Message.Chat.ID)
			}
		}

	}
}

func createAction(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, chatId int64) {
	newEvent := new(Event)
	newEvent.Id = uuid.NewString()

	// Ask the user for the event name.
	msg := tgbotapi.NewMessage(chatId, "What is the name of the event?")
	bot.Send(msg)

	// Wait for the user to reply with the event name.
	eventNameUpdate := <-updates
	if eventNameUpdate.Message != nil {
		newEvent.Name = eventNameUpdate.Message.Text
	}

	// Ask the user for the event date and time.
	msg = tgbotapi.NewMessage(chatId, "What is the date and time of the event? (formats: 2006-01-02 15:04, 2006-01-02, +5h)")
	bot.Send(msg)

	// Wait for the user to reply with the event date and time.
	eventDateTimeUpdate := <-updates
	if eventDateTimeUpdate.Message != nil {
		eventDateTime, err := parseDate(eventDateTimeUpdate.Message.Text)
		if err != nil {
			fmt.Printf("Error parsing date '%s': %s\n", eventDateTime, err)
		} else {
			newEvent.DateTime = eventDateTime
		}
	}

	events[chatId] = append(events[chatId], *newEvent)

	msg = tgbotapi.NewMessage(chatId, fmt.Sprintf("Name: %s, Date: %s.", newEvent.Name, newEvent.DateTime))
	bot.Send(msg)
}

func deleteAction(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, chatId int64) {
	// Ask the user for the event name.
	msg := tgbotapi.NewMessage(chatId, "Give me ID of the event you want to delete")
	bot.Send(msg)
	eventId := ""
	// Wait for the user to reply with the event id.
	eventIdUpdate := <-updates
	if eventIdUpdate.Message != nil {
		eventId = eventIdUpdate.Message.Text
	}

	userEvents := events[chatId]

	for i, event := range userEvents {
		if event.Id == eventId {
			// Remove the event from the slice
			events[chatId] = append(userEvents[:i], userEvents[i+1:]...)

			bot.Send(tgbotapi.NewMessage(chatId, fmt.Sprintf("Event ID: %s deleted.", eventId)))

			return
		}
	}

	bot.Send(tgbotapi.NewMessage(chatId, fmt.Sprintf("Given ID: %s not found.", eventId)))
}

func listAction(bot *tgbotapi.BotAPI, chatId int64) {
	userEvents := events[chatId]

	b, err := json.Marshal(userEvents)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	msg := tgbotapi.NewMessage(chatId, string(b))
	bot.Send(msg)
}

func parseDate(dateStr string) (time.Time, error) {
	// Check if the date string starts with "+" for relative dates
	if strings.HasPrefix(dateStr, "+") {
		// Parse the duration from the input string
		duration, err := time.ParseDuration(dateStr)
		if err != nil {
			return time.Time{}, err
		}
		// Calculate the future date based on the current time and duration
		future := time.Now().Add(duration)
		return future, nil
	} else {
		// Try parsing the input string as a date
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			date, err := time.Parse(format, dateStr)
			if err == nil {
				return date, nil
			}
		}
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}
}
