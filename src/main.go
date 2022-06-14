package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"os"
	"time"
)

var (
	tableName = os.Getenv("TABLE_NAME")
)

// Concert is a model of a concert which contains basic info regarding a concert, taken from dynamoDB
type Concert struct {
	ID              string
	Description     string
	ImageURL        string
	DateTime        int64
	TotalTickets    int
	TicketsSold     int
	FullPrice       float32
	ConcessionPrice float32
}

// ClientConcert is a model of a concert which contains basic info regarding a concert, taken from dynamoDB
type ClientConcert struct {
	ID               string  `json:"id"`
	Description      string  `json:"description"`
	ImageURL         string  `json:"imageURL"`
	Date             string  `json:"date"`
	Time             string  `json:"time"`
	AvailableTickets int     `json:"availableTickets"`
	FullPrice        float32 `json:"fullPrice"`
	ConcessionPrice  float32 `json:"conessionPrice"`
}

// ErrConcertInPast is a custom error message to signify concert is in past and tickets can no longer be purchased for it
type ErrConcertInPast struct {
	Message string
}

func (e ErrConcertInPast) Error() string {
	return e.Message
}

// ErrConcertDoesNotExist is a custom error message to signify the concert with a given ID does not exist
type ErrConcertDoesNotExist struct {
	Message string
}

func (e ErrConcertDoesNotExist) Error() string {
	return e.Message
}

// ConvertEpochSecsToDateAndTimeStrings converts an epoch seconds time stamp to a date and time string in the format of Mon 2 Jan 2006 and 3:04PM
func ConvertEpochSecsToDateAndTimeStrings(dateTime int64) (date string, timeStamp string) {
	t := time.Unix(dateTime, 0)
	fmt.Println(t)
	date = t.Format("Mon 2 Jan 2006")
	timeStamp = t.Format("3:04 PM")
	return
}

// GetConcertFromDynamoDB retrieves a specific concert the dynamoDB table
func GetConcertFromDynamoDB(svc dynamodbiface.DynamoDBAPI, concertID string, concert *Concert) (err error) {
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(concertID),
			},
		},
	})

	if err != nil {
		return
	}

	if result.Item == nil {
		err = ErrConcertDoesNotExist{Message: "Error does not exist"}
		return
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, concert)
	if err != nil {
		fmt.Printf("Issue unmarshalling table data, %v\n", err)
		return
	}

	epochNow := time.Now().Unix()
	if concert.DateTime < epochNow {
		err = ErrConcertInPast{Message: "Error concert in the past, tickets are no longer available"}
		fmt.Printf("Concert %s is in the past. Tickets are no longer available\n", concertID)
		return
	}

	return
}

// GetConcertsFromDynamoDB gets all upcoming concerts from the dynamoDB table
func GetConcertsFromDynamoDB(svc dynamodbiface.DynamoDBAPI, concerts *[]Concert) (err error) {
	epochNow := time.Now().Unix()
	filt := expression.Name("DateTime").GreaterThan(expression.Value(epochNow))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return
	}

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		return
	}

	for _, item := range result.Items {
		concert := Concert{}
		err = dynamodbattribute.UnmarshalMap(item, &concert)
		if err != nil {
			return
		}
		*concerts = append(*concerts, concert)
	}

	return
}

// GetConcert returns a json array of all concerts, marshalled into a byte array
func GetConcert(id string) (jsonByteArray []byte, err error) {
	concert := &Concert{}
	sess := session.New()
	svc := dynamodb.New(sess)
	err = GetConcertFromDynamoDB(svc, id, concert)
	if err != nil {
		return
	}

	dateStr, timeStr := ConvertEpochSecsToDateAndTimeStrings(concert.DateTime)
	c := ClientConcert{
		ID:               concert.ID,
		Description:      concert.Description,
		ImageURL:         concert.ImageURL,
		Date:             dateStr,
		Time:             timeStr,
		AvailableTickets: concert.TotalTickets - concert.TicketsSold,
		FullPrice:        concert.FullPrice,
		ConcessionPrice:  concert.ConcessionPrice,
	}
	jsonByteArray, err = json.Marshal(&c)
	if err != nil {
		return
	}
	return
}

// GetAllConcerts returns a json array of all concerts, marshalled into a byte array
func GetAllConcerts() (jsonByteArray []byte, err error) {
	concerts := make([]Concert, 0, 3)
	sess := session.New()
	svc := dynamodb.New(sess)
	err = GetConcertsFromDynamoDB(svc, &concerts)
	if err != nil {
		return
	}

	clientConcerts := make([]ClientConcert, 0, 3)

	for _, v := range concerts {
		dateStr, timeStr := ConvertEpochSecsToDateAndTimeStrings(v.DateTime)
		c := ClientConcert{
			ID:               v.ID,
			Description:      v.Description,
			ImageURL:         v.ImageURL,
			Date:             dateStr,
			Time:             timeStr,
			AvailableTickets: v.TotalTickets - v.TicketsSold,
			FullPrice:        v.FullPrice,
			ConcessionPrice:  v.ConcessionPrice,
		}
		clientConcerts = append(clientConcerts, c)
	}
	jsonByteArray, err = json.Marshal(&clientConcerts)
	if err != nil {
		return
	}

	return
}

// Handler is lambda handler function that executes the relevant business logic
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	response := events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Unable to retrieve concerts"),
		StatusCode: 404,
	}

	var br []byte
	var err error
	id := request.QueryStringParameters["id"]
	if id == "" {
		br, err = GetAllConcerts()
		if err != nil {
			return response, nil
		}
	} else {
		br, err = GetConcert(id)
		if err != nil {
			return response, nil
		}
	}

	response.Body = string(br)
	response.StatusCode = 200

	return response, nil
}

func main() {
	lambda.Start(Handler)
}
