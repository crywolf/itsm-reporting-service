package channeldownloader

import (
	"context"
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// ChannelDownloader downloads list of channels from the external service
type ChannelDownloader interface {
	// DownloadChannelList downloads and stores list of channels from the external service
	DownloadChannelList(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewChannelDownloader(channelRepository repository.ChannelRepository) ChannelDownloader {
	return &channelDownloader{
		client:            http.DefaultClient,
		channelRepository: channelRepository,
	}
}

type channelDownloader struct {
	client            *http.Client
	channelRepository repository.ChannelRepository
}

func (d *channelDownloader) Close() error {
	d.client.CloseIdleConnections()
	return nil
}

func (d *channelDownloader) DownloadChannelList(ctx context.Context) error {
	// TODO download channel list
	//d.client.Do()

	channelList := channel.List{
		channel.Channel{
			ChannelID: "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc",
			Name:      "First channel",
		},
		channel.Channel{
			ChannelID: "8b6353c3-46ca-485d-87c3-66bc36c70d88",
			Name:      "Second channel",
		},
	}

	if err := d.channelRepository.StoreChannelList(ctx, channelList); err != nil {
		return err
	}

	return nil
}
