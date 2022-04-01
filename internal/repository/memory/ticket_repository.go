package memory

import (
	"context"
	"sync"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewTicketRepositoryMemory returns new initialized ticket repository that keeps data in memory
func NewTicketRepositoryMemory() repository.TicketRepository {
	return &ticketRepositoryMemory{}
}

type ticketRepositoryMemory struct {
	tickets []ticket.Ticket
	mu      sync.Mutex
}

func (r *ticketRepositoryMemory) AddTicketList(_ context.Context, ticketList ticket.List) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range ticketList {
		r.tickets = append(r.tickets, u)
	}

	return nil
}

func (r *ticketRepositoryMemory) GetTicketsByEmail(_ context.Context, userEmail string) (ticket.List, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var list ticket.List

	for _, t := range r.tickets {
		if t.UserEmail == userEmail {
			list = append(list, t)
		}
	}

	return list, nil
}

func (r *ticketRepositoryMemory) Truncate(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tickets = nil

	return nil
}
