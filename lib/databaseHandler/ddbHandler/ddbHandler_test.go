package ddbHandler

import (
	"fmt"
	//	"os"
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

/*
	func TestMain(m *testing.M) {
		rc := m.Run()

		if rc == 0 && testing.CoverMode() != "" {
			c := testing.Coverage()
			fmt.Println(c)
			if c < 0.9 {
				fmt.Printf("Tests passed but coverage was below %d%%\n", int(c*100))
				rc = -1
			}
		}
		os.Exit(rc)
	}
*/
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

type mockDynamoDBClientSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientSuccess) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["ID"] = &dynamodb.AttributeValue{}
	item1["ID"].SetS("AAA")
	item1["Description"] = &dynamodb.AttributeValue{}
	item1["Description"].SetS("Summer Concert")
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
	item2["Description"] = &dynamodb.AttributeValue{}
	item2["Description"].SetS("Winter Concert")
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

func (m *mockDynamoDBClientSuccess) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	epochTomorrow := time.Now().AddDate(0, 0, 1).Unix()
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Description"] = &dynamodb.AttributeValue{}
	item["Description"].SetS("Summer Concert")
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

func TestGetConcertsFromDynamoDBSuccessful(t *testing.T) {
	expectedNumConcerts := 2
	mockSvc := &mockDynamoDBClientSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concerts, err := dynamoHandler.GetConcertsFromDatabase()
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err)
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d\n", expectedNumConcerts, len(concerts))
	}

	firstConcertDescription := "Summer Concert"
	secondConcertDescription := "Winter Concert"

	if concerts[0].Description != firstConcertDescription {
		t.Errorf("Expected first concert returned to be %s, got %s\n", firstConcertDescription, concerts[0].Description)
	}

	if concerts[1].Description != secondConcertDescription {
		t.Errorf("Expected second concert returned to be %s, got %s\n", secondConcertDescription, concerts[1].Description)
	}

	for _, v := range concerts {
		if v.DateTime != nil || v.TotalTickets != nil || v.TicketsSold != nil {
			t.Error("DateTime, TotalTikets and TicketsSold should all be nil")
		}
	}
}

type mockDynamoDBClientNoConcerts struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientNoConcerts) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 0
	items := make([]map[string]*dynamodb.AttributeValue, 0)
	numConcertsI64 := int64(numConcerts)
	output = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func (m *mockDynamoDBClientNoConcerts) GetItem(*dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	cc := dynamodb.ConsumedCapacity{}
	output.SetConsumedCapacity(&cc)
	output.SetItem(nil)
	return
}
func TestGetConcertsFromDynamoDBNoConcerts(t *testing.T) {
	expectedNumConcerts := 0
	mockSvc := &mockDynamoDBClientNoConcerts{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concerts, err := dynamoHandler.GetConcertsFromDatabase()
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err.Error())
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d\n", expectedNumConcerts, len(concerts))
	}
}

type mockDynamoDBClientResourceNotFound struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientResourceNotFound) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	err = &dynamodb.ResourceNotFoundException{}

	return
}

func (m *mockDynamoDBClientResourceNotFound) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	err := &dynamodb.ResourceNotFoundException{}
	return nil, err
}

func TestGetConcertsFromDynamoDBCannotAccessTable(t *testing.T) {
	mockSvc := &mockDynamoDBClientResourceNotFound{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromDatabase()
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected %s error type, got %s\n", expectedErr, err)
	}
}

func TestGetConcertFromDynamoDBSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientSuccess{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	concert, err := dynamoHandler.GetConcertFromDatabase(concertID)
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err.Error())
	}

	if concert.ID != concertID {
		t.Errorf("Expected entry with ID %s, got %s\n", concertID, concert.ID)
	}
	if concert.DateTime != nil || concert.TotalTickets != nil || concert.TicketsSold != nil {
		t.Error("DateTime, TotalTikets and TicketsSold should all be nil")
	}
}

func TestGetConcertFromDynamoDBCannotAccessTable(t *testing.T) {
	mockSvc := &mockDynamoDBClientResourceNotFound{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromDatabase(concertID)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected %s error type, got %s\n", expectedErr, err)
	}
}

func TestGetConcertFromDynamoDBNoConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientNoConcerts{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromDatabase(concertID)

	errMessage, ok := err.(databaseHandler.ErrConcertDoesNotExist)
	if !ok {
		t.Errorf("Expected ErrConcertDoesNotExist error, got %s\n", errMessage)
	}
}

type mockDynamoDBClientInvalidData struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientInvalidData) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	epochYesterday := time.Now().AddDate(0, 0, -1).Unix()
	output := dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Description"] = &dynamodb.AttributeValue{}
	item["Description"].SetS("Summer Concert")
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

func (m *mockDynamoDBClientInvalidData) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["ID"] = &dynamodb.AttributeValue{}
	item1["ID"].SetS("AAA")
	item1["Description"] = &dynamodb.AttributeValue{}
	item1["Description"].SetS("Summer Concert")
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
	item2["Description"] = &dynamodb.AttributeValue{}
	item2["Description"].SetS("Winter Concert")
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

func TestGetConcertFromDynamoDBMissingTicketPrices(t *testing.T) {
	mockSvc := &mockDynamoDBClientInvalidData{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromDatabase(concertID)
	expectedErr, ok := err.(databaseHandler.ErrInvalidConcertData)
	if !ok {
		t.Errorf("Expected %v error, got %v\n", expectedErr.Error(), err.Error())
	}
}

func TestGetConcertsFromDynamoDBMissingDateTime(t *testing.T) {
	mockSvc := &mockDynamoDBClientInvalidData{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromDatabase()
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
	item["Description"] = &dynamodb.AttributeValue{}
	item["Description"].SetS("Summer Concert")
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

func TestGetConcertFromDynamoDBOldConcert(t *testing.T) {
	mockSvc := &mockDynamoDBClientOldConcert{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromDatabase(concertID)

	expectedErr, ok := err.(databaseHandler.ErrConcertInPast)

	if !ok {
		t.Errorf("Expected %s error, got %s\n", expectedErr, err)
	}
}

type mockDynamoDBClientCannotUnmarshal struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientCannotUnmarshal) GetItem(*dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ID"] = &dynamodb.AttributeValue{}
	item["ID"].SetS("AAA")
	item["Description"] = &dynamodb.AttributeValue{}
	item["Description"].SetS("Summer Concert")
	item["ImageURL"] = &dynamodb.AttributeValue{}
	item["ImageURL"].SetS("http://example.com/image.jpg")
	item["DateTime"] = &dynamodb.AttributeValue{}
	item["DateTime"].SetS("Hello")
	output.SetItem(item)
	return
}

func (m *mockDynamoDBClientCannotUnmarshal) Scan(input *dynamodb.ScanInput) (output *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["Description"] = &dynamodb.AttributeValue{}
	item1["Description"].SetS("Summer Concert")
	item1["ImageURL"] = &dynamodb.AttributeValue{}
	item1["ImageURL"].SetS("http://example.com/image.jpg")
	item1["DateTime"] = &dynamodb.AttributeValue{}
	item1["DateTime"].SetS("Hello")
	items = append(items, item1)
	item2 := map[string]*dynamodb.AttributeValue{}
	item2["Description"] = &dynamodb.AttributeValue{}
	item2["Description"].SetS("Winter Concert")
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

func TestGetConcertFromDynamoDBCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientCannotUnmarshal{}
	concertID := "AAA"
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertFromDatabase(concertID)

	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)

	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

func TestGetConcertsFromDynamoDBCannotUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientCannotUnmarshal{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetConcertsFromDatabase()

	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)

	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

func TestGenerateOrderReference(t *testing.T) {
	var size uint8 = 4
	result := generateOrderReference(size)
	if uint8(len(result)) != size {
		t.Errorf("Expected reference of size %v, got %v", size, len(result))
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

func TestGetOrderDetailsFails(t *testing.T) {
	mockSvc := &mockDynamoDBClientResourceNotFound{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderDetails("1234", "A1B2")
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)
	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

func TestGetOrderDetailsNoOrder(t *testing.T) {
	mockSvc := mockDynamoDBClientNoOrder{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderDetails("1234", "A1B2")
	expectedErr, ok := err.(paymentHandler.ErrOrderDoesNotExist)
	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

type mockDynamoDBClientGetOrderSuccess struct {
	dynamodbiface.DynamoDBAPI
}

func (m mockDynamoDBClientGetOrderSuccess) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
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

func TestGetOrderDetailsSuccess(t *testing.T) {
	mockSvc := &mockDynamoDBClientGetOrderSuccess{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order, err := dynamoHandler.GetOrderDetails("1234", "A1B2")
	if err != nil {
		t.Errorf("Expected no error, got %T\n", err)
	}
	expectedConcertId := "1234"
	expectedReference := "A1B2"
	if order.ConcertId != expectedConcertId || order.Reference != expectedReference {
		t.Errorf("Expected concertId of %s and order reference of %s, got %s and %s\n", expectedConcertId, expectedReference, order.ConcertId, order.Reference)
	}
}

type mockDynamoDBClientGetOrderCannotUnmarshal struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientGetOrderCannotUnmarshal) GetItem(input *dynamodb.GetItemInput) (output *dynamodb.GetItemOutput, err error) {
	output = &dynamodb.GetItemOutput{}
	item := map[string]*dynamodb.AttributeValue{}
	item["ConcertId"] = &dynamodb.AttributeValue{}
	item["ConcertId"].SetBOOL(true)
	output.SetItem(item)
	return
}
func TestGetOrderUnableToUnmarshal(t *testing.T) {
	mockSvc := &mockDynamoDBClientGetOrderCannotUnmarshal{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	_, err := dynamoHandler.GetOrderDetails("1234", "A1B2")
	expectedErr, ok := err.(*dynamodbattribute.UnmarshalTypeError)
	if !ok {
		t.Errorf("Expected err %s, got %s\n", expectedErr, err)
	}
}

type mockDynamoDBCannotPut struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBCannotPut) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, &dynamodb.ResourceNotFoundException{}
}

func TestCreateOrderEntryCannotPut(t *testing.T) {
	mockSvc := &mockDynamoDBCannotPut{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.createOrderEntry(order)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}

type mockDynamoDBClientReferenceMatchOnce struct {
	dynamodbiface.DynamoDBAPI
	firstCall bool
}

func (m *mockDynamoDBClientReferenceMatchOnce) PutItem(input *dynamodb.PutItemInput) (output *dynamodb.PutItemOutput, err error) {
	if m.firstCall {
		err = &dynamodb.ConditionalCheckFailedException{}
	}
	m.firstCall = false
	return
}

func TestCreateEntryInOrdersDatabaseReferenceMatchOnce(t *testing.T) {
	mockSvc := &mockDynamoDBClientReferenceMatchOnce{firstCall: true}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.CreateEntryInOrdersDatabase(order)
	if err != nil {
		t.Errorf("Expected nil err, got %T", err)
	}
}

func TestCreateEntryInOrdersDatabaseFails(t *testing.T) {
	mockSvc := &mockDynamoDBCannotPut{}
	dynamoHandler := New(mockSvc, "concerts-table", "orders-table")
	order := paymentHandler.Order{}
	err := dynamoHandler.CreateEntryInOrdersDatabase(order)
	expectedErr, ok := err.(*dynamodb.ResourceNotFoundException)

	if !ok {
		t.Errorf("Expected error of type %T, got %T", expectedErr, err)
	}
}
