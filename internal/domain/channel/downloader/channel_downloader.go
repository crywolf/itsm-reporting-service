package chandownloader

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
)

// ChannelDownloader downloads list of channels from the external service
type ChannelDownloader interface {
	// DownloadChannelList downloads list of channels from the external service
	DownloadChannelList() (channel.List, error)

	// Close closes client connections
	Close() error
}

func NewChannelDownloader() ChannelDownloader {
	return &channelDownloader{
		client: http.DefaultClient,
	}
}

type channelDownloader struct {
	client *http.Client
}

func (d channelDownloader) Close() error {
	d.client.CloseIdleConnections()
	return nil
}

func (d channelDownloader) DownloadChannelList() (channel.List, error) {
	//TODO implement me
	panic("implement me")
}
