package memory

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewTicketRepositoryMemory returns new initialized ticket repository that keeps data in memory
func NewTicketRepositoryMemory() repository.TicketRepository {
	return &ticketRepositoryMemory{}
}

type ticketRepositoryMemory struct {
	tickets []ticket.Ticket
}

func (r *ticketRepositoryMemory) AddTicketList(_ context.Context, ticketList ticket.List) error {
	for _, u := range ticketList {
		r.tickets = append(r.tickets, u)
	}

	return nil
}

func (r ticketRepositoryMemory) GetTicketsByEmail(_ context.Context, userEmail string) (ticket.List, error) {
	var list ticket.List

	for _, u := range r.tickets {
		if u.UserEmail == userEmail {
			list = append(list, u)
		}
	}

	return list, nil
}
