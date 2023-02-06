package ddbHandler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

type DDBHandler struct {
	svc           dynamodbiface.DynamoDBAPI
	concertsTable string
	ordersTable   string
}

func New(svc dynamodbiface.DynamoDBAPI, concertsTable string, ordersTable string) DDBHandler {
	d := DDBHandler{
		svc,
		concertsTable,
		ordersTable,
	}
	return d
}

func convertEpochSecsToDateAndTimeStrings(dateTime int64) (date string, timeStamp string) {
	t := time.Unix(dateTime, 0)
	date = t.Format("Mon 2 Jan 2006")
	timeStamp = t.Format("3:04 PM")
	return
}

func validateConcert(c *databaseHandler.Concert) (valid bool) {
	valid = false

	if c.ID != "" && c.Description != "" && c.ImageURL != "" &&
		c.DateTime != nil && *c.DateTime > 0 && c.TotalTickets != nil && *c.TotalTickets > 0 &&
		c.TicketsSold != nil && c.FullPrice > 0.0 && c.ConcessionPrice > 0.0 {
		valid = true
	}
	return
}

func generateOrderReference(size uint8) string {
	charSet := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.Seed(time.Now().UnixNano())
	arr := make([]byte, size, size)
	var i uint8
	for i = 0; i < size; i++ {
		arr[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(arr)
}

// GetOrderDetails takes a reference id and returns an order struct from the orders database, or nil if the order does not exist. Returns error if fails to read from the database
func (d DDBHandler) GetOrderDetails(concertId string, ref string) (order *paymentHandler.Order, err error) {
	order = &paymentHandler.Order{}
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.ordersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ConcertId": {
				S: aws.String(concertId),
			},
			"Reference": {
				S: aws.String(ref),
			},
		},
	})
	if err != nil {
		return
	} else if result.Item == nil {
		err = paymentHandler.ErrOrderDoesNotExist{Message: "Error does not exist"}
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, order)
	if err != nil {
		return
	}
	return
}

func (d DDBHandler) createOrderEntry(order paymentHandler.Order) (err error) {
	av, err := dynamodbattribute.MarshalMap(order)
	if err != nil {
		return
	}

	_, err = d.svc.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(d.ordersTable),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(ConcertId) AND attribute_not_exists(Reference)"),
	})
	return

}

// CreateEntryInOrdersDatabase takes the paymentRequest struct, generates a new payment references, checks for uniqueness and creates an entry in the orders database. Returns an error if it fails at any point
func (d DDBHandler) CreateEntryInOrdersDatabase(order paymentHandler.Order) (err error) {
	for {
		order.Reference = generateOrderReference(4)
		err = d.createOrderEntry(order)
		if _, refAlreadyExists := err.(*dynamodb.ConditionalCheckFailedException); refAlreadyExists {
			// If we happen to generate a Reference that matches an existing one, start loop again (re-generate reference)
			continue
		} else {
			// In all other cases, break from loop.
			// If generateOrderReference was successful, err will be nil, otherwise (e.g. table doesn't exists) it will have a value
			// the final return will return nil or the error
			break
		}
	}
	return
}

// GetConcertFromDatabase retrieves a specific concert from the dynamoDB table
func (d DDBHandler) GetConcertFromDatabase(concertID string) (concert *databaseHandler.Concert, err error) {
	concert = &databaseHandler.Concert{}
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.concertsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(concertID),
			},
		},
	})
	if err != nil {
		return
	} else if result.Item == nil {
		err = databaseHandler.ErrConcertDoesNotExist{Message: "Error does not exist"}
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, concert)
	if err != nil {
		return
	}

	if !validateConcert(concert) {
		err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Invalid concert data for concert %s\n", concertID)}
		return
	}

	epochNow := time.Now().Unix()
	if *concert.DateTime < epochNow {
		err = databaseHandler.ErrConcertInPast{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", concertID)}
		return
	}

	dateStr, timeStr := convertEpochSecsToDateAndTimeStrings(*concert.DateTime)
	concert.Date = dateStr
	concert.Time = timeStr
	concert.DateTime = nil
	concert.AvailableTickets = *concert.TotalTickets - *concert.TicketsSold
	concert.TotalTickets = nil
	concert.TicketsSold = nil

	return
}

// GetConcertsFromDatabase gets all upcoming concerts from the dynamoDB table
func (d DDBHandler) GetConcertsFromDatabase() (concerts []databaseHandler.Concert, err error) {
	epochNow := time.Now().Unix()
	filt := expression.Name("DateTime").GreaterThan(expression.Value(epochNow))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return
	}

	result, err := d.svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(d.concertsTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		return
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &concerts)
	if err != nil {
		return
	}

	for _, v := range concerts {
		if !validateConcert(&v) {
			err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", v.ID)}
			return
		}
	}

	for i := 0; i < len(concerts); i++ {
		dateStr, timeStr := convertEpochSecsToDateAndTimeStrings(*concerts[i].DateTime)
		concerts[i].Date = dateStr
		concerts[i].Time = timeStr
		concerts[i].DateTime = nil
		concerts[i].AvailableTickets = *concerts[i].TotalTickets - *concerts[i].TicketsSold
		concerts[i].TotalTickets = nil
		concerts[i].TicketsSold = nil
	}

	return
}
