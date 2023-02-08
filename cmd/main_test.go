package cmd

import (
	//	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	//"os"
	"testing"
)

/*
	func TestMain(m *testing.M) {
		rc := m.Run()

		if rc == 0 && testing.CoverMode() != "" {
			c := testing.Coverage()
			fmt.Println(c)
			if c < 0.7 {
				fmt.Printf("Tests passed but coverage was below %d%%\n", int(c*100))
				rc = -1
			}
		}
		os.Exit(rc)
	}
*/

// ===============================================================================================================================
// GET_CONCERT_DATA TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// END GET_CONCERT_DATA TESTS
// ===============================================================================================================================

type mockDDBHandlerGetConcertsFails struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertsFails) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	err = databaseHandler.ErrInvalidConcertData{Message: "Invalid concert data"}
	return
}

func TestGetConcertDataGetConcertsFails(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")
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

func (m mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	return
}

func TestGetConcertDataCannotMarshalConcert(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")
	mockddbHandler := mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice{}
	response, _ := getConcertData(request, mockddbHandler)
	expectedStatusCode := 404
	expectedBody := "Unable to retrieve concerts"

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

// ===============================================================================================================================
// HANDLER TESTS
// ===============================================================================================================================

func TestHandlerEnvironmentVariablesNotSet(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	response, _ := Handler(request)
	expectedStatusCode := 500
	expectedBody := "Unable to retrieve concerts - Internal Server Error"
	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

type mockDDBHandlerGetConcertsSuccess struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertsSuccess) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	c := databaseHandler.Concert{
		ID:               "ABC",
		Description:      "Summer Concert",
		ImageURL:         "https://example.com/image1",
		Date:             "Monday 1 January 2023",
		Time:             "7:00pm",
		AvailableTickets: 30,
		FullPrice:        11.00,
		ConcessionPrice:  9.00,
	}
	concerts = append(concerts, c)
	return
}

func TestGetConcertDataGetConcertsSuccess(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

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

func (m mockDDBHandlerGetConcertFails) GetConcertFromTable(concertId string) (concert *databaseHandler.Concert, err error) {
	err = databaseHandler.ErrConcertInPast{Message: "Concert x is in the past"}
	return
}

func TestGetConcertDataGetConcertFails(t *testing.T) {
	params := make(map[string]string)
	params["id"] = "ABC"
	request := events.APIGatewayProxyRequest{
		QueryStringParameters: params,
	}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

	mockddbHandler := mockDDBHandlerGetConcertFails{}
	_, err := getConcertData(request, mockddbHandler)
	expectedErr, ok := err.(databaseHandler.ErrConcertInPast)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

type mockDDBHandlerGetConcertSuccess struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertSuccess) GetConcertFromTable(concertId string) (concert *databaseHandler.Concert, err error) {
	concert = &databaseHandler.Concert{
		ID:               "ABC",
		Description:      "Summer Concert",
		ImageURL:         "https://example.com/image1",
		Date:             "Monday 1 January 2023",
		Time:             "7:00pm",
		AvailableTickets: 30,
		FullPrice:        11.00,
		ConcessionPrice:  9.00,
	}
	return
}

func TestGetConcertDataGetConcertSuccess(t *testing.T) {
	params := make(map[string]string)
	params["id"] = "ABC"
	request := events.APIGatewayProxyRequest{
		QueryStringParameters: params,
	}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

	mockddbHandler := mockDDBHandlerGetConcertSuccess{}
	response, err := getConcertData(request, mockddbHandler)

	if err != nil {
		t.Errorf("Expected no error, got %T\n", err)
	}

	expectedStatusCode := 200
	expectedBody := `{"id":"ABC","description":"Summer Concert","imageURL":"https://example.com/image1","date":"Monday 1 January 2023","time":"7:00pm","availableTickets":30,"fullPrice":11,"concessionPrice":9}`

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

// ===============================================================================================================================
// END HANDLER TESTS
// ===============================================================================================================================
