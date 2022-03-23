package mocks

import (
	"context"
	"sync"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/stretchr/testify/mock"
)

// ChannelRepositoryMock is a channel repository mock
type ChannelRepositoryMock struct {
	mock.Mock
	Wg sync.WaitGroup
}

func (m *ChannelRepositoryMock) StoreChannelList(_ context.Context, channelList channel.List) error {
	defer m.Wg.Done()
	args := m.Called(channelList)
	return args.Error(0)
}

func (m *ChannelRepositoryMock) GetChannelList(_ context.Context) (channel.List, error) {
	args := m.Called()
	return args.Get(0).(channel.List), args.Error(1)
}
