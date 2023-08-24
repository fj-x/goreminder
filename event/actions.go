package event

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fj-x/goreminder/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

var events = make(map[int64][]Event)

type EventService struct {
	eventRepo *eventRepository
	bot       *telegram.TelegramBot
}

func NewEventService(eventRepo *eventRepository, bot *telegram.TelegramBot) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		bot:       bot,
	}
}

func (service EventService) CreateAction(updates tgbotapi.UpdatesChannel, chatId string) {
	newEvent := new(Event)
	newEvent.Id = uuid.NewString()
	newEvent.UserId = chatId

	message(service.bot, chatId, "What is the name of the event?")

	// Wait for the user to reply with the event name.
	eventNameUpdate := <-updates
	if eventNameUpdate.Message != nil {
		newEvent.Name = eventNameUpdate.Message.Text
	}

	message(service.bot, chatId, "What is the date and time of the event? (formats: 2006-01-02 15:04, 2006-01-02, +5h)")

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

	service.eventRepo.AddEventToDynamoDB(*newEvent)

	message(service.bot, chatId, fmt.Sprintf("Name: %s, Date: %s.", newEvent.Name, newEvent.DateTime))
}

func (service EventService) DeleteAction(updates tgbotapi.UpdatesChannel, chatId string) {
	message(service.bot, chatId, "Give me ID of the event you want to delete")

	eventId := ""
	// Wait for the user to reply with the event id.
	eventIdUpdate := <-updates
	if eventIdUpdate.Message != nil {
		eventId = eventIdUpdate.Message.Text
	}

	err := service.eventRepo.Delete(eventId)
	if err != nil {
		message(service.bot, chatId, fmt.Sprintf("Given ID: %s not found.", eventId))

		fmt.Println("Error deleting item:", err.Error())
		return
	}

	message(service.bot, chatId, fmt.Sprintf("Item ID: %s deleted successfully.", eventId))
}

func (service EventService) ListAction(chatId string) {
	data, _ := service.eventRepo.GetAllByChat(chatId)

	message(service.bot, chatId, string(data))
}

func message(bot *telegram.TelegramBot, chatId string, data string) {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 64)
	bot.Bot.Send(tgbotapi.NewMessage(chatIdInt, data))
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
