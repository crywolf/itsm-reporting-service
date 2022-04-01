package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// UserDownloaderMock is a user downloader mock
type UserDownloaderMock struct {
	mock.Mock
}

func (m *UserDownloaderMock) DownloadUsers(_ context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *UserDownloaderMock) Reset(_ context.Context) error { return nil }

func (m *UserDownloaderMock) Close() error { return nil }
