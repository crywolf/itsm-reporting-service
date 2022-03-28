package ticket

// Ticket domain object
type Ticket struct {
	UserEmail   string
	ChannelName string
	TicketType  string
	TicketData  Data
}

// List of tickets
type List []Ticket

// Data contain all relevant info about the ITSM ticket
type Data struct {
	Number           string
	ShortDescription string
}
