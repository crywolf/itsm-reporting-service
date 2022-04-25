package memory

import (
	"context"
	"sort"
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

	for _, t := range ticketList {
		r.tickets = append(r.tickets, t)
	}

	return nil
}

func (r *ticketRepositoryMemory) GetTicketsByEmail(_ context.Context, userEmail string) (ticket.List, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var list ticket.List

	// first will be Incidents, then Requests
	sort.Slice(r.tickets, func(i, j int) bool {
		return r.tickets[i].TicketType < r.tickets[j].TicketType
	})

	for _, t := range r.tickets {
		if t.UserEmail == userEmail {
			list = append(list, t)
		}
	}

	return list, nil
}

func (r *ticketRepositoryMemory) GetDistinctEmails(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var emails []string
	emailsMap := make(map[string]bool)

	for _, t := range r.tickets {
		_, found := emailsMap[t.UserEmail]
		if found {
			continue
		}
		emailsMap[t.UserEmail] = true
	}

	for email := range emailsMap {
		emails = append(emails, email)
	}

	return emails, nil
}

func (r *ticketRepositoryMemory) Truncate(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tickets = nil

	return nil
}
