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

	channel1ID := "6abf417c-52e3-4340-9713-df2f37e78176"
	channel2ID := "75412c30-9f88-4b0e-a7c3-acfffe5f128b"

	inc1 := ticket.Ticket{
		UserEmail:   "first@user.com",
		ChannelID:   channel1ID,
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
		ChannelID:   channel2ID,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC999999",
			ShortDescription: "Inc 99999",
		},
	}
	inc3 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC55555",
			ShortDescription: "Inc 55555",
		},
	}
	req1 := ticket.Ticket{
		UserEmail:   "third@user.com",
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
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
		ChannelID:   channel2ID,
		TicketType:  "REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ22222",
			ShortDescription: "Req 22222",
		},
	}

	req3 := ticket.Ticket{
		UserEmail:   "third@user.com",
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
		TicketType:  "REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ6587456",
			ShortDescription: "Req 6587456",
		},
	}

	inc4 := ticket.Ticket{
		UserEmail:   email,
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC66666",
			ShortDescription: "Inc 66666",
		},
	}

	inc5 := ticket.Ticket{
		UserEmail:   "third@user.com",
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC357159",
			ShortDescription: "Inc 357159",
		},
	}

	incNoEmail := ticket.Ticket{
		UserEmail:   "", // Not Assigned Incident
		ChannelName: "Other Channel",
		ChannelID:   channel2ID,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC5874236",
			ShortDescription: "Not assigned incident",
		},
	}

	inc6 := ticket.Ticket{
		UserEmail:   "first@user.com",
		ChannelID:   channel2ID,
		ChannelName: "Other Channel",
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC32658",
			ShortDescription: "Inc 32658",
		},
	}

	list2 := ticket.List{req2, req3, inc4, inc5, incNoEmail, inc6}
	err = repo.AddTicketList(ctx, list2)
	require.NoError(t, err)

	// GetTicketsByEmailAddress
	retTicketListByEmail, err := repo.GetTicketsByEmailAddress(ctx, email)
	require.NoError(t, err)

	assert.Len(t, retTicketListByEmail, 4)
	// list should start with incidents
	assert.Equal(t, inc2, retTicketListByEmail[0])
	assert.Equal(t, inc3, retTicketListByEmail[1])
	assert.Equal(t, inc4, retTicketListByEmail[2])
	// requests should follow after incidents
	assert.Equal(t, req2, retTicketListByEmail[3])

	// GetTicketsByChannelID - channel 1
	retTicketListByChannel1, err := repo.GetTicketsByChannelID(ctx, channel1ID)
	require.NoError(t, err)

	assert.Len(t, retTicketListByChannel1, 1)
	assert.Equal(t, inc1, retTicketListByChannel1[0])

	// GetTicketsByChannelID - channel 2
	retTicketListByChannel2, err := repo.GetTicketsByChannelID(ctx, channel2ID)
	require.NoError(t, err)

	assert.Len(t, retTicketListByChannel2, 9)
	// list should start with incidents for user 1 and requests should follow after incidents
	assert.Equal(t, incNoEmail, retTicketListByChannel2[0])
	assert.Equal(t, inc6, retTicketListByChannel2[1])
	assert.Equal(t, inc2, retTicketListByChannel2[2])
	assert.Equal(t, inc3, retTicketListByChannel2[3])
	assert.Equal(t, inc4, retTicketListByChannel2[4])
	assert.Equal(t, req2, retTicketListByChannel2[5])

	// list should continue with incidents for user 2 and requests should follow after incidents
	assert.Equal(t, inc5, retTicketListByChannel2[6])
	assert.Equal(t, req1, retTicketListByChannel2[7])
	assert.Equal(t, req3, retTicketListByChannel2[8])

	// GetDistinctEmailAddresses
	retEmails, err := repo.GetDistinctEmailAddresses(ctx)
	require.NoError(t, err)

	assert.Len(t, retEmails, 3)

	assert.Equal(t, "first@user.com", retEmails[0])
	assert.Equal(t, "second@user.com", retEmails[1])
	assert.Equal(t, "third@user.com", retEmails[2])

	// GetDistinctChannelIDs
	retChannels, err := repo.GetDistinctChannelIDs(ctx)
	require.NoError(t, err)

	assert.Len(t, retChannels, 2)
	assert.Equal(t, channel2ID, retChannels[0])
	assert.Equal(t, channel1ID, retChannels[1])
}
