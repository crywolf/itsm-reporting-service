package ticket

import (
	"fmt"
	"time"
)

// Ticket domain object
type Ticket struct {
	UserID      string
	UserEmail   string
	UserName    string
	UserOrgName string
	ChannelID   string
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
	StateID          int
	Location         string
	CreatedAt        string
}

// StateName converts state ID to its name
func (d Data) StateName() string {
	states := [...]string{
		"New",         // 0
		"On Hold",     // 1
		"In progress", // 2
		"On Hold",     // 3
		"Resolved",    // 4
		"Closed",      // 5
		"Cancelled",   // 6
	}

	if d.StateID >= len(states) {
		return "Unknown"
	}

	return states[d.StateID]
}

// CreatedAtDate returns date of the ticket creation
func (d Data) CreatedAtDate() string {
	datetime, err := time.Parse(time.RFC3339, d.CreatedAt)
	if err != nil {
		return err.Error()
	}

	year, month, day := datetime.UTC().Date()

	return fmt.Sprintf("%d/%d/%d", day, month, year)
}
