package mocks

import (
	"context"
	"sync"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/stretchr/testify/mock"
)

// TicketClientMock is a ticket client mock
type TicketClientMock struct {
	mock.Mock
	Wg sync.WaitGroup
}

func (m *TicketClientMock) GetIncidents(_ context.Context, channel channel.Channel, user user.User) (ticket.List, error) {
	defer m.Wg.Done()
	args := m.Called(channel, user)
	return args.Get(0).(ticket.List), args.Error(1)
}

func (m *TicketClientMock) GetRequests(_ context.Context, channel channel.Channel, user user.User) (ticket.List, error) {
	defer m.Wg.Done()
	args := m.Called(channel, user)
	return args.Get(0).(ticket.List), args.Error(1)
}

func (m *TicketClientMock) Close() error { return nil }
