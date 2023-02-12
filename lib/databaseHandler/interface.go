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
	TotalTickets     *uint16 `json:"totalTickets,omitempty"`
	TicketsSold      *uint16 `json:"ticketsSold,omitempty"`
	AvailableTickets uint16  `json:"availableTickets"`
	FullPrice        float32 `json:"fullPrice"`
	ConcessionPrice  float32 `json:"concessionPrice"`
}

type DatabaseHandler interface {
	CreateOrderInTable(order paymentHandler.Order) (err error)
	GetConcertFromTable(concertID string) (concert *Concert, err error)
	GetConcertsFromTable() (concerts []Concert, err error)
	GetOrderFromTable(concertId string, ref string) (order *paymentHandler.Order, err error)
	ReformatDateTimeAndTickets(concert *Concert) (err error)
	UpdateTicketsSoldInTable(concertID string, ticketsSold uint16) (err error)
}
