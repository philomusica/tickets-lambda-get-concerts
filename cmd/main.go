package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler/ddbHandler"
)

const DEFAULT_JSON_RESPONSE string = `{"error": "unable to retrieve concert data"}`

// ===============================================================================================================================
// PRIVATE FUNCTIONS
// ===============================================================================================================================
func getConcertData(request events.APIGatewayProxyRequest, dynamoHandler databaseHandler.DatabaseHandler) (response events.APIGatewayProxyResponse, err error) {

	response = events.APIGatewayProxyResponse{
		Body:       DEFAULT_JSON_RESPONSE,
		StatusCode: 404,
		Headers:    map[string]string{"Access-Control-Allow-Origin": "https://philomusica.org.uk"},
	}
	var byteArray []byte
	id := request.QueryStringParameters["id"]
	if id == "" {
		var concerts []databaseHandler.Concert
		concerts, err = dynamoHandler.GetConcertsFromTable()
		if err != nil || len(concerts) == 0 {
			fmt.Println("error from GetConcertsFromTable is ", err)
			return
		}
		for i := 0; i < len(concerts); i++ {
			err = dynamoHandler.ReformatDateTimeAndTickets(&concerts[i])
			if err != nil {
				fmt.Println("error from ReformatDateTimeAndTickets is ", err)
				return
			}
		}
		byteArray, err = json.Marshal(&concerts)
		if err != nil {
			fmt.Println("error from Marshal is ", err)
			return
		}
	} else {
		var concert *databaseHandler.Concert
		concert, err = dynamoHandler.GetConcertFromTable(id)
		if err != nil {
			return
		}
		err = dynamoHandler.ReformatDateTimeAndTickets(concert)
		if err != nil {
			return
		}
		byteArray, err = json.Marshal(&concert)
		if err != nil {
			return
		}
	}

	response.Body = string(byteArray)
	response.StatusCode = 200

	return
}

// ===============================================================================================================================
// END OF PRIVATE FUNCTIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

// Handler is lambda handler function that executes the relevant business logic
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	response := events.APIGatewayProxyResponse{
		Body:       DEFAULT_JSON_RESPONSE,
		StatusCode: 404,
		Headers:    map[string]string{"Access-Control-Allow-Origin": "https://philomusica.org.uk"},
	}

	sess, err := session.NewSession()
	if err != nil {
		return response, nil
	}
	svc := dynamodb.New(sess)

	concertsTable := os.Getenv("CONCERTS_TABLE")
	ordersTable := os.Getenv("ORDERS_TABLE")
	if concertsTable == "" || ordersTable == "" {
		fmt.Println("CONCERTS_TABLE and/or ORDERS_TABLE environment variables not set")
		response.StatusCode = 500
		return response, nil
	}

	dynamoHandler := ddbHandler.New(svc, concertsTable, ordersTable)

	response, err = getConcertData(request, dynamoHandler)
	if err != nil {
		fmt.Println(err)
	}
	return response, nil
}

func main() {
	lambda.Start(Handler)
}

// ===============================================================================================================================
// END OF PUBLIC FUNCTIONS
// ===============================================================================================================================
