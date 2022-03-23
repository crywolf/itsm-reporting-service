package mocks

import (
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/stretchr/testify/mock"
)

// ChannelDownloaderMock is a channel downloader mock
type ChannelDownloaderMock struct {
	mock.Mock
}

func (m *ChannelDownloaderMock) DownloadChannelList() (channel.List, error) {
	args := m.Called()
	return args.Get(0).(channel.List), args.Error(1)
}

func (m *ChannelDownloaderMock) Close() error { return nil }
