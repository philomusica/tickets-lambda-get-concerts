package cmd

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"os"
	"testing"
)

func TestHandlerEnvironmentVariablesNotSet(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	response, _ := Handler(request)
	expectedStatusCode := 500
	expectedBody := "Unable to retrieve concerts - Internal Server Error"
	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

type mockDDBHandlerGetConcertsFails struct {
	databaseHandler.DatabaseHandler
} 

func (m mockDDBHandlerGetConcertsFails) GetConcertsFromDatabase() (concerts []databaseHandler.Concert, err error) {
	err = databaseHandler.ErrInvalidConcertData{Message: "Invalid concert data"}
	return
}

func TestHandlerCallToGetConcertsFails(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	os.Setenv("CONCERTS_TABLE", "concerts-table")
	os.Setenv("PURCHASED_TICKETS_LE", "purchased-tickets-table")
	mockddbHandler := mockDDBHandlerGetConcertsFails{}
	_, err := getConcertData(request, mockddbHandler)
	
	expectedErr, ok := err.(databaseHandler.ErrInvalidConcertData)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

type mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice struct {
	databaseHandler.DatabaseHandler
} 

func (m mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice) GetConcertsFromDatabase() (concerts []databaseHandler.Concert, err error) {
	return
}

func TestHandlerUnableToMarshalConcert(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	os.Setenv("CONCERTS_TABLE", "concerts-table")
	os.Setenv("PURCHASED_TICKETS_LE", "purchased-tickets-table")
	mockddbHandler := mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice{}
	response, _ := getConcertData(request, mockddbHandler)
	expectedStatusCode := 404
	expectedBody := "Unable to retrieve concerts"
	
	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

type mockDDBHandlerGetConcertsSuccess struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertsSuccess) GetConcertsFromDatabase() (concerts []databaseHandler.Concert, err error) {
	c := databaseHandler.Concert{
		ID: "ABC",
		Description: "Summer Concert",
		ImageURL: "https://example.com/image1",
		Date: "Monday 1 January 2023",
		Time: "7:00pm",
		AvailableTickets: 30,
		FullPrice: 11.00,
		ConcessionPrice: 9.00,
	}
	concerts = append(concerts, c)
	return
}

func TestHandlerGetConcertsSuccess(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	os.Setenv("CONCERTS_TABLE", "concerts-table")
	os.Setenv("PURCHASED_TICKETS_LE", "purchased-tickets-table")

	mockddbHandler := mockDDBHandlerGetConcertsSuccess{}
	response, _ := getConcertData(request, mockddbHandler)

	expectedStatusCode := 200
	expectedBody := `[{"id":"ABC","description":"Summer Concert","imageURL":"https://example.com/image1","date":"Monday 1 January 2023","time":"7:00pm","availableTickets":30,"fullPrice":11,"concessionPrice":9}]`

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

type mockDDBHandlerGetConcertFails struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertFails) GetConcertsFromDatabase(concertId string) (concert *databaseHandler.Concert, err error) {
	err = databaseHandler.ErrConcertInPast{Message: "Concert x is in the past"}
	return
}

func TestHandlerGetConcertReturnsError(t *testing.T) {

}
