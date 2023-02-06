package databaseHandler

import (
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

// Concert is a model of a concert which contains basic info regarding a concert, taken from dynamoDB
type Concert struct {
	ID               string  `json:"id"`
	Description      string  `json:"description"`
	ImageURL         string  `json:"imageURL"`
	DateTime         *int64  `json:"dateTime,omitempty"`
	Date             string  `json:"date"`
	Time             string  `json:"time"`
	TotalTickets     *uint8  `json:"totalTickets,omitempty"`
	TicketsSold      *uint8  `json:"ticketsSold,omitempty"`
	AvailableTickets uint8   `json:"availableTickets"`
	FullPrice        float32 `json:"fullPrice"`
	ConcessionPrice  float32 `json:"concessionPrice"`
}

type DatabaseHandler interface {
	GetConcertFromDatabase(concertID string) (concert *Concert, err error)
	GetConcertsFromDatabase() (concerts []Concert, err error)
	CreateEntryInOrdersDatabase(order paymentHandler.Order) (err error)
	GetOrderDetails(concertId string, ref string) (order *paymentHandler.Order, err error)
}
