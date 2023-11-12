package main

import (
	//	"fmt"
	//	"os"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
)

/*
func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.7 {
			fmt.Printf("Tests passed but coverage was below %d%%\n", int(c*100))
			rc = -1
		}
	}
	os.Exit(rc)
}
*/

var (
	dt             int64                   = 1672599600
	tt             uint16                  = 300
	ts             uint16                  = 100
	exampleConcert databaseHandler.Concert = databaseHandler.Concert{
		ID:               "ABC",
		Title:            "Summer Concert",
		ImageURL:         "https://example.com/image1",
		Location:         "Holy Trinity, Longlevens, GL2 0AJ",
		DateTime:         &dt,
		Date:             "",
		Time:             "",
		TotalTickets:     &tt,
		TicketsSold:      &ts,
		AvailableTickets: 0,
		FullPrice:        11.00,
		ConcessionPrice:  9.00,
	}
)

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

func TestGetConcertDataCannotUnmarshalConcert(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")
	mockddbHandler := mockDDBHandlerGetConcertsReturnsEmptyConcertsSlice{}
	response, _ := getConcertData(request, mockddbHandler)
	expectedStatusCode := 404
	expectedBody := DEFAULT_JSON_RESPONSE

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

type mockDDBHandlerGetConcertsReformatFails struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertsReformatFails) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	c := exampleConcert
	concerts = append(concerts, c)
	return
}

func (m mockDDBHandlerGetConcertsReformatFails) GetConcertFromTable(concertId string) (concert *databaseHandler.Concert, err error) {
	concert = &exampleConcert
	return
}

func (m mockDDBHandlerGetConcertsReformatFails) ReformatDateTimeAndTickets(concert *databaseHandler.Concert) (err error) {
	err = databaseHandler.ErrConcertDoesNotExist{Message: "Nil value passed to reformater"}
	return
}

func TestGetConcertDataReformatingConcertsFails(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

	mockddbHandler := mockDDBHandlerGetConcertsReformatFails{}
	_, err := getConcertData(request, mockddbHandler)

	expectedErr, ok := err.(databaseHandler.ErrConcertDoesNotExist)

	if !ok {
		t.Errorf("Expected error of type %T, got %T\n", expectedErr, err)
	}
}

type mockDDBHandlerGetConcertsSuccess struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertsSuccess) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	c := exampleConcert
	concerts = append(concerts, c)
	return
}

func (m mockDDBHandlerGetConcertsSuccess) ReformatDateTimeAndTickets(concert *databaseHandler.Concert) (err error) {
	t := time.Unix(*concert.DateTime, 0)
	dateStr := t.Format("Mon 2 Jan 2006")
	timeStr := t.Format("3:04 PM")
	concert.Date = dateStr
	concert.Time = timeStr
	concert.DateTime = nil
	concert.AvailableTickets = *concert.TotalTickets - *concert.TicketsSold
	concert.TotalTickets = nil
	concert.TicketsSold = nil
	return
}

func TestGetConcertDataGetConcertsSuccess(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

	mockddbHandler := mockDDBHandlerGetConcertsSuccess{}
	response, _ := getConcertData(request, mockddbHandler)

	expectedStatusCode := 200
	expectedBody := `[{"id":"ABC","title":"Summer Concert","imageURL":"https://example.com/image1","location":"Holy Trinity, Longlevens, GL2 0AJ","date":"Sun 1 Jan 2023","time":"7:00 PM","availableTickets":200,"fullPrice":11,"concessionPrice":9}]`

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", expectedStatusCode, expectedBody, response.StatusCode, response.Body)
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

func TestGetConcertDataReformatingConcertFails(t *testing.T) {
	params := make(map[string]string)
	params["id"] = "ABC"
	request := events.APIGatewayProxyRequest{
		QueryStringParameters: params,
	}
	t.Setenv("CONCERTS_TABLE", "concerts-table")
	t.Setenv("ORDERS_TABLE", "orders-table")

	mockddbHandler := mockDDBHandlerGetConcertsReformatFails{}
	_, err := getConcertData(request, mockddbHandler)

	expectedErr, ok := err.(databaseHandler.ErrConcertDoesNotExist)

	if !ok {
		t.Errorf("Expected error of type %T, got %T\n", expectedErr, err)
	}
}

type mockDDBHandlerGetConcertSuccess struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerGetConcertSuccess) GetConcertFromTable(concertId string) (concert *databaseHandler.Concert, err error) {
	concert = &exampleConcert
	return
}

func (m mockDDBHandlerGetConcertSuccess) ReformatDateTimeAndTickets(concert *databaseHandler.Concert) (err error) {
	t := time.Unix(*concert.DateTime, 0)
	dateStr := t.Format("Mon 2 Jan 2006")
	timeStr := t.Format("3:04 PM")
	concert.Date = dateStr
	concert.Time = timeStr
	concert.DateTime = nil
	concert.AvailableTickets = *concert.TotalTickets - *concert.TicketsSold
	concert.TotalTickets = nil
	concert.TicketsSold = nil
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
	expectedBody := `{"id":"ABC","title":"Summer Concert","imageURL":"https://example.com/image1","location":"Holy Trinity, Longlevens, GL2 0AJ","date":"Sun 1 Jan 2023","time":"7:00 PM","availableTickets":200,"fullPrice":11,"concessionPrice":9}`

	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", expectedStatusCode, expectedBody, response.StatusCode, response.Body)
	}
}

// ===============================================================================================================================
// HANDLER TESTS
// ===============================================================================================================================

func TestHandlerEnvironmentVariablesNotSet(t *testing.T) {
	request := events.APIGatewayProxyRequest{}
	response, _ := Handler(request)
	expectedStatusCode := 500
	expectedBody := DEFAULT_JSON_RESPONSE
	if response.StatusCode != expectedStatusCode || response.Body != expectedBody {
		t.Errorf("Expected status code %d and body %s, got %d and %s\n", response.StatusCode, response.Body, expectedStatusCode, expectedBody)
	}
}

// ===============================================================================================================================
// END HANDLER TESTS
// ===============================================================================================================================
