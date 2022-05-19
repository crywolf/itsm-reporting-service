package mocks

import (
	"context"
	"sync"

	"github.com/stretchr/testify/mock"
)

// EmailSenderMock is an email sender mock
type EmailSenderMock struct {
	mock.Mock
	Wg sync.WaitGroup
}

func (m *EmailSenderMock) SendEmailsForFieldEngineers(_ context.Context) error {
	defer m.Wg.Done()
	args := m.Called()
	return args.Error(0)
}

func (m *EmailSenderMock) SendEmailsForServiceDesk(_ context.Context) error {
	defer m.Wg.Done()
	args := m.Called()
	return args.Error(0)
}
