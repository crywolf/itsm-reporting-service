package mocks

import (
	"context"
	"sync"

	"github.com/stretchr/testify/mock"
)

// TicketDownloaderMock is a ticket downloader mock
type TicketDownloaderMock struct {
	mock.Mock
	Wg sync.WaitGroup
}

func (m *TicketDownloaderMock) DownloadTickets(_ context.Context) error {
	defer m.Wg.Done()
	args := m.Called()
	return args.Error(0)
}

func (m *TicketDownloaderMock) Close() error { return nil }
