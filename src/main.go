package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"time"
)

// Concert is a model of a concert which contains basic info regarding a concert, taken from dynamoDB
type Concert struct {
	Description string
	ImageURL    string
	Date        string
	Time        string
}

// ConvertEpochSecsToDateAndTimeStrings converts an epoch seconds time stamp to a date and time string in the format of Mon 2 Jan 2006 and 3:04PM
func ConvertEpochSecsToDateAndTimeStrings(dateTime int64) (date string, timeStamp string) {
	t := time.Unix(dateTime, 0)
	fmt.Println(t)
	date = t.Format("Mon 2 Jan 2006")
	timeStamp = t.Format("3:04 PM")
	return
}

// Handler is lambda handler function that executes the relevant business logic
func Handler() (events.APIGatewayProxyResponse, error) {

	response := events.APIGatewayProxyResponse{
		/* Body: string(b), */
		Body:       fmt.Sprintf("Unable to retrieve concerts"),
		StatusCode: 404,
	}
	concerts := make([]Concert, 3)

	br, err := json.Marshal(&concerts)
	if err != nil {
		return response, err
	}

	response.Body = string(br)

	return response, nil
}

func main() {
	lambda.Start(Handler)
}
