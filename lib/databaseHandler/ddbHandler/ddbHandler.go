package ddbHandler

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"os"
	"time"
)

var (
	tableName = os.Getenv("TABLE_NAME")
)

type DDBHandler struct {
	svc dynamodbiface.DynamoDBAPI
}

func New(svc dynamodbiface.DynamoDBAPI) DDBHandler {
	d := DDBHandler{svc}
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
	fmt.Println("Checking", *c)

	if c.ID != "" && c.Description != "" && c.ImageURL != "" &&
		c.DateTime != nil && *c.DateTime > 0 && c.TotalTickets != nil && *c.TotalTickets > 0 &&
		c.TicketsSold != nil && c.FullPrice > 0.0 && c.ConcessionPrice > 0.0 {
		valid = true
	}
	return
}

// GetConcertFromDatabase retrieves a specific concert from the dynamoDB table
func (d DDBHandler) GetConcertFromDatabase(concertID string) (concert *databaseHandler.Concert, err error) {
	concert = &databaseHandler.Concert{}
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(concertID),
			},
		},
	})
	if err != nil {
		fmt.Println("Issue getting item", err.Error())
		return
	} else if result.Item == nil {
		err = databaseHandler.ErrConcertDoesNotExist{Message: "Error does not exist"}
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, concert)
	if err != nil {
		fmt.Printf("Issue unmarshalling table data, %v\n", err)
		return
	}

	if !validateConcert(concert) {
		err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Invalid concert data for concert %s\n", concertID)}
		return
	}

	epochNow := time.Now().Unix()
	if *concert.DateTime < epochNow {
		err = databaseHandler.ErrConcertInPast{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", concertID)}
		fmt.Println(err.Error())
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
		TableName:                 aws.String(tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		fmt.Println("Issue getting items", err.Error())
		return
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &concerts)
	if err != nil {
		fmt.Printf("Issue unmarshalling table data, %v\n", err)
		return
	}

	for _, v := range concerts {
		if !validateConcert(&v) {
			err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", v.ID)}
			fmt.Println(err.Error())
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
