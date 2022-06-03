package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"testing"
	"time"
)

var summerEpoch int64 = 1656176400 // 25/06/22 18:00
var winterEpoch int64 = 1671991200 // 25/12/22 18:00

func TestConvertEpochSecsToDateAndTimeStringsDateValueSummer(t *testing.T) {
	timeStamp := time.Unix(summerEpoch, 0)
	expectedDate := timeStamp.Format("Mon 2 Jan 2006")
	result, _ := ConvertEpochSecsToDateAndTimeStrings(summerEpoch)

	if result != expectedDate {
		t.Errorf("Expected %s, got %s", expectedDate, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsTimeValueSummer(t *testing.T) {
	timeStamp := time.Unix(summerEpoch, 0)
	expectedTime := timeStamp.Format("3:04 PM")
	_, result := ConvertEpochSecsToDateAndTimeStrings(summerEpoch)
	if result != expectedTime {
		t.Errorf("Expected %s, got %s", expectedTime, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsDateValueWinter(t *testing.T) {
	timeStamp := time.Unix(winterEpoch, 0)
	expectedDate := timeStamp.Format("Mon 2 Jan 2006")
	result, _ := ConvertEpochSecsToDateAndTimeStrings(winterEpoch)

	if result != expectedDate {
		t.Errorf("Expected %s, got %s", expectedDate, result)
	}
}

func TestConvertEpochSecsToDateAndTimeStringsTimeValueWinter(t *testing.T) {
	timeStamp := time.Unix(winterEpoch, 0)
	expectedTime := timeStamp.Format("3:04 PM")
	_, result := ConvertEpochSecsToDateAndTimeStrings(winterEpoch)
	if result != expectedTime {
		t.Errorf("Expected %s, got %s", expectedTime, result)
	}
}

type mockDynamoDBClientSuccess struct {
	dynamodbiface.DynamoDBAPI
}
func (m *mockDynamoDBClientSuccess) Scan(input *dynamodb.ScanInput) (response *dynamodb.ScanOutput, err error) {
	numConcerts := 2
	items := make([]map[string]*dynamodb.AttributeValue, 0, numConcerts)
	item1 := map[string]*dynamodb.AttributeValue{}
	item1["Description"] = &dynamodb.AttributeValue{}
	item1["Description"].SetS("Summer Concert")
	item1["ImageURL"] = &dynamodb.AttributeValue{}
	item1["ImageURL"].SetS("http://example.com/image.jpg")
	item1["ConcertDateTime"] = &dynamodb.AttributeValue{}
	item1["ConcertDateTime"].SetN(fmt.Sprintf("%d", summerEpoch))
	items = append(items, item1)
	item2 := map[string]*dynamodb.AttributeValue{}
	item2["Description"] = &dynamodb.AttributeValue{}
	item2["Description"].SetS("Winter Concert")
	item2["ImageURL"] = &dynamodb.AttributeValue{}
	item2["ImageURL"].SetS("http://example.com/image2.jpg")
	item2["ConcertDateTime"] = &dynamodb.AttributeValue{}
	item2["ConcertDateTime"].SetN(fmt.Sprintf("%d", winterEpoch))
	items = append(items, item2)
	numConcertsI64 := int64(numConcerts)
	response = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func TestGetConcertsFromDynamoDBSucessful(t *testing.T) {
	expectedNumConcerts := 2
	concerts := make([]Concert, 0, expectedNumConcerts)
	mockSvc := &mockDynamoDBClientSuccess{}
	err := GetConcertsFromDynamoDB(mockSvc, &concerts)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d", expectedNumConcerts, len(concerts))
	}

	firstConcertDescription := "Summer Concert"
	secondConcertDescription := "Winter Concert"

	if concerts[0].Description != firstConcertDescription {
		t.Errorf("Expected first concert returned to be %s, got %s", firstConcertDescription, concerts[0].Description)
	}

	if concerts[1].Description != secondConcertDescription {
		t.Errorf("Expected second concert returned to be %s, got %s", secondConcertDescription, concerts[1].Description)
	}
}

type mockDynamoDBClientNoConcerts struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientNoConcerts) Scan(input *dynamodb.ScanInput) (response *dynamodb.ScanOutput, err error) {
	numConcerts := 0
	items := make([]map[string]*dynamodb.AttributeValue, 0)
	numConcertsI64 := int64(numConcerts)
	response = &dynamodb.ScanOutput{
		Count: &numConcertsI64,
		Items: items,
	}
	return
}

func TestGetConcertsFromDynamoDBNoConcerts(t *testing.T) {
	expectedNumConcerts := 0
	concerts := make([]Concert, 0, expectedNumConcerts)
	mockSvc := &mockDynamoDBClientNoConcerts{}
	err := GetConcertsFromDynamoDB(mockSvc, &concerts)
	if err != err.(ErrNoConcerts) {
		t.Errorf("Expected ErrNoConcerts type , got %s", err)
	}

	if len(concerts) != expectedNumConcerts {
		t.Errorf("Expected %d concerts returned, got %d", expectedNumConcerts, len(concerts))
	}
}

type mockDynamoDBClientResourceNotFound struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClientResourceNotFound ) Scan(input *dynamodb.ScanInput) (response *dynamodb.ScanOutput, err error) {
	err = &dynamodb.ResourceNotFoundException{}
	
	return
}

func TestGetConcertsFromDynamoDBCannotAccessTable(t *testing.T) {
	concerts := make([]Concert, 0)
	mockSvc := &mockDynamoDBClientResourceNotFound {}
	err := GetConcertsFromDynamoDB(mockSvc, &concerts)
	if err != err.(*dynamodb.ResourceNotFoundException) {
		t.Errorf("Expected %s error type, got %s", "ResourceNotFoundException", err)
	}
}