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
	"os"
)

func getConcertData(request events.APIGatewayProxyRequest, dynamoHandler databaseHandler.DatabaseHandler) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Unable to retrieve concerts"),
		StatusCode: 404,
	}
	var byteArray []byte
	var err error
	id := request.QueryStringParameters["id"]
	if id == "" {
		var concerts []databaseHandler.Concert
		concerts, err = dynamoHandler.GetConcertsFromDatabase()
		if err != nil || len(concerts) == 0 {
			return response, err
		}
		byteArray, err = json.Marshal(&concerts)
		if err != nil {
			return response, err
		}
	} else {
		var concert *databaseHandler.Concert
		concert, err = dynamoHandler.GetConcertFromDatabase(id)
		if err != nil {
			return response, err
		}
		byteArray, err = json.Marshal(&concert)
		if err != nil {
			return response, err
		}
	}

	response.Body = string(byteArray)
	response.StatusCode = 200

	return response, nil
}

// Handler is lambda handler function that executes the relevant business logic
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Unable to retrieve concerts - Internal Server Error"),
		StatusCode: 404,
	}

	sess := session.New()
	svc := dynamodb.New(sess)

	concertsTable := os.Getenv("CONCERTS_TABLE")
	ordersTable := os.Getenv("ORDERS_TABLE")
	if concertsTable == "" || ordersTable == "" {
		fmt.Println("CONCERT_TABLE and/or ORDERS_TABLE environment variables not set")
		response.StatusCode = 500
		return response, nil
	}

	dynamoHandler := ddbHandler.New(svc, concertsTable, ordersTable)

	return getConcertData(request, dynamoHandler)
}

func main() {
	lambda.Start(Handler)
}
