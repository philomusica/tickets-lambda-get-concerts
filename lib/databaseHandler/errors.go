package databaseHandler

// ErrConcertInPast is a custom error message to signify concert is in past and tickets can no longer be purchased for it
type ErrConcertInPast struct {
	Message string
}

func (e ErrConcertInPast) Error() string {
	return e.Message
}

// ErrConcertDoesNotExist is a custom error message to signify the concert with a given ID does not exist
type ErrConcertDoesNotExist struct {
	Message string
}

func (e ErrConcertDoesNotExist) Error() string {
	return e.Message
}

// ErrInvalidConcertData is a custom error message to signify the data from dynamoDB that has been unmarshalled into a struct is incomplete
type ErrInvalidConcertData struct {
	Message string
}

func (e ErrInvalidConcertData) Error() string {
	return e.Message
}
