package cmd

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/dynamodb"
)


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
		br, err = dynamodb.GetAllConcerts()
		if err != nil {
			return response, nil
		}
	} else {
		br, err = dynamodb.GetConcert(id)
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
