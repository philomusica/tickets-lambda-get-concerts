package ddbHandler

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

var summerEpoch int64 = 1656176400 // 25/06/22 18:00
var winterEpoch int64 = 1671991200 // 25/12/22 18:00

func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.9 {
			fmt.Printf("Tests passed but coverage was below %d%%\n", int(c*100))
			rc = -1
		}
	}
	os.Exit(rc)
}

// ===============================================================================================================================
// CONVERT_EPOCH_SECS_TO_DATE_AND_TIME_STRINGS TESTS
// ===============================================================================================================================
func TestConvertEpochSecsToDateAndTimeStringsDateValueSummer(t *testing.T) {
	timeStamp := time.Unix(summerEpoch, 0)
	expectedDate := timeStamp.Format("Mon 2 Jan 2006")
	result, _ := convertEpochSecsToDateAndTimeStrings(summerEpoch)

	if result != expectedDate {
		t.Errorf("Expected %s, got %s\n", expectedDate, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsTimeValueSummer(t *testing.T) {
	timeStamp := time.Unix(summerEpoch, 0)
	expectedTime := timeStamp.Format("3:04 PM")
	_, result := convertEpochSecsToDateAndTimeStrings(summerEpoch)
	if result != expectedTime {
		t.Errorf("Expected %s, got %s\n", expectedTime, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsDateValueWinter(t *testing.T) {
	timeStamp := time.Unix(winterEpoch, 0)
	expectedDate := timeStamp.Format("Mon 2 Jan 2006")
	result, _ := convertEpochSecsToDateAndTimeStrings(winterEpoch)

	if result != expectedDate {
		t.Errorf("Expected %s, got %s\n", expectedDate, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsTimeValueWinter(t *testing.T) {
	timeStamp := time.Unix(winterEpoch, 0)
	expectedTime := timeStamp.Format("3:04 PM")
	_, result := convertEpochSecsToDateAndTimeStrings(winterEpoch)
	if result != expectedTime {
		t.Errorf("Expected %s, got %s\n", expectedTime, result)
	}
}

// ===============================================================================================================================
// END CONVERT_EPOCH_SECS_TO_DATE_AND_TIME_STRINGS TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// CREATE_ORDER_ENTRY TESTS
// ===============================================================================================================================
type mockDynamoDBClientCannotPut struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientCannotPut) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, &dynamodb.ResourceNotFoundException{}
}

func TestCreateOrderEntryCannotPut(t *testing.T) {
	mockSvc := &mockDynamoDBClientCannotPut{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.createOrderEntry(&order)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

// ===============================================================================================================================
// END CREATE_ORDER_ENTRY TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// GENERATE_ORDER_REFERENCE TESTS
// ===============================================================================================================================

func TestGenerateOrderReference(t *testing.T) {
	var size uint8 = 4
	result := generateOrderReference(size)
	if uint8(len(result)) != size {
		t.Errorf("Expected reference of size %v, got %v", size, len(result))
	}
}

// ===============================================================================================================================
// END GENERATE_ORDER_REFERENCE TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// VALIDATE_CONCERTS TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// END VALIDATE_CONCERTS TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// CREATE_ORDER_IN_TABLE TESTS
// ===============================================================================================================================

type mockDynamoDBClientOrderReferenceMatchOnce struct {
	dynamodbiface.DynamoDBAPI
	firstCall bool
}

func (m *mockDynamoDBClientOrderReferenceMatchOnce) PutItem(input *dynamodb.PutItemInput) (output *dynamodb.PutItemOutput, err error) {
	if m.firstCall {
		err = &dynamodb.ConditionalCheckFailedException{}
	}
	m.firstCall = false
	return
}

func TestCreateEntryInOrdersTableReferenceMatchOnce(t *testing.T) {
	mockSvc := &mockDynamoDBClientOrderReferenceMatchOnce{firstCall: true}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.CreateOrderInTable(order)
	if err != nil {
		t.Errorf("Expected nil err, got %T", err)
	}
}

type mockDynamoDBClientOrderCannotPut struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientOrderCannotPut) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, &dynamodb.ResourceNotFoundException{}
}

func TestCreateEntryInOrdersTableFails(t *testing.T) {
	mockSvc := &mockDynamoDBClientOrderCannotPut{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.CreateOrderInTable(order)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

// ===============================================================================================================================
// CREATE_ORDER_IN_TABLE TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// GET_CONCERT_FROM_TABLE TESTS
// ===============================================================================================================================

type mockDynamoDBClientConcertSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientConcertSuccess) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["ID"] = &dynamodb.AttributeValue{}
	item1["ID"].SetS("AAA")
	item1["Title"] = &dynamodb.AttributeValue{}
	item1["Title"].SetS("Summer Concert")
	item1["ImageURL"] = &dynamodb.AttributeValue{}
	item1["ImageURL"].SetS("http://example.com/image.jpg")
	item1["DateTime"] = &dynamodb.AttributeValue{}
	item1["DateTime"].SetN(fmt.Sprint(summerEpoch))
	item1["TotalTickets"] = &dynamodb.AttributeValue{}
	item1["TotalTickets"].SetN(fmt.Sprint(250))
	item1["TicketsSold"] = &dynamodb.AttributeValue{}
	item1["TicketsSold"].SetN(fmt.Sprint(50))
	item1["FullPrice"] = &dynamodb.AttributeValue{}
	item1["FullPrice"].SetN(fmt.Sprint(12.00))
	item1["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item1["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	items = append(items, item1)
	item2 := map[string]*dynamodb.AttributeValue{}
	item2["ID"] = &dynamodb.AttributeValue{}
	item2["ID"].SetS("BBB")
	item2["Title"] = &dynamodb.AttributeValue{}
	item2["Title"].SetS("Winter Concert")
	item2["ImageURL"] = &dynamodb.AttributeValue{}
	item2["ImageURL"].SetS("http://example.com/image2.jpg")
	item2["DateTime"] = &dynamodb.AttributeValue{}
	item2["DateTime"].SetN(fmt.Sprint(summerEpoch))
	item2["TotalTickets"] = &dynamodb.AttributeValue{}
	item2["TotalTickets"].SetN(fmt.Sprint(250))
	item2["TicketsSold"] = &dynamodb.AttributeValue{}
	item2["TicketsSold"].SetN(fmt.Sprint(50))
	item2["FullPrice"] = &dynamodb.AttributeValue{}
	item2["FullPrice"].SetN(fmt.Sprint(12.00))
	item2["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item2["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	items = append(items, item2)
	numConcertsI64 := int64(numConcerts)
	output = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func (m *mockDynamoDBClientConcertSuccess) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	epochTomorrow := time.Now().AddDate(0, 0, 1).Unix()
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetN(fmt.Sprint(epochTomorrow))
	item["TotalTickets"] = &dynamodb.AttributeValue{}
	item["TotalTickets"].SetN(fmt.Sprint(250))
	item["TicketsSold"] = &dynamodb.AttributeValue{}
	item["TicketsSold"].SetN(fmt.Sprint(50))
	item["FullPrice"] = &dynamodb.AttributeValue{}
	item["FullPrice"].SetN(fmt.Sprint(12.00))
	item["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	output.SetItem(item)
	return &output, nil
}

func TestGetConcertFromTableSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertSuccess{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concert, err := dynamoHandler.GetConcertFromTable(concertID)
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err.Error())
	}

	if concert.ID != concertID {
		t.Errorf("Expected entry with ID %s, got %s\n", concertID, concert.ID)
	}
}

type mockDynamoDBClientConcertResourceNotFound struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientConcertResourceNotFound) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	err = &dynamodb.ResourceNotFoundException{}

	return
}

func (m *mockDynamoDBClientConcertResourceNotFound) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	err := &dynamodb.ResourceNotFoundException{}
	return nil, err
}

func TestGetConcertFromTableCannotAccessTable(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertResourceNotFound{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromTable(concertID)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected %s error type, got %s\n", expectedErr, err)
	}
}

type mockDynamoDBClientNoConcert struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientNoConcert) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 0
	items := make([]map[string]*dynamodb.AttributeValue, 0)
	numConcertsI64 := int64(numConcerts)
	output = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func (m *mockDynamoDBClientNoConcert) GetItem(*dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	cc := dynamodb.ConsumedCapacity{}
	output.SetConsumedCapacity(&cc)
	output.SetItem(nil)
	return
}

func TestGetConcertFromTableNoConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientNoConcert{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromTable(concertID)

	errMessage, ok := err.(databaseHandler.ErrConcertDoesNotExist)
	if !ok {
		t.Errorf("Expected ErrConcertDoesNotExist error, got %s\n", errMessage)
	}
}

type mockDynamoDBClientConcertInvalidData struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientConcertInvalidData) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	epochYesterday := time.Now().AddDate(0, 0, -1).Unix()
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetN(fmt.Sprint(epochYesterday))
	item["TotalTickets"] = &dynamodb.AttributeValue{}
	item["TotalTickets"].SetN(fmt.Sprint(250))
	item["TicketsSold"] = &dynamodb.AttributeValue{}
	item["TicketsSold"].SetN(fmt.Sprint(50))
	output.SetItem(item)
	return &output, nil
}

func (m *mockDynamoDBClientConcertInvalidData) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["ID"] = &dynamodb.AttributeValue{}
	item1["ID"].SetS("AAA")
	item1["Title"] = &dynamodb.AttributeValue{}
	item1["Title"].SetS("Summer Concert")
	item1["ImageURL"] = &dynamodb.AttributeValue{}
	item1["ImageURL"].SetS("http://example.com/image.jpg")
	item1["DateTime"] = &dynamodb.AttributeValue{}
	item1["DateTime"].SetN(fmt.Sprint(summerEpoch))
	item1["TotalTickets"] = &dynamodb.AttributeValue{}
	item1["TotalTickets"].SetN(fmt.Sprint(250))
	item1["TicketsSold"] = &dynamodb.AttributeValue{}
	item1["TicketsSold"].SetN(fmt.Sprint(50))
	item1["FullPrice"] = &dynamodb.AttributeValue{}
	item1["FullPrice"].SetN(fmt.Sprint(12.00))
	item1["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item1["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	items = append(items, item1)
	item2 := map[string]*dynamodb.AttributeValue{}
	item2["ID"] = &dynamodb.AttributeValue{}
	item2["ID"].SetS("BBB")
	item2["Title"] = &dynamodb.AttributeValue{}
	item2["Title"].SetS("Winter Concert")
	item2["ImageURL"] = &dynamodb.AttributeValue{}
	item2["ImageURL"].SetS("http://example.com/image2.jpg")
	item2["TotalTickets"] = &dynamodb.AttributeValue{}
	item2["TotalTickets"].SetN(fmt.Sprint(250))
	item2["TicketsSold"] = &dynamodb.AttributeValue{}
	item2["TicketsSold"].SetN(fmt.Sprint(50))
	item2["FullPrice"] = &dynamodb.AttributeValue{}
	item2["FullPrice"].SetN(fmt.Sprint(12.00))
	item2["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item2["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	items = append(items, item2)
	numConcertsI64 := int64(numConcerts)
	output = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func TestGetConcertFromTableMissingTicketPrices(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertInvalidData{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromTable(concertID)
	expectedErr, ok := err.(databaseHandler.ErrInvalidConcertData)
	if !ok {
		t.Errorf("Expected %v error, got %v\n", expectedErr.Error(), err.Error())
	}
}

type mockDynamoDBClientOldConcert struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientOldConcert) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	epochYesterday := time.Now().AddDate(0, 0, -1).Unix()
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetN(fmt.Sprint(epochYesterday))
	item["TotalTickets"] = &dynamodb.AttributeValue{}
	item["TotalTickets"].SetN(fmt.Sprint(250))
	item["TicketsSold"] = &dynamodb.AttributeValue{}
	item["TicketsSold"].SetN(fmt.Sprint(50))
	item["FullPrice"] = &dynamodb.AttributeValue{}
	item["FullPrice"].SetN(fmt.Sprint(12.00))
	item["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	output.SetItem(item)
	return &output, nil
}

func TestGetConcertFromTableOldConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientOldConcert{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromTable(concertID)

	expectedErr, ok := err.(databaseHandler.ErrConcertInPast)

	if !ok {
		t.Errorf("Expected %s error, got %s\n", expectedErr, err)
	}
}

type mockDynamoDBClientConcertCannotUnmarshal struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientConcertCannotUnmarshal) GetItem(*dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetS("Hello")
	output.SetItem(item)
	return
}

func (m *mockDynamoDBClientConcertCannotUnmarshal) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["Title"] = &dynamodb.AttributeValue{}
	item1["Title"].SetS("Summer Concert")
	item1["ImageURL"] = &dynamodb.AttributeValue{}
	item1["ImageURL"].SetS("http://example.com/image.jpg")
	item1["DateTime"] = &dynamodb.AttributeValue{}
	item1["DateTime"].SetS("Hello")
	items = append(items, item1)
	item2 := map[string]*dynamodb.AttributeValue{}
	item2["Title"] = &dynamodb.AttributeValue{}
	item2["Title"].SetS("Winter Concert")
	item2["ImageURL"] = &dynamodb.AttributeValue{}
	item2["ImageURL"].SetS("http://example.com/image2.jpg")
	item2["DateTime"] = &dynamodb.AttributeValue{}
	item2["DateTime"].SetN(fmt.Sprintf("%d", winterEpoch))
	items = append(items, item2)
	numConcertsI64 := int64(numConcerts)
	output = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func TestGetConcertFromConcertCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertCannotUnmarshal{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromTable(concertID)

	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)

	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

// ===============================================================================================================================
// END GET_CONCERT_FROM_TABLE TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// GET_CONCERTS_FROM_TABLE TESTS
// ===============================================================================================================================

func TestGetConcertsFromTableSuccessful(t *testing.T) {
	expectedNumConcerts := 2
	mockSvc := &mockDynamoDBClientConcertSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concerts, err := dynamoHandler.GetConcertsFromTable()
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err)
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d\n", expectedNumConcerts, len(concerts))
	}

	firstConcertTitle := "Summer Concert"
	secondConcertTitle := "Winter Concert"

	if concerts[0].Title != firstConcertTitle {
		t.Errorf("Expected first concert returned to be %s, got %s\n", firstConcertTitle, concerts[0].Title)
	}

	if concerts[1].Title != secondConcertTitle {
		t.Errorf("Expected second concert returned to be %s, got %s\n", secondConcertTitle, concerts[1].Title)
	}
}

func TestGetConcertsFromTableNoConcerts(t *testing.T) {
	expectedNumConcerts := 0
	mockSvc := &mockDynamoDBClientNoConcert{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concerts, err := dynamoHandler.GetConcertsFromTable()
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err.Error())
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d\n", expectedNumConcerts, len(concerts))
	}
}

func TestGetConcertsFromTableResourceNotFound(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertResourceNotFound{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromTable()
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected %s error type, got %s\n", expectedErr, err)
	}
}

func TestGetConcertsFromTableMissingDateTime(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertInvalidData{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromTable()
	expectedErr, ok := err.(databaseHandler.ErrInvalidConcertData)
	if !ok {
		t.Errorf("Expected %v error, got %v\n", expectedErr.Error(), err.Error())
	}
}

func TestGetConcertsFromTableCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertCannotUnmarshal{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromTable()

	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)

	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

// ===============================================================================================================================
// END GET_CONCERTS_FROM_TABLE TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// GET_CONCERTS_FROM_TABLE TESTS
// ===============================================================================================================================

func TestGetOrderFromTableResourceNotFound(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertResourceNotFound{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderFromTable("1234", "A1B2")
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)
	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

type mockDynamoDBClientNoOrder struct {
	dynamodbiface.DynamoDBAPI
}

func (m mockDynamoDBClientNoOrder) GetItem(*dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	cc := dynamodb.ConsumedCapacity{}
	output.SetConsumedCapacity(&cc)
	output.SetItem(nil)
	return
}

func TestGetOrderFromTableNoOrder(t *testing.T) {
	mockSvc := mockDynamoDBClientNoOrder{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderFromTable("1234", "A1B2")
	expectedErr, ok := err.(paymentHandler.ErrOrderDoesNotExist)
	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

type mockDynamoDBClientOrderSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m mockDynamoDBClientOrderSuccess) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ConcertId"] = &dynamodb.AttributeValue{}
	item["ConcertId"].SetS("1234")
	item["Reference"] = &dynamodb.AttributeValue{}
	item["Reference"].SetS("A1B2")
	item["FirstName"] = &dynamodb.AttributeValue{}
	item["FirstName"].SetS("John")
	item["LastName"] = &dynamodb.AttributeValue{}
	item["LastName"].SetS("Smith")
	item["Email"] = &dynamodb.AttributeValue{}
	item["Email"].SetS("johnsmith@gmail.com")
	item["NumOfFullPrice"] = &dynamodb.AttributeValue{}
	item["NumOfFullPrice"].SetN(fmt.Sprint(2))
	item["NumOfConcessions"] = &dynamodb.AttributeValue{}
	item["NumOfConcessions"].SetN(fmt.Sprint(2))
	output.SetItem(item)
	return &output, nil
}

func TestGetOrderFromTableSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientOrderSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order, err := dynamoHandler.GetOrderFromTable("1234", "A1B2")
	if err != nil {
		t.Errorf("Expected no error, got %T\n", err)
	}
	expectedConcertId := "1234"
	expectedReference := "A1B2"
	if order.ConcertId != expectedConcertId || order.Reference != expectedReference {
		t.Errorf("Expected concertId of %s and order reference of %s, got %s and %s\n", expectedConcertId, expectedReference, order.ConcertId, order.Reference)
	}
}

type mockDynamoDBClientOrderCannotUnmarshal struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientOrderCannotUnmarshal) GetItem(input *dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ConcertId"] = &dynamodb.AttributeValue{}
	item["ConcertId"].SetBOOL(true)
	output.SetItem(item)
	return
}
func TestGetOrderFromTableCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientOrderCannotUnmarshal{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderFromTable("1234", "A1B2")
	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)
	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

// ===============================================================================================================================
// END GET_CONCERTS_FROM_TABLE TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// NEW TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// END NEW TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// REFORMAT_DATE_TIME_AND_TICKETS TESTS
// ===============================================================================================================================

func TestReformatDateTimeAndTicketsNilConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.ReformatDateTimeAndTickets(nil)
	expectedErr, ok := err.(databaseHandler.ErrConcertDoesNotExist)
	if !ok {
		t.Errorf("Expected error of type %T, go %T\n", expectedErr, err)
	}
}

func TestReformatDateTimeAndTicketsSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	var dT int64 = summerEpoch
	var tt uint16 = 300
	var ts uint16 = 100
	concert := databaseHandler.Concert{
		ID:              "ABC",
		Title:           "Summer Concert",
		ImageURL:        "http://example.com/image.jpg",
		DateTime:        &dT,
		TotalTickets:    &tt,
		TicketsSold:     &ts,
		FullPrice:       12.0,
		ConcessionPrice: 10.0,
	}
	err := dynamoHandler.ReformatDateTimeAndTickets(&concert)

	if err != nil {
		t.Errorf("Expected nil error, got %T", err)
	}

	if concert.AvailableTickets != 200 {
		t.Errorf("Expected available tickets to be calculated to 200, got %d\n", concert.AvailableTickets)
	}

	if concert.DateTime != nil || concert.TotalTickets != nil || concert.TicketsSold != nil {
		t.Errorf("Expected DateTime, TotalTickets and TicketsSold to all be nil\n")
	}
}

// ===============================================================================================================================
// END REFORMAT_DATE_TIME_AND_TICKETS TESTS
// ===============================================================================================================================

// ===============================================================================================================================
// UPDATE_TICKETS_SOLD_IN_TABLE TESTS
// ===============================================================================================================================

func TestUpdateTicketsSoldInTableResourceNotFound(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertResourceNotFound{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.UpdateTicketsSoldInTable("1234", 4)

	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

func TestUpdateTicketsSoldInTableNoConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientNoConcert{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.UpdateTicketsSoldInTable(concertID, 4)

	errMessage, ok := err.(databaseHandler.ErrConcertDoesNotExist)
	if !ok {
		t.Errorf("Expected ErrConcertDoesNotExist error, got %s\n", errMessage)
	}
}

func TestUpdateTicketsSoldInTableCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientConcertCannotUnmarshal{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.UpdateTicketsSoldInTable(concertID, 4)

	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)

	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

type mockDynamoDBClientUpdateConcertsFails struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientUpdateConcertsFails) GetItem(input *dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	epochTomorrow := time.Now().AddDate(0, 0, 1).Unix()
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetN(fmt.Sprint(epochTomorrow))
	item["TotalTickets"] = &dynamodb.AttributeValue{}
	item["TotalTickets"].SetN(fmt.Sprint(250))
	item["TicketsSold"] = &dynamodb.AttributeValue{}
	item["TicketsSold"].SetN(fmt.Sprint(50))
	item["FullPrice"] = &dynamodb.AttributeValue{}
	item["FullPrice"].SetN(fmt.Sprint(12.00))
	item["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	output.SetItem(item)
	return
}

func (m *mockDynamoDBClientUpdateConcertsFails) UpdateItem(input *dynamodb.UpdateItemInput) (output *dynamodb.UpdateItemOutput, err error) {
	err = &dynamodb.ResourceNotFoundException{}
	return
}

func TestUpdateTicketsSoldInTableUpdateFails(t *testing.T) {
	mockSvc := &mockDynamoDBClientUpdateConcertsFails{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.UpdateTicketsSoldInTable("ABC", 4)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T\n", expectedErr, err)
	}
}

type mockDynamoDBClientUpdateConcertsSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientUpdateConcertsSuccess) GetItem(input *dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	epochTomorrow := time.Now().AddDate(0, 0, 1).Unix()
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Title"] = &dynamodb.AttributeValue{}
	item["Title"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetN(fmt.Sprint(epochTomorrow))
	item["TotalTickets"] = &dynamodb.AttributeValue{}
	item["TotalTickets"].SetN(fmt.Sprint(250))
	item["TicketsSold"] = &dynamodb.AttributeValue{}
	item["TicketsSold"].SetN(fmt.Sprint(50))
	item["FullPrice"] = &dynamodb.AttributeValue{}
	item["FullPrice"].SetN(fmt.Sprint(12.00))
	item["ConcessionPrice"] = &dynamodb.AttributeValue{}
	item["ConcessionPrice"].SetN(fmt.Sprint(10.00))
	output.SetItem(item)
	return
}

func (m *mockDynamoDBClientUpdateConcertsSuccess) UpdateItem(input *dynamodb.UpdateItemInput) (output *dynamodb.UpdateItemOutput, err error) {
	return
}

func TestUpdateTicketsSoldInTableUpdateSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientUpdateConcertsSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	err := dynamoHandler.UpdateTicketsSoldInTable("ABC", 4)
	if err != nil {
		t.Errorf("Expected nil error, got %T\n", err)
	}
}

// ===============================================================================================================================
// END UPDATE_TICKETS_SOLD_IN_TABLE TESTS
// ===============================================================================================================================
