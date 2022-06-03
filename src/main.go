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
	"os"
	"time"
)

var (
	tableName = os.Getenv("TABLE_NAME")
)

// ErrNoConcerts is a error type to handle if there are no upcoming concerts in the dynamoDB table
type ErrNoConcerts struct {
	message string
}

// NewErrNoConcerts is an function to generate a new ErrNoConcerts error type
func NewErrNoConcerts(message string) ErrNoConcerts {
	return ErrNoConcerts{
		message: message,
	}
}
func (e ErrNoConcerts) Error() string {
	return e.message
}

// Concert is a model of a concert which contains basic info regarding a concert, taken from dynamoDB
type Concert struct {
	Description     string
	ImageURL        string
	Date            string
	Time            string
	ConcertDateTime int64
}

// ConvertEpochSecsToDateAndTimeStrings converts an epoch seconds time stamp to a date and time string in the format of Mon 2 Jan 2006 and 3:04PM
func ConvertEpochSecsToDateAndTimeStrings(dateTime int64) (date string, timeStamp string) {
	t := time.Unix(dateTime, 0)
	fmt.Println(t)
	date = t.Format("Mon 2 Jan 2006")
	timeStamp = t.Format("3:04 PM")
	return
}

// GetConcertsFromDynamoDB gets all upcoming concerts from the dynamoDB table
func GetConcertsFromDynamoDB(svc dynamodbiface.DynamoDBAPI, concerts *[]Concert) (err error) {
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		return
	}
	if *result.Count == int64(0) {
		err = NewErrNoConcerts("No upcoming concerts")
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
	fmt.Println(*concerts)

	return
}

// Handler is lambda handler function that executes the relevant business logic
func Handler() (response events.APIGatewayProxyResponse, err error) {

	response = events.APIGatewayProxyResponse{
		/* Body: string(b), */
		Body:       fmt.Sprintf("Unable to retrieve concerts"),
		StatusCode: 404,
	}
	concerts := make([]Concert, 3)
	sess := session.New()
	svc := dynamodb.New(sess)
	err = GetConcertsFromDynamoDB(svc, &concerts)
	if err != nil {
		switch err.(type) {
		case ErrNoConcerts:
			response.Body = err.Error()
			response.StatusCode = 200
		}

		return
	}

	for _, v := range concerts {
		dateStr, timeStr := ConvertEpochSecsToDateAndTimeStrings(v.ConcertDateTime)
		v.Date = dateStr
		v.Time = timeStr
	}

	br, err := json.Marshal(&concerts)
	if err != nil {
		return
	}

	response.Body = string(br)

	return
}

func main() {
	lambda.Start(Handler)
}
