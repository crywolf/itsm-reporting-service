package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// TicketDownloaderMock is a ticket downloader mock
type TicketDownloaderMock struct {
	mock.Mock
}

func (m *TicketDownloaderMock) DownloadTickets(_ context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *TicketDownloaderMock) Reset(_ context.Context) error { return nil }

func (m *TicketDownloaderMock) Close() error { return nil }
