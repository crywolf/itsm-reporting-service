package mocks

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/stretchr/testify/mock"
)

// UserClientMock is a user client mock
type UserClientMock struct {
	mock.Mock
}

func (m *UserClientMock) GetEngineers(_ context.Context, channel channel.Channel) (user.List, error) {
	args := m.Called(channel)
	return args.Get(0).(user.List), args.Error(1)
}

func (m *UserClientMock) Close() error { return nil }
