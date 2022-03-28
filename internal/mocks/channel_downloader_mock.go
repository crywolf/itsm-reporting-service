package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// ChannelDownloaderMock is a channel downloader mock
type ChannelDownloaderMock struct {
	mock.Mock
}

func (m *ChannelDownloaderMock) DownloadChannelList(_ context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *ChannelDownloaderMock) Close() error { return nil }
