package memory

import (
	"context"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketRepositoryMemory_AddingAndGettingTickets(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepositoryMemory()

	email := "second@user.com"

	t1 := ticket.Ticket{
		UserEmail:   "first@user.com",
		ChannelName: "Some Channel",
		TicketType:  "incident",
		TicketData: ticket.Data{
			Number:           "INC123456",
			ShortDescription: "Inc 1",
		},
	}
	t2 := ticket.Ticket{
		UserEmail:   "second@user.com",
		ChannelName: "Other Channel",
		TicketType:  "incident",
		TicketData: ticket.Data{
			Number:           "INC999999",
			ShortDescription: "Inc 99999",
		},
	}
	t3 := ticket.Ticket{
		UserEmail:   "second@user.com",
		ChannelName: "Other Channel",
		TicketType:  "incident",
		TicketData: ticket.Data{
			Number:           "INC55555",
			ShortDescription: "Inc 55555",
		},
	}
	t4 := ticket.Ticket{
		UserEmail:   "third@user.com",
		ChannelName: "Other Channel",
		TicketType:  "request",
		TicketData: ticket.Data{
			Number:           "REQ987456",
			ShortDescription: "Req 987456",
		},
	}

	list := ticket.List{t1, t2, t3, t4}
	err := repo.AddTicketList(ctx, list)
	require.NoError(t, err)

	t5 := ticket.Ticket{
		UserEmail:   "second@user.com",
		ChannelName: "Other Channel",
		TicketType:  "incident",
		TicketData: ticket.Data{
			Number:           "INC66666",
			ShortDescription: "Inc 66666",
		},
	}

	list2 := ticket.List{t5}
	err = repo.AddTicketList(ctx, list2)
	require.NoError(t, err)

	retTicketList, err := repo.GetTicketsByEmail(ctx, email)
	require.NoError(t, err)

	assert.Len(t, retTicketList, 3)
	assert.Equal(t, t2, retTicketList[0])
	assert.Equal(t, t3, retTicketList[1])
	assert.Equal(t, t5, retTicketList[2])
}
