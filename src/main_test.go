package main

import (
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

func TestGetConcertsFromDynamoDBSucessful(t *testing.T) {
	
