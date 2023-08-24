package event

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

var events = make(map[int64][]Event)

func CreateAction(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, chatId string, db *dynamodb.DynamoDB) {
	newEvent := new(Event)
	newEvent.Id = uuid.NewString()
	newEvent.UserId = chatId

	message(bot, chatId, "What is the name of the event?")

	// Wait for the user to reply with the event name.
	eventNameUpdate := <-updates
	if eventNameUpdate.Message != nil {
		newEvent.Name = eventNameUpdate.Message.Text
	}

	message(bot, chatId, "What is the date and time of the event? (formats: 2006-01-02 15:04, 2006-01-02, +5h)")

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

	addEventToDynamoDB(*newEvent, db)
	message(bot, chatId, fmt.Sprintf("Name: %s, Date: %s.", newEvent.Name, newEvent.DateTime))
}

func DeleteAction(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, chatId string, db *dynamodb.DynamoDB) {
	message(bot, chatId, "Give me ID of the event you want to delete")

	eventId := ""
	// Wait for the user to reply with the event id.
	eventIdUpdate := <-updates
	if eventIdUpdate.Message != nil {
		eventId = eventIdUpdate.Message.Text
	}

	// Create the input for the DeleteItem operation
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Events"),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(eventId),
			},
		},
	}

	// Call the DeleteItem operation
	_, err := db.DeleteItem(input)
	if err != nil {
		message(bot, chatId, fmt.Sprintf("Given ID: %s not found.", eventId))

		fmt.Println("Error deleting item:", err.Error())
		return
	}

	message(bot, chatId, fmt.Sprintf("Item ID: %s deleted successfully.", eventId))
}

func ListAction(bot *tgbotapi.BotAPI, chatId string, db *dynamodb.DynamoDB) {

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("Events"),
		IndexName:              aws.String("UserUdIdx"),
		KeyConditionExpression: aws.String("UserId = :userId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userId": {
				S: aws.String(chatId),
			},
		},
	}

	// execute the query and print the results
	queryOutput, err := db.Query(queryInput)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var events []Event
	// Unmarshal the query results to the Events slice
	for _, item := range queryOutput.Items {
		var event Event
		err := dynamodbattribute.UnmarshalMap(item, &event)
		if err != nil {
			fmt.Println("Error unmarshaling event:", err.Error())
			continue
		}
		events = append(events, event)
	}

	jsn, _ := json.Marshal(events)
	message(bot, chatId, string(jsn))
}

func message(bot *tgbotapi.BotAPI, chatId string, data string) {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 64)
	bot.Send(tgbotapi.NewMessage(chatIdInt, data))
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
