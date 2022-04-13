package memory

import (
	"context"
	"sort"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketRepositoryMemory_AddingAndGettingTickets(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepositoryMemory()

	email := "second@user.com"

	inc1 := ticket.Ticket{
		UserEmail:   "first@user.com",
		ChannelName: "Some Channel",
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC123456",
			ShortDescription: "Inc 1",
		},
	}
	inc2 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC999999",
			ShortDescription: "Inc 99999",
		},
	}
	inc3 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC55555",
			ShortDescription: "Inc 55555",
		},
	}
	req1 := ticket.Ticket{
		UserEmail:   "third@user.com",
		ChannelName: "Other Channel",
		TicketType:  "REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ987456",
			ShortDescription: "Req 987456",
		},
	}

	list := ticket.List{inc1, inc2, inc3, req1}
	err := repo.AddTicketList(ctx, list)
	require.NoError(t, err)

	req2 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		TicketType:  "REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ22222",
			ShortDescription: "Req 22222",
		},
	}

	inc4 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC66666",
			ShortDescription: "Inc 66666",
		},
	}

	list2 := ticket.List{req2, inc4}
	err = repo.AddTicketList(ctx, list2)
	require.NoError(t, err)

	// GetTicketsByEmail
	retTicketList, err := repo.GetTicketsByEmail(ctx, email)
	require.NoError(t, err)

	assert.Len(t, retTicketList, 4)
	// list should start with incidents
	assert.Equal(t, inc2, retTicketList[0])
	assert.Equal(t, inc3, retTicketList[1])
	assert.Equal(t, inc4, retTicketList[2])
	// requests should follow after incidents
	assert.Equal(t, req2, retTicketList[3])

	// GetDistinctEmails
	retEmails, err := repo.GetDistinctEmails(ctx)
	require.NoError(t, err)

	assert.Len(t, retEmails, 3)
	sort.Strings(retEmails) // sort returned emails to enable easy testing
	assert.Equal(t, "first@user.com", retEmails[0])
	assert.Equal(t, "second@user.com", retEmails[1])
	assert.Equal(t, "third@user.com", retEmails[2])
}
