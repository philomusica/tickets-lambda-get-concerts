package ddbHandler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/philomusica/tickets-lambda-basket-service/lib/paymentHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
)

// ===============================================================================================================================
// TYPE DEFINITIONS
// ===============================================================================================================================
type DDBHandler struct {
	svc           dynamodbiface.DynamoDBAPI
	concertsTable string
	ordersTable   string
}

// ===============================================================================================================================
// END OF TYPE DEFINITIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PRIVATE FUNCTIONS
// ===============================================================================================================================

// convertEpochSecsToDateAndTimeStrings takes a dateTime int64 and returns two strings, date (formatted as Mon 2 Jan 2006) and timeStamp (formatted as 3:04 PM)
func convertEpochSecsToDateAndTimeStrings(dateTime int64) (date string, timeStamp string) {
	t := time.Unix(dateTime, 0)
	date = t.Format("Mon 2 Jan 2006")
	timeStamp = t.Format("3:04 PM")
	return
}

// generateRandomString takes a uint8 indicating the num of random characters to generate, and returns the random in the form of a string
func generateRandomString(size uint8) string {
	charSet := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.Seed(time.Now().UnixNano())
	arr := make([]byte, size)
	var i uint8
	for i = 0; i < size; i++ {
		arr[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(arr)
}

// validateConcert takes a pointer to a databaseHandler.Concert struct and checks the validity of the members
func validateConcert(c *databaseHandler.Concert) (valid bool) {
	valid = false

	if c.ID != "" && c.Title != "" && c.ImageURL != "" &&
		c.DateTime != nil && *c.DateTime > 0 && c.TotalTickets != nil && *c.TotalTickets > 0 &&
		c.TicketsSold != nil && c.FullPrice > 0.0 {
		valid = true
	}
	return
}

// ===============================================================================================================================
// END OF PRIVATE FUNCTIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

// CreateOrderInTable takes the paymentRequest struct, generates a new payment references, checks for uniqueness and creates an entry in the orders table. Returns an error if it fails at any point
func (d DDBHandler) CreateOrderInTable(order paymentHandler.Order) (err error) {
	av, err := dynamodbattribute.MarshalMap(order)
	if err != nil {
		return
	}

	_, err = d.svc.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(d.ordersTable),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(orderReference) AND attribute_not_exists(concertID)"),
	})
	return
}

// GenerateOrderReference takes a uint8 indicating the num of random characters to generate. In an infinite loop it calls the generateRandomString function, check if the returned string is unique in the dynamoDB table and breaks and returns the random in the form of a string once a unique string is found
func (d DDBHandler) GenerateOrderReference(size uint8) (ref string) {
	for {
		ref = generateRandomString(size)
		orders, err := d.GetOrdersByOrderReferenceFromTable(ref)
		if len(orders) == 0 || err != nil {
			break
		}
	}
	return
}

// GetConcertFromTable retrieves a specific concert from the dynamoDB table, returns a pointer to a databaseHandler.Concert struct and error (nil if successful).
func (d DDBHandler) GetConcertFromTable(concertID string) (concert *databaseHandler.Concert, err error) {
	concert = &databaseHandler.Concert{}
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.concertsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(concertID),
			},
		},
	})
	if err != nil {
		return
	} else if result.Item == nil {
		err = databaseHandler.ErrConcertDoesNotExist{Message: "Error does not exist"}
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, concert)
	if err != nil {
		return
	}

	if !validateConcert(concert) {
		err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Invalid concert data for concert %s\n", concertID)}
		return
	}

	epochNow := time.Now().Unix()
	if *concert.DateTime < epochNow {
		err = databaseHandler.ErrConcertInPast{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", concertID)}
		return
	}

	return
}

// GetConcertsFromTable gets all upcoming concerts from the dynamoDB table, returning a slice of databaseHandler.Concert structs and an error (nil if successful).
func (d DDBHandler) GetConcertsFromTable() (concerts []databaseHandler.Concert, err error) {
	epochNow := time.Now().Unix()
	filt := expression.Name("DateTime").GreaterThan(expression.Value(epochNow))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return
	}

	result, err := d.svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(d.concertsTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		return
	}

	fmt.Println(result)

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &concerts)
	if err != nil {
		return
	}

	for _, v := range concerts {
		if !validateConcert(&v) {
			err = databaseHandler.ErrInvalidConcertData{Message: fmt.Sprintf("Error concert %s in the past, tickets are no longer available", v.ID)}
			return
		}
	}

	return
}

// GetOrderFromTable takes a reference id and returns an paymentHandler.Order struct, or nil if the order does not exist. The second return type is error which will be nil if successful or not nil if an error occur retriving the entry
func (d DDBHandler) GetOrderFromTable(concertId string, ref string) (order *paymentHandler.Order, err error) {
	order = &paymentHandler.Order{}
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.ordersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"orderReference": {
				S: aws.String(ref),
			},
			"concertID": {
				S: aws.String(concertId),
			},
		},
	})
	if err != nil {
		return
	} else if result.Item == nil {
		err = paymentHandler.ErrOrderDoesNotExist{Message: "Error does not exist"}
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, order)
	if err != nil {
		return
	}
	return
}

// GetOrdersByOrderReferenceFromTable takes an order reference and return a slice of orders, or an error if this failed
func (d DDBHandler) GetOrdersByOrderReferenceFromTable(ref string) (orders []paymentHandler.Order, err error) {
	filt := expression.Name("orderReference").Equal(expression.Value(ref))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return
	}
	result, err := d.svc.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(d.ordersTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		return
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &orders)
	return
}

// New takes the AWS DynamoDBAPI interface, the name of the concerts and orders tables (both strings) and returns a newly created DDBHandler struct
func New(svc dynamodbiface.DynamoDBAPI, concertsTable string, ordersTable string) DDBHandler {
	return DDBHandler{
		svc,
		concertsTable,
		ordersTable,
	}
}

// ReformatDateTimeAndTickets takes a pointer to a databaseHandler.Concert struct, modifying it in-place to convert DateTime epoch into a date and time string, and converts num of tickets sold into num of tickets available. Returns an error if nil is passed
func (d DDBHandler) ReformatDateTimeAndTickets(concert *databaseHandler.Concert) (err error) {
	if concert == nil {
		err = databaseHandler.ErrConcertDoesNotExist{Message: "Nil value passed to reformater"}
		return
	}

	dateStr, timeStr := convertEpochSecsToDateAndTimeStrings(*concert.DateTime)
	concert.Date = dateStr
	concert.Time = timeStr
	concert.DateTime = nil
	concert.AvailableTickets = *concert.TotalTickets - *concert.TicketsSold
	concert.TotalTickets = nil
	concert.TicketsSold = nil
	return
}

// UpdateOrderInTable takes a concertID and payment reference and new status value, fetches the order from DynamoDB, and updates the order with the new status value. Returns an error if unsuccessful, or nil if successful
func (d DDBHandler) UpdateOrderInTable(concertID string, reference string, newStatus string) (err error) {
	order, err := d.GetOrderFromTable(concertID, reference)
	if err != nil || order == nil {
		return
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":ns": {
				S: aws.String(newStatus),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"orderReference": {
				S: aws.String(reference),
			},
			"concertID": {
				S: aws.String(concertID),
			},
		},
	}

	_, err = d.svc.UpdateItem(input)
	return
}

// UpdateTicketsSoldInTable takes the concertID and the number of tickets sold, fetches the concert from DynamoDB, then increments the ticketsSold field with the provided parameter
func (d DDBHandler) UpdateTicketsSoldInTable(concertID string, ticketsSold uint16) (err error) {
	concert, err := d.GetConcertFromTable(concertID)
	if err != nil {
		return
	}

	ticketsSoldUpdated := *concert.TicketsSold + ticketsSold

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":ts": {
				N: aws.String(fmt.Sprint(ticketsSoldUpdated)),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(concert.ID),
			},
		},
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(d.concertsTable),
		UpdateExpression: aws.String("set TicketsSold = :ts"),
	}

	_, err = d.svc.UpdateItem(input)
	return
}

// ===============================================================================================================================
// END OF PUBLIC FUNCTIONS
// ===============================================================================================================================
