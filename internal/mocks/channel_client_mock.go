package mocks

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/stretchr/testify/mock"
)

// ChannelClientMock is a channel client mock
type ChannelClientMock struct {
	mock.Mock
}

func (m *ChannelClientMock) GetChannels(_ context.Context) (channel.List, error) {
	args := m.Called()
	return args.Get(0).(channel.List), args.Error(1)
}

func (m *ChannelClientMock) Close() error { return nil }
