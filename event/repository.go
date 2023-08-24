package event

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/fj-x/goreminder/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type eventRepository struct {
	db *db.Connection
}

func NewEventRepository(db *db.Connection) *eventRepository {
	return &eventRepository{db: db}
}

func (repo *eventRepository) AddEventToDynamoDB(event Event) {
	av, err := dynamodbattribute.MarshalMap(event)
	if err != nil {
		log.Printf("Got error marshalling event: %s", err)
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Events"),
	}

	_, err = repo.db.Client.PutItem(input)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
	}
}

func (repo *eventRepository) GetAllByChat(chatId string) ([]byte, error) {
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
	queryOutput, err := repo.db.Client.Query(queryInput)
	if err != nil {
		fmt.Println(err.Error())
		// return /// Error management
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

	return json.Marshal(events)
}

func (repo *eventRepository) Delete(eventId string) error {
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
	_, err := repo.db.Client.DeleteItem(input)

	return err
}
