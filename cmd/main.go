package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler/ddbHandler"
)

// Handler is lambda handler function that executes the relevant business logic
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	response := events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Unable to retrieve concerts"),
		StatusCode: 404,
	}

	sess := session.New()
	svc := dynamodb.New(sess)
	dynamoHandler := ddbHandler.New(svc)

	var byteArray []byte
	var err error
	id := request.QueryStringParameters["id"]
	if id == "" {
		var concerts []databaseHandler.Concert
		concerts, err = dynamoHandler.GetConcertsFromDatabase()
		if err != nil {
			return response, nil
		}
		byteArray, err = json.Marshal(&concerts)
		if err != nil {
			return response, nil
		}
	} else {
		var concert *databaseHandler.Concert
		concert, err = dynamoHandler.GetConcertFromDatabase(id)
		if err != nil {
			return response, nil
		}
		byteArray, err = json.Marshal(concert)
		if err != nil {
			return response, nil
		}
	}

	response.Body = string(byteArray)
	response.StatusCode = 200

	return response, nil
}

func main() {
	lambda.Start(Handler)
}
